package irpc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/semaphore"
)

var ErrServiceAlreadyRegistered = errors.New("service already registered")
var ErrEndpointClosed = errors.New("endpoint is closed")
var ErrServiceNotFound = errors.New("service not found")
var errProtocolError = errors.New("rpc protocol error")
var ErrContextWaitTimedOut = errors.New("context wait timed out")

// todo: make one irpc error with err type iota inside

type packetType uint8

const (
	rpcRequestPacketType packetType = iota + 1
	rpcResponsePacketType
	closingNowPacketType // inform counterpart that i will immediately close the connection
	ctxEndPacketType     // informs service runner that the provided function context expired
)

// todo: should be configurable for each endpoint
const (
	ParallelWorkers = 3
)

type packetHeader struct {
	typ packetType
}

type requestPacket struct {
	ReqNum    ReqNumT
	ServiceId RegisteredServiceId
	FuncId    FuncId
}

type responsePacket struct {
	ReqNum ReqNumT // request number that initiated this response
}

type ctxEndPacket struct {
	ReqNum ReqNumT
	ErrStr string
}

// client registers to a service (by hash). upon registration, service is given 'RegisteredServiceId'
type RegisteredServiceId uint16
type FuncId uint16
type ReqNumT uint16

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

	connCloser io.Closer     // close the connection
	bufWriter  *bufio.Writer // buffered Writer of the connection	// only used for flushing. rest of code should use encoder
	enc        *Encoder      // does encoding on the bufWriter

	reqNum ReqNumT // serial number of our next request	// todo: sort out overflow (and test?)

	pendingRequests map[ReqNumT]pendingRequest

	// limits the number of active workers
	serviceWorkersSem *semaphore.Weighted

	// active running workers
	// lock m before accessing
	serviceWorkers map[ReqNumT]serviceWorker

	// error returned by message read go routine
	readLoopErrC chan error

	m sync.Mutex

	closing atomic.Bool // todo: maybe replace with Endpoint.Ctx

	// context ends with the closing of endpoint
	Ctx       context.Context // todo: this is a test. not sure about the design.
	ctxCancel context.CancelCauseFunc
}

func NewEndpoint(conn io.ReadWriteCloser, services ...Service) *Endpoint {
	bufWriter := bufio.NewWriter(conn)
	ctx, cancel := context.WithCancelCause(context.Background())
	ep := &Endpoint{
		serviceHashes:     make(map[string]RegisteredServiceId),
		services:          make(map[RegisteredServiceId]Service),
		pendingRequests:   make(map[ReqNumT]pendingRequest),
		serviceWorkers:    make(map[ReqNumT]serviceWorker),
		nextServiceId:     clientRegistrationServiceId + 1,
		connCloser:        conn,
		bufWriter:         bufWriter,
		enc:               NewEncoder(bufWriter),
		readLoopErrC:      make(chan error, 1),
		Ctx:               ctx,
		ctxCancel:         cancel,
		serviceWorkersSem: semaphore.NewWeighted(ParallelWorkers),
	}

	// default service for registering clients
	regService := &clientRegisterService{ep: ep}
	ep.services[clientRegistrationServiceId] = regService

	ep.RegisterServices(services...)

	// our read loop
	go func() {
		dec := NewDecoder(bufio.NewReader(conn))
		switch err := ep.readMsgs(dec); err {
		case ErrEndpointClosed:
			// panic(err)
			// todo: we need to call close
			ep.readLoopErrC <- ErrEndpointClosed // todo: wrong?
		default:
			// log.Fatalf("err: %+v", err)
			// panic(err)
			// todo: close etc...
			ep.readLoopErrC <- ErrEndpointClosed // wrong!
			// panic(fmt.Sprintf("readMsg: %s", err.Error()))	// todo: uncomment and figure out proper closing pattern
		}
	}()

	return ep
}

// this is currently just a placeholder for internal err state which should eventually
// do a clean Close of the endpoint
func (e *Endpoint) errClose(err error) {
	log.Printf("internal error close(): %+v", err)
	panic(err)
}

// Close immediately stops serving any requests
// Close closes the underlying connection
// Close returns ErrEndpointClosed if the Endpoint was already closed
func (e *Endpoint) Close() error {
	if e.closing.Load() {
		return ErrEndpointClosed
	}

	e.m.Lock()
	defer e.m.Unlock()

	// we have mutex. no other writes happening now
	// sign to our counterpart, that we are about to close connection
	if err := e.serializePacketToConnLocked(packetHeader{typ: closingNowPacketType}); err != nil {
		return fmt.Errorf("notify counterpart: %w", err)
	}

	// no more conn writing after this
	e.closing.Store(true)
	e.ctxCancel(ErrEndpointClosed) // this errs out waiting rpc calls

	// we are the ones closing the underlying connection
	// this makes read loop err out
	if err := e.connCloser.Close(); err != nil {
		return fmt.Errorf("connection.Close(): %w", err)
	}
	e.connCloser = nil

	// todo: we need to err out all pending requests

	// conn close triggers reader error
	// we don't care about the error
	<-e.readLoopErrC

	return nil
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

func (e *Endpoint) genNewRequestNumberLocked() ReqNumT {
	reqNum := e.reqNum
	e.reqNum += 1
	return reqNum
}

func (e *Endpoint) sendRpcRequest(serviceId RegisteredServiceId, funcId FuncId, reqData Serializable, respData Deserializable) (pendingRequest, error) {
	e.m.Lock()
	defer e.m.Unlock()

	reqNum := e.genNewRequestNumberLocked()

	header := packetHeader{
		typ: rpcRequestPacketType,
	}

	requestDef := requestPacket{
		ReqNum:    reqNum,
		ServiceId: serviceId,
		FuncId:    funcId,
	}

	if err := e.serializePacketToConnLocked(header, requestDef, reqData); err != nil {
		return pendingRequest{}, err
	}

	pr := pendingRequest{
		reqNum:    reqNum,
		resp:      respData,
		deserErrC: make(chan error, 1),
	}

	e.pendingRequests[reqNum] = pr

	return pr, nil
}

func (e *Endpoint) sendContextCancelation(req ReqNumT, cause error) error {
	log.Printf("sending cancel context request with err: %+v", cause)
	header := packetHeader{typ: ctxEndPacketType}
	ctxEndDef := ctxEndPacket{
		ReqNum: req,
		ErrStr: cause.Error(),
	}

	e.m.Lock()
	defer e.m.Unlock()
	if err := e.serializePacketToConnLocked(header, ctxEndDef); err != nil {
		return err
	}

	return nil
}

func (e *Endpoint) CallRemoteFunc(ctx context.Context, serviceId RegisteredServiceId, funcId FuncId, reqData Serializable, respData Deserializable) error {
	pendingReq, err := e.sendRpcRequest(serviceId, funcId, reqData, respData)
	if err != nil {
		return err
	}

	select {
	// response has arrived
	case err := <-pendingReq.deserErrC:
		return err

	// global endpoint's context ended
	case <-e.Ctx.Done():

		return ErrEndpointClosed

	// the client provided context expired
	case <-ctx.Done():
		if err := e.sendContextCancelation(pendingReq.reqNum, context.Cause(ctx)); err != nil {
			return fmt.Errorf("failed to cancel request %d: %w", e.reqNum, err)
		}

		select {
		case err := <-pendingReq.deserErrC:
			return err
		case <-e.Ctx.Done():
			return ErrEndpointClosed
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
	if e.closing.Load() {
		return ErrEndpointClosed
	}

	for _, s := range data {
		if err := s.Serialize(e.enc); err != nil {
			return fmt.Errorf("Serialize: %w", err)
		}
	}

	if err := e.bufWriter.Flush(); err != nil {
		return fmt.Errorf("bufWriter.Flush(): %w", err)
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

	e.m.Lock()
	defer e.m.Unlock()

	if err := e.serializePacketToConnLocked(header, resp, respData); err != nil {
		return fmt.Errorf("failed to serialize response to connection: %w", err)
	}

	return nil
}

// readMsgs is the main incoming messages processing loop
// returns wrapped errProtocolError(todo: actually it doesn't) if the error is on a protocol level	// todo: make sure it does
// returns ErrEndpointClosed if we received close request packet
func (e *Endpoint) readMsgs(dec *Decoder) error {
	for {
		var h packetHeader
		if err := h.Deserialize(dec); err != nil {
			return fmt.Errorf("read header: %w", err)
		}

		switch h.typ {
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
			err = e.startServiceWorker(req.ReqNum, funcExec)
			if err != nil {
				return fmt.Errorf("new worker: %w", err)
			}

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

		case closingNowPacketType:
			// todo: maybe just return err and let handler close the endpoint?
			e.closing.Store(true)          // todo: is this used?
			e.ctxCancel(ErrEndpointClosed) // close waiting functions
			return ErrEndpointClosed

		case ctxEndPacketType:
			var cancelRequest ctxEndPacket
			if err := cancelRequest.Deserialize(dec); err != nil {
				return fmt.Errorf("failed to deserialize context cancelation request: %w", err)
			}
			log.Printf("service: context for req %d has been cancelled with %q", cancelRequest.ReqNum, cancelRequest.ErrStr)
			clientErr := errors.New(cancelRequest.ErrStr)
			e.cancelRequest(cancelRequest.ReqNum, clientErr)

		default:
			panic(fmt.Sprintf("unexpected packet type: %+v", h.typ)) // todo: remove panic, just return error
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
func (e *Endpoint) startServiceWorker(reqNum ReqNumT, exe FuncExecutor) error {
	// waits until worker slot is available (blocks here on too many long rpcs)
	// todo: probably switch to channel based worker pool, which allows to check for context cancellation
	if err := e.serviceWorkersSem.Acquire(e.Ctx, 1); err != nil {
		return fmt.Errorf("worker semaphore.Acquire(): %w", err)
	}

	ctxCancelable, cancel := context.WithCancelCause(e.Ctx)

	rw := serviceWorker{
		cancel: cancel,
	}

	// a goroutine is created for each remote call

	e.m.Lock()
	e.serviceWorkers[reqNum] = rw
	e.m.Unlock()

	go func() {
		defer func() {
			e.m.Lock()
			defer e.m.Unlock()
			delete(e.serviceWorkers, reqNum)
		}()

		defer e.serviceWorkersSem.Release(1)

		resp := exe(ctxCancelable)

		// log.Printf("exec finished. sending response for request %d", reqNum)
		if err := e.sendResponse(reqNum, resp); err != nil {
			if errors.Is(err, ErrEndpointClosed) {
				return
			}
			e.errClose(fmt.Errorf("failed to send response for request %d: %v", reqNum, err))
		}
	}()

	return nil
}

func (e *Endpoint) popPendingRequest(reqNum ReqNumT) (pendingRequest, error) {
	e.m.Lock()
	defer e.m.Unlock()

	pr, found := e.pendingRequests[reqNum]
	if !found {
		return pendingRequest{}, fmt.Errorf("pending request %d not found", reqNum)
	}
	delete(e.pendingRequests, reqNum)
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
