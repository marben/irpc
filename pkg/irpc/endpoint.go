package irpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
)

var ErrServiceAlreadyRegistered = errors.New("service already registered")
var ErrEndpointClosed = errors.New("endpoint is closed")
var ErrEndpointClosedByCounterpart = errors.Join(ErrEndpointClosed, errors.New("endpoint closed by counterpart"))
var ErrServiceNotFound = errors.New("service not found")
var errProtocolError = errors.New("rpc protocol error")
var ErrContextWaitTimedOut = errors.New("context wait timed out")
var errCounterpartClosing = errors.New("counterpart is closing now")

const (
	rpcRequestPacketType packetType = iota + 1
	rpcResponsePacketType
	closingNowPacketType // inform counterpart that i will immediately close the connection
	ctxEndPacketType     // informs service runner that the provided function context expired
)

// todo: should be configurable for each endpoint
const (
	ParallelWorkers     = 3
	ParallelClientCalls = ParallelWorkers + 1
)

const (
	// clientRegistrationServiceId is used to register other services (and give them their ids)
	// 0 is not used, so that uninitialized clients (with service id = 0), errors out
	clientRegistrationServiceId RegisteredServiceId = iota + 1
)

type Service interface {
	Hash() []byte // unique hash of the service
	GetFuncCall(funcId FuncId) (ArgDeserializer, error)
}

type ArgDeserializer func(d *Decoder) (FuncExecutor, error)
type FuncExecutor func(ctx context.Context) Serializable

// our generated code function calls' arguments implements Serializable/Deserializble
type Serializable interface {
	Serialize(e *Encoder) error
}
type Deserializable interface {
	Deserialize(d *Decoder) error
}

// Endpoint represents one side of a connection
// there needs to be a serving endpoint on both sides of connection for communication to work
type Endpoint struct {
	// maps registered service's hash to it's given id
	serviceHashes map[string]RegisteredServiceId
	services      map[RegisteredServiceId]Service
	nextServiceId RegisteredServiceId

	connCloser io.Closer // close the connection
	enc        *Encoder  // does encoding on the bufWriter

	reqNumsC chan ReqNumT
	writeMux Semaphore

	pendingRequests map[ReqNumT]pendingRequest

	// active running workers
	// lock m before accessing
	serviceWorkers map[ReqNumT]serviceWorker
	wrkrQueue      chan struct{}

	// closed on read loop end
	readLoopRunningC chan struct{}

	m sync.Mutex

	// context ends with the closing of endpoint
	Ctx       context.Context // todo: not sure this should be exported
	ctxCancel context.CancelCauseFunc
}

func NewEndpoint(conn io.ReadWriteCloser, services ...Service) *Endpoint {
	globalContext, cancel := context.WithCancelCause(context.Background())

	reqNumsC := make(chan ReqNumT, ParallelClientCalls)
	for i := range ParallelClientCalls {
		reqNumsC <- ReqNumT(i)
	}

	ep := &Endpoint{
		serviceHashes:    make(map[string]RegisteredServiceId),
		services:         make(map[RegisteredServiceId]Service),
		pendingRequests:  make(map[ReqNumT]pendingRequest),
		serviceWorkers:   make(map[ReqNumT]serviceWorker),
		nextServiceId:    clientRegistrationServiceId + 1,
		connCloser:       conn,
		enc:              NewEncoder(conn),
		reqNumsC:         reqNumsC,
		readLoopRunningC: make(chan struct{}, 1),
		wrkrQueue:        make(chan struct{}, ParallelWorkers),
		writeMux:         NewContextSemaphore(1),
		Ctx:              globalContext,
		ctxCancel:        cancel,
	}

	// default service for registering clients
	regService := &clientRegisterService{ep: ep}
	ep.services[clientRegistrationServiceId] = regService

	ep.RegisterServices(services...)

	// our read loop
	go func() {
		defer func() {
			ep.readLoopRunningC <- struct{}{}
		}()

		dec := NewDecoder(conn)
		err := ep.readMsgs(globalContext, dec)
		switch err {
		case context.Canceled:
			// our Close() must have been called
			return
		case errCounterpartClosing:
			// countepart Ep sent us closing packed
			// we need to close our part
			ep.ctxCancel(ErrEndpointClosedByCounterpart)
		default:

			// unknown connection error
			// close the endpoint
			ep.ctxCancel(errors.Join(ErrEndpointClosed, err))
		}
	}()

	return ep
}

// called on internal Endpoint error(connection drop etc)
// we cannot do it using the normal Close call, because  we don't know what could be failing
func (e *Endpoint) closeOnReadError(err error) {
	// log.Println("close on read error!!!")
	// switch
	e.ctxCancel(err)
	e.connCloser.Close()
	// log.Printf("internal error close(): %+v", err)
}

func (e *Endpoint) closeAlreadyCalled() bool {
	return e.Ctx.Err() != nil
}

func (e *Endpoint) signalOurClosing(ctx context.Context) error {
	unlock, err := e.writeMux.Lock(ctx)
	if err != nil {
		return err
	}
	defer unlock()

	if err := e.serializePacketToConnLocked(packetHeader{typ: closingNowPacketType}); err != nil {
		return fmt.Errorf("sending closingNowPacket: %w", err)
	}

	return nil
}

// Close immediately stops serving any requests
// Close closes the underlying connection
// returns nil on successfull closing
// Close returns ErrEndpointClosed if the Endpoint was already closed
func (e *Endpoint) Close() error {
	// if the context was already canceled, there was already an error or this is second call
	if e.closeAlreadyCalled() {
		return context.Cause(e.Ctx)
	}

	// todo: made up number
	timeout := 300 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var closeError error

	// closes readMsg loop
	// notifies all workers to cancel
	// all new requests (CallRemoteFunc()) will err out immediately
	e.ctxCancel(ErrEndpointClosed)

	// we have mutex. no other writes happening now
	// sign to our counterpart, that we are about to close connection
	if err := e.signalOurClosing(ctx); err != nil {
		closeError = errors.Join(closeError, fmt.Errorf("failed to signal our closing: %w", err))
	}
	// close the underlying connection
	// this errors the read loop in case it didn't notice the context cancelation
	if err := e.connCloser.Close(); err != nil {
		closeError = errors.Join(closeError, fmt.Errorf("connection.Close(): %w", err))
	}

	// make sure the read loop has canceled
	select {
	case <-e.readLoopRunningC:
		// ok
	case <-ctx.Done():
		closeError = errors.Join(closeError, fmt.Errorf("close timed out after %s", timeout))
	}

	return closeError
}

// registers client on remote endpoint
// if there is no corresponding service registered on remote node, returns error
// todo: rename to RegisterRemoteClient
func (e *Endpoint) RegisterClient(ctx context.Context, serviceHash []byte) (RegisteredServiceId, error) {
	req := clientRegisterReq{ServiceHash: serviceHash}
	resp := clientRegisterResp{}
	if err := e.CallRemoteFunc(ctx, clientRegistrationServiceId, 0 /* currently there is just one function */, req, &resp); err != nil {
		return 0, fmt.Errorf("failed to register client %q: %w", serviceHash, err)
	}

	return resp.ServiceId, nil
}

// getService returns ErrServiceNotFound if id was not found,
// nil otherwise
func (e *Endpoint) getService(id RegisteredServiceId) (Service, error) {
	e.m.Lock()
	defer e.m.Unlock()

	s, found := e.services[id]
	if !found {
		return nil, ErrServiceNotFound
	}
	return s, nil
}

func (e *Endpoint) newRequestNumber(ctx context.Context) (ReqNumT, error) {
	select {
	case n := <-e.reqNumsC:
		return n, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func (e *Endpoint) sendRpcRequest(ctx context.Context, serviceId RegisteredServiceId, funcId FuncId, reqData Serializable, respData Deserializable) (pendingRequest, error) {
	unlock, err := e.writeMux.Lock(ctx)
	if err != nil {
		return pendingRequest{}, err
	}
	defer unlock()

	pr, err := e.addPendingRequest(ctx, respData)
	if err != nil {
		return pendingRequest{}, fmt.Errorf("addPendingRequest(): %w", err)
	}

	header := packetHeader{
		typ: rpcRequestPacketType,
	}

	requestDef := requestPacket{
		ReqNum:    pr.reqNum,
		ServiceId: serviceId,
		FuncId:    funcId,
	}

	if err := e.serializePacketToConnLocked(header, requestDef, reqData); err != nil {
		_, popErr := e.popPendingRequest(pr.reqNum)
		if popErr != nil {
			panic(fmt.Errorf("failed to pop a just added pending request: %w", popErr))
		}
		return pendingRequest{}, err
	}

	return pr, nil
}

func (e *Endpoint) sendRequestContextCancelation(req ReqNumT, cause error) error {
	header := packetHeader{typ: ctxEndPacketType}
	ctxEndDef := ctxEndPacket{
		ReqNum: req,
		ErrStr: cause.Error(),
	}

	unlock, err := e.writeMux.Lock(e.Ctx)
	if err != nil {
		return err
	}
	defer unlock()

	if err := e.serializePacketToConnLocked(header, ctxEndDef); err != nil {
		return err
	}

	return nil
}

func (e *Endpoint) CallRemoteFunc(ctx context.Context, serviceId RegisteredServiceId, funcId FuncId, reqData Serializable, respData Deserializable) error {
	// check if our endpoint was closed
	if e.closeAlreadyCalled() {
		return context.Cause(e.Ctx)
	}

	pendingReq, err := e.sendRpcRequest(ctx, serviceId, funcId, reqData, respData)
	if err != nil {
		return err
	}

	select {
	// response has arrived
	case err := <-pendingReq.deserErrC:
		return err

	// global endpoint's context ended
	case <-e.Ctx.Done():
		return context.Cause(e.Ctx)

	// the client provided context expired
	case <-ctx.Done():
		// we notify the remote endpoint to speed up the function completion.
		// otherwise we still just wait for response (or endpoint exit)
		if err := e.sendRequestContextCancelation(pendingReq.reqNum, context.Cause(ctx)); err != nil {
			return fmt.Errorf("failed to cancel request %d: %w", pendingReq.reqNum, err)
		}

		select {
		case err := <-pendingReq.deserErrC:
			return err
		case <-e.Ctx.Done():
			return context.Cause(e.Ctx)
		}
	}
}

// todo: variadic is not such a great idea - what if only one error fails to register? what to return?
func (e *Endpoint) RegisterServices(services ...Service) error {
	e.m.Lock()
	defer e.m.Unlock()

	for _, service := range services {
		if _, found := e.serviceHashes[string(service.Hash())]; found {
			return fmt.Errorf("%s: %w", service.Hash(), ErrServiceAlreadyRegistered)
		}

		id := e.nextServiceId
		e.nextServiceId++

		e.serviceHashes[string(service.Hash())] = id
		e.services[id] = service
	}

	return nil
}

// expects the mutex to be locked
// returns ErrEndpointClosed if endpoint is in closing state
func (e *Endpoint) serializePacketToConnLocked(data ...Serializable) error {
	// todo: do the locking here, not in the above function? (that would probably need splitting the mutext)

	for _, s := range data {
		if err := s.Serialize(e.enc); err != nil {
			return fmt.Errorf("Serialize: %w", err)
		}
	}

	if err := e.enc.flush(); err != nil {
		return fmt.Errorf("encoder.Flush(): %w", err)
	}

	return nil
}

// respData is the actual serialized return data from the function
func (e *Endpoint) sendResponse(reqNum ReqNumT, respData Serializable) error {
	resp := responsePacket{
		ReqNum: reqNum,
	}

	header := packetHeader{
		typ: rpcResponsePacketType,
	}

	unlock, err := e.writeMux.Lock(e.Ctx)
	if err != nil {
		return err
	}
	defer unlock()

	if err := e.serializePacketToConnLocked(header, resp, respData); err != nil {
		return fmt.Errorf("failed to serialize response to connection: %w", err)
	}

	return nil
}

// readMsgs is the main incoming messages processing loop
// returns wrapped errProtocolError(todo: actually it doesn't) if the error is on a protocol level	// todo: make sure it does
// returns context.Canceled if context has been canceled
func (e *Endpoint) readMsgs(ctx context.Context, dec *Decoder) error {
	for {
		if ctx.Err() != nil {
			log.Printf("%p: readMsg: returning because of canceled context", e)
			return context.Canceled
		}

		var h packetHeader
		if err := h.Deserialize(dec); err != nil {
			return fmt.Errorf("read header: %w", err)
		}

		switch h.typ {
		// counterpart requested us to run a function
		case rpcRequestPacketType:
			var req requestPacket
			if err := req.Deserialize(dec); err != nil {
				return fmt.Errorf("read request packet:%w", err)
			}

			service, err := e.getService(req.ServiceId)
			if err != nil {
				return errors.Join(errProtocolError, fmt.Errorf("getService() %v: %w", req, err))
			}
			argDeser, err := service.GetFuncCall(req.FuncId)
			if err != nil {
				return fmt.Errorf("GetFuncCall() %v: %w", req, err)
			}
			funcExec, err := argDeser(dec)
			if err != nil {
				return fmt.Errorf("argDeserialize: %w", err)
			}

			// waits until worker slot is available (blocks here on too many long rpcs)
			err = e.startServiceWorker(ctx, req.ReqNum, funcExec)
			if err != nil {
				return fmt.Errorf("new worker: %w", err)
			}

		// response from a function we requested from counterpart
		case rpcResponsePacketType:
			var resp responsePacket
			if err := resp.Deserialize(dec); err != nil {
				return fmt.Errorf("failed to read response data:%w", err)
			}
			pr, err := e.popPendingRequest(resp.ReqNum)
			if err != nil {
				return fmt.Errorf("request not found: %w", err)
			}

			pr.deserErrC <- pr.resp.Deserialize(dec)

		// counterpart is closing
		case closingNowPacketType:
			return errCounterpartClosing

		// counterpart informs us that context of some function it requested us to execute, has expired
		// we cancel corresponting worker's context and hope response gets sent fast
		case ctxEndPacketType:
			var cancelRequest ctxEndPacket
			if err := cancelRequest.Deserialize(dec); err != nil {
				return fmt.Errorf("failed to deserialize context cancelation request: %w", err)
			}
			clientErr := errors.New(cancelRequest.ErrStr)
			e.cancelRequest(cancelRequest.ReqNum, clientErr)

		default:
			return fmt.Errorf("unexpected packet type: %+v", h.typ)
		}
	}
}

func (e *Endpoint) cancelRequest(rnum ReqNumT, cancelErr error) {
	e.m.Lock()
	defer e.m.Unlock()

	rw, found := e.serviceWorkers[rnum]
	if !found {
		// the work may have finished, while this request was on the way.
		// we don't care
		return
	}

	rw.cancel(cancelErr)
}

// blocks until worker slot is available
// todo: take context as argument
func (e *Endpoint) startServiceWorker(ctx context.Context, reqNum ReqNumT, exe FuncExecutor) error {
	// waits until worker slot is available (blocks here on too many long rpcs)
	select {
	case e.wrkrQueue <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	ctxCancelable, cancel := context.WithCancelCause(ctx)

	rw := serviceWorker{
		cancel: cancel,
	}

	// a goroutine is created for each remote call

	e.m.Lock()
	e.serviceWorkers[reqNum] = rw
	e.m.Unlock()

	go func() {
		// release the worker queue
		defer func() { <-e.wrkrQueue }()

		defer func() {
			e.m.Lock()
			defer e.m.Unlock()
			delete(e.serviceWorkers, reqNum)
		}()

		resp := exe(ctxCancelable)

		if err := e.sendResponse(reqNum, resp); err != nil {
			// if errors.Is(err, ErrEndpointClosed) {
			// 	return
			// }
			e.closeOnReadError(fmt.Errorf("sendResponse() for request %d: %v", reqNum, err))
		}
	}()

	return nil
}

func (e *Endpoint) addPendingRequest(ctx context.Context, resp Deserializable) (pendingRequest, error) {
	reqNum, err := e.newRequestNumber(ctx)
	if err != nil {
		return pendingRequest{}, fmt.Errorf("newRequestNumber: %w", err)
	}

	pr := pendingRequest{
		reqNum:    reqNum,
		resp:      resp,
		deserErrC: make(chan error, 1),
	}

	e.m.Lock()
	defer e.m.Unlock()
	e.pendingRequests[reqNum] = pr

	return pr, nil
}

func (e *Endpoint) popPendingRequest(reqNum ReqNumT) (pendingRequest, error) {
	e.m.Lock()
	defer e.m.Unlock()

	pr, found := e.pendingRequests[reqNum]
	if !found {
		return pendingRequest{}, fmt.Errorf("pending request %d not found", reqNum)
	}
	delete(e.pendingRequests, reqNum)
	e.reqNumsC <- reqNum
	return pr, nil
}

// represents a pending request that is being executed on the opposing endpoint
type pendingRequest struct {
	reqNum    ReqNumT
	resp      Deserializable
	deserErrC chan error
}

// a request from opposing endpoint, that we are executing
type serviceWorker struct {
	cancel context.CancelCauseFunc // todo: we don't need this struct for just a function pointer
}
