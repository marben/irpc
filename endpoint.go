package irpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/marben/irpc/irpcgen"
)

// serviceHashLen is the length of the service hash used to identify services in the Endpoint.
// services have longer ids, but we just us the first n bytes of it
const serviceHashLen = 4

// DefaultParallelWorkers is the default number of parallel workers servicing our counterpart's requests
// It can be overridden for each endpoint with [WithParallelWorkers] option
var DefaultParallelWorkers = 3

// DefaultParallelClientCalls is the number of parallel calls we allow to our counterpart at the same time.
// It can be overridden for each endpoint with [WithParallelClientCalls] option
var DefaultParallelClientCalls = DefaultParallelWorkers + 1

// Endpoint related errors
var (
	ErrEndpointClosed              = errors.New("endpoint is closed")
	ErrEndpointClosedByCounterpart = errors.Join(ErrEndpointClosed, errors.New("endpoint closed by counterpart"))
	ErrServiceNotFound             = errors.New("service not found")
	errProtocolError               = errors.New("protocol error")
)

// Endpoint represents one side of a socket connection.
// There needs to be a serving endpoint on both sides of connection for communication to work.
type Endpoint struct {
	// maps serviceId to service. uses array, because slices are not comparable in maps
	services    map[[serviceHashLen]byte]irpcgen.Service
	servicesMux sync.Mutex

	enc    *irpcgen.Encoder // encoder for writing messages to the connection
	encMux sync.Mutex

	dec *irpcgen.Decoder // decoder for reading messages from the connection

	ourPendingRequests *ourPendingRequestsLog

	connCloser io.Closer // closes our connection

	// context is Done() after the endpoint is closed
	Ctx       context.Context
	ctxCancel context.CancelCauseFunc

	// some options - possibly move to a separate struct?
	parallelWorkers     int // number of parallel workers servicing counterpart's requests
	parallelClientCalls int // number of parallel calls we allow to our counterpart at the same time
}

// NewEndpoint creates and runs a new Endpoint with the given connection and options.
// It immeditely starts a go routine servicing the communication.
// Services provided by this enpoint can be added as with [WithService] oprtion, or can be registered later with [Endpoint.RegisterService] call
func NewEndpoint(conn io.ReadWriteCloser, opts ...Option) *Endpoint {
	epCtx, endpointContextCancel := context.WithCancelCause(context.Background())

	ep := &Endpoint{
		services:            make(map[[serviceHashLen]byte]irpcgen.Service),
		enc:                 irpcgen.NewEncoder(conn),
		dec:                 irpcgen.NewDecoder(conn),
		connCloser:          conn,
		Ctx:                 epCtx,
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

// Done returns a channel that's closed when work of this endpoint is done
func (e *Endpoint) Done() <-chan struct{} {
	return e.Ctx.Done()
}

// Err returns the underlying cause for ending of this endpoint
func (e *Endpoint) Err() error {
	return context.Cause(e.Ctx)
}

func (e *Endpoint) serve(ctx context.Context) {
	exec := newExecutor(ctx, e.parallelWorkers)
	readC := make(chan error, 1)
	go func() {
		readC <- e.readMsgs(exec)
	}()

	var err error
	select {
	case <-ctx.Done():
	// Close() was called
	case err = <-readC:
		e.ctxCancel(errors.Join(ErrEndpointClosed, err))
	case err = <-exec.errC:
		e.ctxCancel(errors.Join(ErrEndpointClosed, err))
	}

	// following 2 calls may fail for various reasons, but we want to be sure, they were made
	e.serializePacket(packetHeader{typ: closingNowPacketType})
	e.connCloser.Close()
}

// Close immediately stops serving any requests,
// signals closing of our endpoint, closes the underlying connection
// returns nil on successfull closing
// Close returns ErrEndpointClosed if the Endpoint was already closed
func (e *Endpoint) Close() error {
	// if the context was already canceled, there was already an error
	if e.Err() != nil {
		return e.Err()
	}

	e.ctxCancel(ErrEndpointClosed)
	return nil
}

// RegisterClient registers client on remote endpoint - currently a no-op
func (e *Endpoint) RegisterClient(serviceId []byte) error {
	// currently a no-op
	// we could use this call to negotiate shortcut for serviceId to reduce further service data
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

func (e *Endpoint) sendRequestContextCancelation(req ReqNumT, cause error) error {
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

func (e *Endpoint) sendResponse(reqNum ReqNumT, respData irpcgen.Serializable) error {
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

func (e *Endpoint) CallRemoteFunc(ctx context.Context, serviceId []byte, funcId irpcgen.FuncId, reqData irpcgen.Serializable, respData irpcgen.Deserializable) error {
	pendingReq, err := e.sendRpcRequest(ctx, serviceId[:serviceHashLen], funcId, reqData, respData)
	if err != nil {
		// check if endpoint was closed
		if e.Err() != nil {
			return e.Err()
		}
		return err
	}

	select {
	// response has arrived
	case err := <-pendingReq.deserErrC:
		return err

	// global endpoint's context ended
	case <-e.Done():
		return e.Err()

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

func (e *Endpoint) RegisterService(services ...irpcgen.Service) {
	e.servicesMux.Lock()
	defer e.servicesMux.Unlock()

	for _, s := range services {
		hashArray := [serviceHashLen]byte{}
		copy(hashArray[:], s.Id())
		e.services[hashArray] = s
	}
}

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

	// waits until worker slot is available (blocks here on too many long rpcs)
	err = exec.startServiceWorker(req.ReqNum, funcExec, e.sendResponse)
	if err != nil {
		return fmt.Errorf("new worker: %w", err)
	}

	return nil
}

// readMsgs is the main incoming messages processing loop
func (e *Endpoint) readMsgs(exec *executor) error {
	for {
		var h packetHeader
		if err := h.Deserialize(e.dec); err != nil {
			return fmt.Errorf("read header: %w", err)
		}

		switch h.typ {
		// counterpart requested us to run a function
		case rpcRequestPacketType:
			if err := e.processRequest(e.dec, exec); err != nil {
				return fmt.Errorf("processRequest: %w", err)
			}

		// response from a function we requested from counterpart
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

		// counterpart is closing
		case closingNowPacketType:
			return ErrEndpointClosedByCounterpart

		// counterpart informed us that context of some function it requested us to execute, has expired
		// we cancel corresponting worker's context
		// this doesn't mean the work will stop immediately. we don't have such power over goroutines.
		// goroutine needs to notice the context cancelation and stop itself
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

type Option func(*Endpoint)

func WithService(s ...irpcgen.Service) Option {
	return func(ep *Endpoint) {
		ep.RegisterService(s...)
	}
}

func WithParallelWorkers(parallelWorkers int) Option {
	return func(ep *Endpoint) {
		ep.parallelWorkers = parallelWorkers
	}
}
func WithParallelClientCalls(parallelClientCalls int) Option {
	return func(ep *Endpoint) {
		ep.parallelClientCalls = parallelClientCalls
	}
}
