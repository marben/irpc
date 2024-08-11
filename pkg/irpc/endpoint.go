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
)

var ErrServiceAlreadyRegistered = errors.New("service already registered")
var ErrEndpointClosed = errors.New("endpoint is closed")
var ErrServiceNotFound = errors.New("service not found")
var errProtocolError = errors.New("rpc protocol error")

type packetType uint8

const (
	rpcRequestPacket packetType = iota + 1
	rpcResponsePacket
	closingNowPacket // inform counterpart that i will immediately close the connection
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
type FuncExecutor func() Serializable

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

	responseReaders map[ReqNumT]func(d *Decoder)

	msgReadErrC chan error // error returned by message read go routine

	m sync.Mutex

	closing atomic.Bool // todo: maybe replace with Enpoint.Ctx

	// context ends with the closing of endpoint
	Ctx       context.Context // todo: this is a test. not sure about the design.
	ctxCancel context.CancelFunc
}

func NewEndpoint(conn io.ReadWriteCloser, services ...Service) *Endpoint {
	bufWriter := bufio.NewWriter(conn)
	ctx, cancel := context.WithCancel(context.Background())
	ep := &Endpoint{
		serviceHashes:   make(map[string]RegisteredServiceId),
		services:        make(map[RegisteredServiceId]Service),
		responseReaders: make(map[ReqNumT]func(d *Decoder)),
		nextServiceId:   clientRegistrationServiceId + 1,
		connCloser:      conn,
		bufWriter:       bufWriter,
		enc:             NewEncoder(bufWriter),
		msgReadErrC:     make(chan error, 1),
		Ctx:             ctx,
		ctxCancel:       cancel,
	}

	// default service for registering clients
	regService := &clientRegisterService{ep: ep}
	ep.services[clientRegistrationServiceId] = regService

	ep.RegisterServices(services...)

	// our read loop
	go func() {
		dec := NewDecoder(bufio.NewReader(conn))
		ep.msgReadErrC <- ep.readMsgs(dec)
	}()

	return ep
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
	if err := e.serializePacketToConnLocked(packetHeader{typ: closingNowPacket}); err != nil {
		return fmt.Errorf("notify counterpart: %w", err)
	}

	// no more conn writing after this
	e.closing.Store(true)
	e.ctxCancel() // this errs out waiting rpc calls

	// we are the ones closing the underlying connection
	// this makes read loop err out
	if err := e.connCloser.Close(); err != nil {
		return fmt.Errorf("connection.Close(): %w", err)
	}
	e.connCloser = nil

	// todo: we need to err out all pending requests

	// conn close triggers reader error
	// we don't care about the error
	<-e.msgReadErrC

	return nil
}

// registers client on remote endpoint
// if there is no corresponding service registered on remote node, returns error
// todo: rename to RegisterRemoteClient
func (e *Endpoint) RegisterClient(serviceHash []byte) (RegisteredServiceId, error) {
	req := clientRegisterReq{ServiceHash: serviceHash}
	resp := clientRegisterResp{}
	if err := e.CallRemoteFunc(clientRegistrationServiceId, 0 /* currently there is just one function */, req, &resp); err != nil {
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

func (e *Endpoint) sendRpcRequest(serviceId RegisteredServiceId, funcId FuncId, reqData Serializable, respData Deserializable) (chan error, error) {
	e.m.Lock()
	defer e.m.Unlock()

	reqNum := e.genNewRequestNumberLocked()

	reqType := packetHeader{
		typ: rpcRequestPacket,
	}

	requestHeader := requestPacket{
		ReqNum:    reqNum,
		ServiceId: serviceId,
		FuncId:    funcId,
	}

	if err := e.serializePacketToConnLocked(reqType, requestHeader, reqData); err != nil {
		return nil, err
	}

	deserErrC := make(chan error, 1)

	// upon response receive, this function will be called
	e.responseReaders[reqNum] = func(d *Decoder) {
		deserErrC <- respData.Deserialize(d)
	}

	return deserErrC, nil
}

func (e *Endpoint) CallRemoteFunc(serviceId RegisteredServiceId, funcId FuncId, reqData Serializable, respData Deserializable) error {
	deserErrC, err := e.sendRpcRequest(serviceId, funcId, reqData, respData)
	if err != nil {
		return err
	}

	select {
	case err := <-deserErrC:
		return err
	case <-e.Ctx.Done():
		return ErrEndpointClosed
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
		typ: rpcResponsePacket,
	}

	e.m.Lock()
	defer e.m.Unlock()

	if err := e.serializePacketToConnLocked(header, resp, respData); err != nil {
		return fmt.Errorf("failed to serialize response to connection: %w", err)
	}

	return nil
}

// readMsgs is the main incoming messages process loop
// returns wrapped errProtocolError if the error is on a protocol level
// returns wrapped reader error otherwise (conn closed etc...)
func (e *Endpoint) readMsgs(dec *Decoder) error {
	respCrrC := make(chan error) // todo: get rid of this channel. seems wrong
	for {
		select {
		case err := <-respCrrC:
			return err
		default:
		}

		var h packetHeader
		if err := h.Deserialize(dec); err != nil {
			return fmt.Errorf("read header: %w", err)
		}

		switch h.typ {
		case rpcRequestPacket:
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
			go func() {
				// call the function
				resp := funcExec()

				// todo: imho, this allows for multiple goroutines to keep writing to the channel even though it is already abandoned -> deadlock
				if err := e.sendResponse(req.ReqNum, resp); err != nil {
					respCrrC <- fmt.Errorf("failed to send response for request %d: %w", req.ReqNum, err)
					return
				}
			}()

		case rpcResponsePacket:
			var resp responsePacket
			if err := resp.Deserialize(dec); err != nil {
				return fmt.Errorf("failed to read response data:%w", err)
			}
			readFunc, err := e.popResponseReaderFunc(resp)
			if err != nil {
				return fmt.Errorf("request %q not found", resp.ReqNum)
			}
			readFunc(dec)

		case closingNowPacket:
			log.Println("counterpart is closing now")
			e.closing.Store(true)
			e.ctxCancel() // close waiting functions
			return ErrEndpointClosed

		default:
			return fmt.Errorf("unexpected msg type: %d", h.typ)
		}
	}
}

func (e *Endpoint) popResponseReaderFunc(resp responsePacket) (func(d *Decoder), error) {
	e.m.Lock()
	defer e.m.Unlock()

	f, found := e.responseReaders[resp.ReqNum]
	if !found {
		return nil, fmt.Errorf("response reader num %d not found", resp.ReqNum)
	}
	delete(e.responseReaders, resp.ReqNum)
	return f, nil
}
