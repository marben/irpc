package irpc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"sort"
	"sync"
)

var ErrServiceAlreadyRegistered = errors.New("service already registered")
var ErrEndpointClosed = errors.New("endpoint is closed")
var ErrServiceNotFound = errors.New("service not found")
var errProtocolError = errors.New("rpc protocol error")

type packetType uint8

const (
	rpcRequest packetType = iota + 1
	rpcResponse
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
	servicesMux   sync.RWMutex

	connCloser io.Closer     // close the connection
	bufWriter  *bufio.Writer // buffered Writer of the connection	// only used for flushing. rest of code should use encoder
	enc        *Encoder      // does encoding on the bufWriter
	wMux       sync.Mutex    // todo: make it one shared mutex for whole endpoint?

	dec *Decoder

	reqNum    ReqNumT    // serial number of our next request	// todo: sort out overflow (and test?)
	reqNumMux sync.Mutex // todo: maybe not necessary as request requires writer mutex lock - can be shared?

	responseReaders    map[ReqNumT]func(d *Decoder)
	responseReadersMux sync.Mutex

	// closed    bool
	serveErrC chan error
}

func NewEndpoint(conn io.ReadWriteCloser, services ...Service) *Endpoint {
	bufWriter := bufio.NewWriter(conn)
	ep := &Endpoint{
		serviceHashes:   make(map[string]RegisteredServiceId),
		services:        make(map[RegisteredServiceId]Service),
		responseReaders: make(map[ReqNumT]func(d *Decoder)),
		nextServiceId:   clientRegistrationServiceId + 1,
		connCloser:      conn,
		bufWriter:       bufWriter,
		enc:             NewEncoder(bufWriter),
		dec:             NewDecoder(bufio.NewReader(conn)),
		serveErrC:       make(chan error, 1),
	}

	// default service for registering clients
	regService := &clientRegisterService{ep: ep}
	ep.services[clientRegistrationServiceId] = regService

	ep.RegisterServices(services...)

	go func() { ep.serveErrC <- ep.serve() }()

	return ep
}

// Close immediately stops serving any requests
// Close closes the underlying connection
// Close returns ErrEndpointClosed if the Endpoint was already closed
func (e *Endpoint) Close() error {
	e.wMux.Lock()
	defer e.wMux.Unlock()

	if e.connCloser == nil {
		return ErrEndpointClosed
	}

	// e.wMux.Lock()
	// defer e.wMux.Unlock()

	// if e.isClosed() {
	// 	return ErrEndpointClosed
	// }

	// if e.closed.Load() {
	// 	return ErrEndpointClosed
	// }

	// e.encMux.Lock()
	// defer e.encMux.Unlock()

	// e.closed.Store(true)

	// if e.connCloser == nil {
	// log.Println("close called before serve")
	// close was called before serve
	// return nil
	// }

	/*
		e.wMux.Lock()
		if e.connCloser == nil {
			// serve has already deleted the conn => conn already closed
			e.wMux.Unlock()
			return ErrEndpointClosed
		}

		// connection is still running, so we close it
		// that triggers the running e.serve(), which does the cleanup and send err to error channel
		if err := e.connCloser.Close(); err != nil {
			return fmt.Errorf("connection.Close(): %w", err)
		}
		e.wMux.Unlock()
		// log.Println("setting connCloser to nil")
		// e.connCloser = nil
	*/

	// closing connection signs endpint.serve() that it has to quit by causing the reads to err
	// if err := e.closeConnIfOpened(); err != nil {
	// 	return err
	// }

	if err := e.connCloser.Close(); err != nil {
		return fmt.Errorf("conn.Close(): %w", err)
	}

	e.connCloser = nil

	// error from serve()
	if err := <-e.serveErrC; err != nil {
		return fmt.Errorf("serve(): %w", err)
	}

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
	e.servicesMux.RLock()
	defer e.servicesMux.RUnlock()

	s, found := e.services[id]
	if !found {
		return nil, ErrServiceNotFound
	}
	return s, nil
}

func (e *Endpoint) genNewRequestNumber() ReqNumT {
	e.reqNumMux.Lock()
	defer e.reqNumMux.Unlock()

	reqNum := e.reqNum
	e.reqNum += 1
	return reqNum
}

func (e *Endpoint) isClosed() bool {
	return e.connCloser == nil
}

func (e *Endpoint) CallRemoteFunc(serviceId RegisteredServiceId, funcId FuncId, reqData Serializable, respData Deserializable) error {
	if e.isClosed() {
		return ErrEndpointClosed
	}

	reqNum := e.genNewRequestNumber()

	reqType := packetHeader{
		typ: rpcRequest,
	}

	requestHeader := requestPacket{
		ReqNum:    reqNum,
		ServiceId: serviceId,
		FuncId:    funcId,
	}

	c := make(chan error, 1)
	e.responseReadersMux.Lock()
	e.responseReaders[reqNum] = func(d *Decoder) {
		c <- respData.Deserialize(d)
	}
	e.responseReadersMux.Unlock()

	// todo: fail the whole endpoint?
	if err := e.serializeToConn(reqType, requestHeader, reqData); err != nil {
		e.responseReadersMux.Lock()
		delete(e.responseReaders, reqNum)
		e.responseReadersMux.Unlock()
		return fmt.Errorf("write request: %w", err)
	}

	// todo: should we close the endpoint on error?
	if err := <-c; err != nil {
		return fmt.Errorf("deserialization failed: %w", err)
	}

	return nil
}

// todo: variadic is not such a great idea - what if only one error fails to register? what to return?
func (e *Endpoint) RegisterServices(services ...Service) error {
	e.servicesMux.Lock()
	defer e.servicesMux.Unlock()

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

// prints hashes of all registered services
func (e *Endpoint) RegisteredServices() []string {
	e.servicesMux.RLock()
	defer e.servicesMux.RUnlock()

	ss := []string{}
	for s := range e.serviceHashes {
		ss = append(ss, s)
	}

	sort.Strings(ss)

	return ss
}

// func (e *Endpoint) serializeToConn(header packetHeader, data Serializable) error {
func (e *Endpoint) serializeToConn(data ...Serializable) error {
	// make sure we have a connection
	// <-e.awaitConnC

	e.wMux.Lock()
	defer e.wMux.Unlock()

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
		typ: rpcResponse,
	}

	if err := e.serializeToConn(header, resp, respData); err != nil {
		return fmt.Errorf("failed to serialize response to connection: %w", err)
	}

	return nil
}

// processMsgs is the main incoming messages process loop
// returns wrapped errProtocolError if the error is on a protocol level
// returns wrapped reader error otherwise (conn closed etc...)
func (e *Endpoint) processMsgs(dec *Decoder) error {
	respCrrC := make(chan error)
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
		case rpcRequest:
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

		case rpcResponse:
			var resp responsePacket
			if err := resp.Deserialize(dec); err != nil {
				return fmt.Errorf("failed to read response data:%w", err)
			}
			readFunc, err := e.popResponseReaderFunc(resp)
			if err != nil {
				return fmt.Errorf("request %q not found", resp.ReqNum)
			}
			readFunc(dec)

		default:
			return fmt.Errorf("unexpected msg type: %d", h.typ)
		}
	}
}

func (e *Endpoint) popResponseReaderFunc(resp responsePacket) (func(d *Decoder), error) {
	e.responseReadersMux.Lock()
	defer e.responseReadersMux.Unlock()

	f, found := e.responseReaders[resp.ReqNum]
	if !found {
		return nil, fmt.Errorf("response reader num %d not found", resp.ReqNum)
	}
	delete(e.responseReaders, resp.ReqNum)
	return f, nil
}

// serve returns ErrEndpointClosed on normal close
func (e *Endpoint) serve() error {
	// if e.closed.Load() {
	// 	return ErrEndpointClosed
	// }
	// defer e.closed.Store(true)

	// e.connWMux.Lock()
	// e.connCloser = conn
	// e.bufWriter = bufio.NewWriter(conn)
	// e.enc = NewEncoder(e.bufWriter)
	// e.connWMux.Unlock()

	// unblock all waiting client requests
	// close(e.awaitConnC)

	// bufReader := bufio.NewReader(conn)

	// dec := NewDecoder(bufReader)
	// defer func() {
	// 	e.wMux.Lock()
	// 	defer e.wMux.Unlock()

	// 	e.connCloser = nil
	// }()

	if err := e.processMsgs(e.dec); err != nil {

		// net.ErrClose seems to be returned when connection has been closed on our side
		// io.EOF seems to be returned, when connection has been closed on the other side
		if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
			// TODO: this creates dependency on net package. not sure if it's wanted + doesn't solve pipes, etc...
			// TODO: maybe simply implement connection close handshake and get rid of this bit altogether?
			// these are sign of correct connecton closing, co no error
			return nil
		}

		// uncertain about the error, we call connection.Close(), just in case
		return err
		// log.Printf("readMsgs: %+v\net.connCloser: %+v", err, e.connCloser)
		// if !e.isClosed() {
		// 	if errClose := e.connCloser.Close(); errClose != nil {
		// 		return errors.Join(err, fmt.Errorf("conn.Close(): %w", err))
		// 	}
		// }
	}

	return nil
}
