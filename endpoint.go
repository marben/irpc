package irpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/marben/irpc/irpcgen"
	"io"
	"sync"
)

var ErrEndpointClosed = errors.New("endpoint is closed")
var ErrEndpointClosedByCounterpart = errors.Join(ErrEndpointClosed, errors.New("endpoint closed by counterpart"))
var ErrServiceNotFound = errors.New("service not found")
var errProtocolError = errors.New("protocol error")

const ServiceHashLen = 4

// todo: should be configurable for each endpoint
const (
	ParallelWorkers     = 3
	ParallelClientCalls = ParallelWorkers + 1
)

type Service interface {
	Id() []byte // unique id of the service
	GetFuncCall(funcId FuncId) (irpcgen.ArgDeserializer, error)
}

// Endpoint represents one side of a connection
// there needs to be a serving endpoint on both sides of connection for communication to work
type Endpoint struct {
	// maps serviceId to service
	services map[[ServiceHashLen]byte]Service

	serialize *serializer

	ourPendingRequests *ourPendingRequestsLog

	// closed on read loop end
	readLoopRunningC chan struct{}

	m sync.Mutex

	// context is Done() with the closing of endpoint
	Ctx       context.Context
	ctxCancel context.CancelCauseFunc
}

func NewEndpoint(conn io.ReadWriteCloser, services ...Service) *Endpoint {
	epCtx, endpointContextCancel := context.WithCancelCause(context.Background())

	ep := &Endpoint{
		services:           make(map[[ServiceHashLen]byte]Service),
		ourPendingRequests: newOurPendingRequestsLog(),
		serialize:          newSerializer(conn),
		readLoopRunningC:   make(chan struct{}, 1),
		Ctx:                epCtx,
		ctxCancel:          endpointContextCancel,
	}

	ep.RegisterServices(services...)

	go func() {
		ep.serve(epCtx, conn)
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

func (e *Endpoint) serve(ctx context.Context, conn io.ReadWriteCloser) {
	// our read loop
	defer func() {
		e.readLoopRunningC <- struct{}{}
	}()

	exec := newExecutor(ctx)

	readC := make(chan error, 1)
	go func() {
		readC <- e.readMsgs(conn, exec)
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
	e.signalOurClosingAndCloseConn(conn)
}

func (e *Endpoint) closeAlreadyCalled() bool {
	return e.Ctx.Err() != nil
}

func (e *Endpoint) signalOurClosingAndCloseConn(closer io.ReadCloser) {
	e.m.Lock()
	defer e.m.Unlock()

	// following calls may fail for various reasons, but we want to be sure, they were made
	e.serialize.serializePacketLocked(packetHeader{typ: closingNowPacketType})
	closer.Close()
}

// Close immediately stops serving any requests
// Close closes the underlying connection
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
func (e *Endpoint) getService(serviceHash []byte) (s Service, found bool) {
	e.m.Lock()
	defer e.m.Unlock()

	hashArray := [ServiceHashLen]byte{}
	copy(hashArray[:], serviceHash)
	s, found = e.services[hashArray]
	return s, found
}

func (e *Endpoint) sendRpcRequest(ctx context.Context, serviceId []byte, funcId FuncId, reqData irpcgen.Serializable, respData irpcgen.Deserializable) (ourPendingRequest, error) {
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

	if err := e.serialize.serializePacket(header, requestDef, reqData); err != nil {
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

	if err := e.serialize.serializePacket(header, ctxEndDef); err != nil {
		return err
	}

	return nil
}

func (e *Endpoint) CallRemoteFunc(ctx context.Context, serviceId []byte, funcId FuncId, reqData irpcgen.Serializable, respData irpcgen.Deserializable) error {
	pendingReq, err := e.sendRpcRequest(ctx, serviceId, funcId, reqData, respData)
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

func (e *Endpoint) RegisterServices(services ...Service) {
	e.m.Lock()
	defer e.m.Unlock()

	for _, s := range services {
		hashArray := [ServiceHashLen]byte{}
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

	argDeser, err := service.GetFuncCall(req.FuncId)
	if err != nil {
		return errors.Join(errProtocolError, fmt.Errorf("GetFuncCall() %v: %w", req, err))
	}
	funcExec, err := argDeser(dec)
	if err != nil {
		return fmt.Errorf("argDeserialize: %w", err)
	}

	// waits until worker slot is available (blocks here on too many long rpcs)
	err = exec.startServiceWorker(req.ReqNum, funcExec, e.serialize)
	if err != nil {
		return fmt.Errorf("new worker: %w", err)
	}

	return nil
}

// readMsgs is the main incoming messages processing loop
// returns context.Canceled if context has been canceled
func (e *Endpoint) readMsgs(conn io.ReadWriteCloser, exec *executor) error {
	dec := irpcgen.NewDecoder(conn)
	for {
		var h packetHeader
		if err := h.Deserialize(dec); err != nil {
			return fmt.Errorf("read header: %w", err)
		}

		switch h.typ {
		// counterpart requested us to run a function
		case rpcRequestPacketType:
			if err := e.processRequest(dec, exec); err != nil {
				return fmt.Errorf("processRequest: %w", err)
			}

		// response from a function we requested from counterpart
		case rpcResponsePacketType:
			var resp responsePacket
			if err := resp.Deserialize(dec); err != nil {
				return fmt.Errorf("failed to read response data:%w", err)
			}
			pr, err := e.ourPendingRequests.popPendingRequest(resp.ReqNum)
			if err != nil {
				return fmt.Errorf("request not found: %w", err)
			}

			pr.deserErrC <- pr.resp.Deserialize(dec)

		// counterpart is closing
		case closingNowPacketType:
			return ErrEndpointClosedByCounterpart

		// counterpart informs us that context of some function it requested us to execute, has expired
		// we cancel corresponting worker's context and hope response gets sent fast
		case ctxEndPacketType:
			var cancelRequest ctxEndPacket
			if err := cancelRequest.Deserialize(dec); err != nil {
				return fmt.Errorf("failed to deserialize context cancelation request: %w", err)
			}
			clientErr := errors.New(cancelRequest.ErrStr)
			exec.cancelRequest(cancelRequest.ReqNum, clientErr)

		default:
			return fmt.Errorf("unexpected packet type: %+v", h.typ)
		}
	}
}
