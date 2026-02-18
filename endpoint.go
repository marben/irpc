package irpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/marben/irpc/irpcgen"
)

// serviceHashLen is the length of the service hash used to identify services in the Endpoint.
// services have longer ids, but we just us the first n bytes of it
const serviceHashLen = 4

// DefaultParallelWorkers is the default number of parallel workers servicing our peer's requests
// It can be overridden for each endpoint with [WithParallelWorkers] option
var DefaultParallelWorkers = 3

// DefaultParallelClientCalls is the number of parallel calls we allow to our peer at the same time.
// It can be overridden for each endpoint with [WithParallelClientCalls] option
var DefaultParallelClientCalls = DefaultParallelWorkers + 1

// Endpoint related errors
var (
	ErrEndpointClosed       = errors.New("irpc: endpoint is closed")
	ErrEndpointClosedByPeer = errors.New("irpc: endpoint closed by peer")
	ErrServiceNotFound      = errors.New("irpc: service not found") // todo: not used?
	errProtocolError        = errors.New("protocol error")
)

// Endpoint represents one side of an active RPC connection.
//
// Endpoint owns the lifetime of the underlying connection and the goroutines
// servicing it. The lifetime is exposed via Context().
//
// The context returned by Context() is canceled when:
//   - Close is called locally
//   - the remote endpoint (peer) closes the connection
//   - a protocol or I/O error occurs
//
// The cancellation cause can be inspected using [context.Cause].
//
// Endpoint implements [irpcgen.Endpoint].
type Endpoint struct {
	// maps serviceId to service. uses array, because slices are not comparable in maps
	services    map[[serviceHashLen]byte]irpcgen.Service
	servicesMux sync.Mutex

	// localAddr and remoteAddr are nil, when not set with Option
	localAddr  net.Addr // our network address if available
	remoteAddr net.Addr // peer network address if available

	enc    *irpcgen.Encoder // encoder for writing messages to the connection
	encMux sync.Mutex

	dec *irpcgen.Decoder // decoder for reading messages from the connection

	ourPendingRequests *ourPendingRequestsLog

	connCloser io.Closer // closes our connection

	closeOnce sync.Once

	// context is Done() after the endpoint is closed
	ctx       context.Context
	ctxCancel context.CancelCauseFunc

	// some options - possibly move to a separate struct?
	parallelWorkers     int // number of parallel workers servicing peer's requests
	parallelClientCalls int // number of parallel calls we allow to our peer at the same time
}

// NewEndpoint creates and runs a new Endpoint with the given connection and options.
// It immeditely starts a go routine servicing the communication.
// Services provided by this endpoint can be added as with [WithEndpointServices] option, or can be registered later with [Endpoint.RegisterService] call
func NewEndpoint(conn io.ReadWriteCloser, opts ...EndpointOption) *Endpoint {
	epCtx, endpointContextCancel := context.WithCancelCause(context.Background())

	ep := &Endpoint{
		services:            make(map[[serviceHashLen]byte]irpcgen.Service),
		enc:                 irpcgen.NewEncoder(conn),
		dec:                 irpcgen.NewDecoder(conn),
		connCloser:          conn,
		ctx:                 epCtx,
		ctxCancel:           endpointContextCancel,
		parallelWorkers:     DefaultParallelWorkers,
		parallelClientCalls: DefaultParallelClientCalls,
	}

	for _, opt := range opts {
		opt(ep)
	}

	ep.ourPendingRequests = newOurPendingRequestsLog(ep.parallelClientCalls)

	go func() {
		ep.serve(epCtx)
	}()

	return ep
}

// Context returns a context that is canceled when the endpoint is closed.
//
// The cancellation cause describes why the endpoint terminated and can be
// retrieved using [context.Cause].
func (e *Endpoint) Context() context.Context {
	return e.ctx
}

func (e *Endpoint) serve(ctx context.Context) {
	exec := newExecutor(ctx, e.parallelWorkers)
	readC := make(chan error, 1)
	go func() {
		readC <- e.readLoop(exec)
	}()

	var err error
	select {
	case <-ctx.Done():
	// termination already initiated (Close or other path)
	case err = <-readC:
		e.handleIOError(err)
	case err = <-exec.errC:
		e.terminate(errors.Join(ErrEndpointClosed, err))
	}

	// following 2 calls may fail for various reasons, but we want to be sure, they were made
	e.serializePacket(packetHeader{typ: closingNowPacketType})
	e.connCloser.Close()
}

func (e *Endpoint) terminate(cause error) {
	e.closeOnce.Do(func() {
		// Always close transport first to unblock goroutines
		_ = e.connCloser.Close()

		e.ctxCancel(cause)
	})
}

func (e *Endpoint) handleIOError(err error) {
	e.closeOnce.Do(func() {
		var cause error

		switch {
		case errors.Is(err, ErrEndpointClosedByPeer),
			errors.Is(err, io.EOF), errors.Is(err, net.ErrClosed):
			// Clean remote shutdown
			cause = ErrEndpointClosedByPeer

		default:
			// Transport or protocol error
			cause = errors.Join(ErrEndpointClosed, err)
		}

		_ = e.connCloser.Close()
		e.ctxCancel(cause)
	})
}

// Close initiates shutdown of the endpoint.
//
// Close is idempotent. Calling Close cancels the endpoint's context with
// ErrEndpointClosed as the cause.
func (e *Endpoint) Close() error {
	if cause := context.Cause(e.ctx); cause != nil {
		return cause
	}
	e.terminate(ErrEndpointClosed)
	return nil
}

// RegisterClient registers client on remote endpoint - currently a no-op
func (e *Endpoint) RegisterClient(serviceId []byte) error {
	// currently a no-op
	// we could use this call to negotiate shortcut for serviceId to further reduce service data
	// we could also use this to make sure service is registered on remote ep. this would increase initial latency, so perhaps make it configurable?
	return nil
}

// getService returns false if id was not found,
func (e *Endpoint) getService(serviceHash []byte) (s irpcgen.Service, found bool) {
	if len(serviceHash) != serviceHashLen {
		return nil, false
	}

	e.servicesMux.Lock()
	defer e.servicesMux.Unlock()

	hashArray := [serviceHashLen]byte{}
	copy(hashArray[:], serviceHash)
	s, found = e.services[hashArray]
	return s, found
}

func (e *Endpoint) sendRpcRequest(ctx context.Context, serviceId []byte, funcId irpcgen.FuncId, reqData irpcgen.Serializable, respData irpcgen.Deserializable) (ourPendingRequest, error) {
	pr, err := e.ourPendingRequests.addPendingRequest(ctx, respData)
	if err != nil {
		return ourPendingRequest{}, fmt.Errorf("addPendingRequest(): %w", err)
	}

	header := packetHeader{
		typ: rpcRequestPacketType,
	}

	requestDef := requestPacket{
		ReqNum:    pr.reqNum,
		ServiceId: serviceId,
		FuncId:    funcId,
	}

	if err := e.serializePacket(header, requestDef, reqData); err != nil {
		_, popErr := e.ourPendingRequests.popPendingRequest(pr.reqNum)
		if popErr != nil {
			panic(fmt.Errorf("failed to pop a just added pending request: %w", popErr))
		}
		return ourPendingRequest{}, err
	}

	return pr, nil
}

func (e *Endpoint) sendRequestContextCancelation(req reqNumT, cause error) error {
	header := packetHeader{typ: ctxEndPacketType}
	ctxEndDef := ctxEndPacket{
		ReqNum: req,
		ErrStr: cause.Error(),
	}

	if err := e.serializePacket(header, ctxEndDef); err != nil {
		return err
	}

	return nil
}

func (e *Endpoint) sendResponse(reqNum reqNumT, respData irpcgen.Serializable) error {
	resp := responsePacket{
		ReqNum: reqNum,
	}

	header := packetHeader{
		typ: rpcResponsePacketType,
	}

	if err := e.serializePacket(header, resp, respData); err != nil {
		return fmt.Errorf("failed to serialize response to connection: %w", err)
	}

	return nil
}

func (e *Endpoint) serializePacket(data ...irpcgen.Serializable) error {
	e.encMux.Lock()
	defer e.encMux.Unlock()

	for _, d := range data {
		if err := d.Serialize(e.enc); err != nil {
			return fmt.Errorf("data.Serialize(): %w", err)
		}
	}

	if err := e.enc.Flush(); err != nil {
		return fmt.Errorf("encoder.Flush(): %w", err)
	}

	return nil
}

// CallRemoteFunc invokes a function on the peer Endpoint.
//
// CallRemoteFunc implements [irpcgen.Service]
func (e *Endpoint) CallRemoteFunc(ctx context.Context, serviceId []byte, funcId irpcgen.FuncId, reqData irpcgen.Serializable, respData irpcgen.Deserializable) error {
	pendingReq, err := e.sendRpcRequest(ctx, serviceId[:serviceHashLen], funcId, reqData, respData)
	if err != nil {
		// check if endpoint was closed
		if cause := context.Cause(e.ctx); cause != nil {
			return cause
		}
		return err
	}

	select {
	// response has arrived
	case err := <-pendingReq.deserErrC:
		return err

	// global endpoint's context ended
	case <-e.ctx.Done():
		return context.Cause(e.ctx)

	// the client provided context expired
	case <-ctx.Done():
		// we notify the remote endpoint to speed up the function completion.
		// otherwise we still just wait for response (or endpoint exit)
		if err := e.sendRequestContextCancelation(pendingReq.reqNum, context.Cause(ctx)); err != nil {
			e.handleIOError(err)
			return context.Cause(e.ctx)
		}

		select {
		case err := <-pendingReq.deserErrC:
			return err
		case <-e.ctx.Done():
			return context.Cause(e.ctx)
		}
	}
}

func (e *Endpoint) RegisterService(services ...irpcgen.Service) {
	e.servicesMux.Lock()
	defer e.servicesMux.Unlock()

	for _, s := range services {
		hashArray := [serviceHashLen]byte{}
		copy(hashArray[:], s.Id())
		e.services[hashArray] = s
	}
}

func (e *Endpoint) LocalAddr() net.Addr {
	return e.localAddr
}

func (e *Endpoint) RemoteAddr() net.Addr {
	return e.remoteAddr
}

// processRequest decodes the request and runs the function in separate goroutine
// blocks until worker slot is available
func (e *Endpoint) processRequest(dec *irpcgen.Decoder, exec *executor) error {
	var req requestPacket
	if err := req.Deserialize(dec); err != nil {
		return fmt.Errorf("read request packet:%w", err)
	}

	service, found := e.getService(req.ServiceId)
	if !found {
		return errors.Join(errProtocolError, fmt.Errorf("getService(): %w", ErrServiceNotFound))
	}

	argDeser, err := service.GetFuncCall(irpcgen.FuncId(req.FuncId))
	if err != nil {
		return errors.Join(errProtocolError, fmt.Errorf("GetFuncCall() %v: %w", req, err))
	}
	funcExec, err := argDeser(dec)
	if err != nil {
		return fmt.Errorf("argDeserialize: %w", err)
	}

	err = exec.runServiceWorker(req.ReqNum, funcExec, e.sendResponse)
	if err != nil {
		return fmt.Errorf("new worker: %w", err)
	}

	return nil
}

// readLoop is the main incoming messages processing loop
func (e *Endpoint) readLoop(exec *executor) error {
	for {
		var h packetHeader
		if err := h.Deserialize(e.dec); err != nil {
			return fmt.Errorf("read header: %w", err)
		}

		switch h.typ {
		// peer requested us to run a function
		case rpcRequestPacketType:
			if err := e.processRequest(e.dec, exec); err != nil {
				return fmt.Errorf("processRequest: %w", err)
			}
			// response from a function we requested from peer
		case rpcResponsePacketType:
			var resp responsePacket
			if err := resp.Deserialize(e.dec); err != nil {
				return fmt.Errorf("failed to read response data:%w", err)
			}
			pr, err := e.ourPendingRequests.popPendingRequest(resp.ReqNum)
			if err != nil {
				return fmt.Errorf("request not found: %w", err)
			}

			pr.deserErrC <- pr.resp.Deserialize(e.dec)

		// peer is closing
		case closingNowPacketType:
			e.terminate(ErrEndpointClosedByPeer)

		// peer informed us that context of some function it requested us to execute has expired
		// we cancel corresponding worker's context.
		// this doesn't mean the work will stop immediately
		// goroutine needs to notice the context cancellation and stop itself
		case ctxEndPacketType:
			var cancelRequest ctxEndPacket
			if err := cancelRequest.Deserialize(e.dec); err != nil {
				return fmt.Errorf("failed to deserialize context cancelation request: %w", err)
			}
			clientErr := errors.New(cancelRequest.ErrStr)
			exec.cancelRequest(cancelRequest.ReqNum, clientErr)

		default:
			return fmt.Errorf("unexpected packet type: %+v", h.typ)
		}
	}
}

type EndpointOption func(*Endpoint)

func WithEndpointServices(s ...irpcgen.Service) EndpointOption {
	return func(ep *Endpoint) {
		ep.RegisterService(s...)
	}
}

func WithLocalAddress(addr net.Addr) EndpointOption {
	return func(ep *Endpoint) {
		ep.localAddr = addr
	}
}

func WithRemoteAddress(addr net.Addr) EndpointOption {
	return func(ep *Endpoint) {
		ep.remoteAddr = addr
	}
}

func WithParallelWorkers(parallelWorkers int) EndpointOption {
	return func(ep *Endpoint) {
		ep.parallelWorkers = parallelWorkers
	}
}
func WithParallelClientCalls(parallelClientCalls int) EndpointOption {
	return func(ep *Endpoint) {
		ep.parallelClientCalls = parallelClientCalls
	}
}
