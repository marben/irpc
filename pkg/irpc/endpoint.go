package irpc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"sort"
	"sync"
	"sync/atomic"
)

var ErrServiceAlreadyRegistered = errors.New("service already registered")
var ErrEndpointClosed = errors.New("endpoint is closed")
var ErrServiceNotFound = errors.New("service not found")

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
	// clientRegistrationService is used to register other services (and give them their ids)
	// 0 is not used, so that uninitialized clients (with service id = 0), errors out
	clientRegistrationService RegisteredServiceId = iota + 1
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
	serviceHashes   map[string]RegisteredServiceId
	clientsServices map[RegisteredServiceId]Service
	nextServiceId   RegisteredServiceId
	servicesMux     sync.RWMutex

	closed atomic.Bool

	connCloser io.Closer     // close the connection
	bufWriter  *bufio.Writer // buffered Writer of the connection
	enc        *Encoder      // does encoding on the bufWriter
	connWMux   sync.Mutex    // locks connection and encoder

	reqNum    ReqNumT // serial number of our next request	// todo: sort out overflow (and test?)
	reqNumMux sync.Mutex

	awaitConnC chan struct{} // read from this channel will unblock once Endpoint.Serve() was called, meaning we are sending/receiving (and conn != nil)

	responseReaders    map[ReqNumT]func(d *Decoder)
	responseReadersMux sync.Mutex
}

func NewEndpoint() *Endpoint {
	ep := &Endpoint{
		serviceHashes:   make(map[string]RegisteredServiceId),
		clientsServices: make(map[RegisteredServiceId]Service),
		responseReaders: make(map[ReqNumT]func(d *Decoder)),
		awaitConnC:      make(chan struct{}),
		nextServiceId:   clientRegistrationService + 1,
	}

	// default service for registering clients
	regService := &clientRegisterService{ep: ep}
	ep.clientsServices[clientRegistrationService] = regService

	return ep
}

// immediately stops serving any requests
// Close closes the underlying connection and return conn.Close()'s
func (e *Endpoint) Close() error {
	if e.closed.Load() {
		return ErrEndpointClosed
	}

	e.connWMux.Lock()
	defer e.connWMux.Unlock()

	e.closed.Store(true)

	if e.connCloser == nil {
		// log.Println("close called before serve")
		// close was called before serve
		return nil
	}

	return e.connCloser.Close()
}

// registers client on remote endpoint
// if there is no corresponding service registered on remote node, returns error
// todo: rename to RegisterRemoteClient
func (e *Endpoint) RegisterClient(serviceHash []byte) (RegisteredServiceId, error) {
	req := clientRegisterReq{ServiceHash: serviceHash}
	resp := clientRegisterResp{}
	if err := e.CallRemoteFunc(clientRegistrationService, 0 /* currently there is just one function */, req, &resp); err != nil {
		return 0, fmt.Errorf("failed to register client %q: %w", serviceHash, err)
	}

	return resp.ServiceId, nil
}

func (e *Endpoint) getService(id RegisteredServiceId) (Service, error) {
	e.servicesMux.RLock()
	defer e.servicesMux.RUnlock()

	s, found := e.clientsServices[id]
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

func (e *Endpoint) CallRemoteFunc(serviceId RegisteredServiceId, funcId FuncId, reqData Serializable, respData Deserializable) error {
	if e.closed.Load() {
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
		e.clientsServices[id] = service
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
	<-e.awaitConnC

	e.connWMux.Lock()
	defer e.connWMux.Unlock()

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

func (e *Endpoint) readMsgs(dec *Decoder) error {
	errC := make(chan error)
	for {
		select {
		case err := <-errC:
			return err
		default:
		}

		var h packetHeader
		if err := h.Deserialize(dec); err != nil {
			return fmt.Errorf("failed to read header: %w", err)
		}

		switch h.typ {
		case rpcRequest:
			var req requestPacket
			if err := req.Deserialize(dec); err != nil {
				return fmt.Errorf("failed to read received request :%w", err)
			}

			service, err := e.getService(req.ServiceId)
			if err != nil {
				return fmt.Errorf("getService() %v: %w", req, err)
			}
			argDeser, err := service.GetFuncCall(req.FuncId)
			if err != nil {
				return fmt.Errorf("GetFuncCall() %v: %w", req, err)
			}
			exe, err := argDeser(dec)
			if err != nil {
				return fmt.Errorf("argDeserialize: %w", err)
			}
			go func() {
				resp := exe()

				if err := e.sendResponse(req.ReqNum, resp); err != nil {
					errC <- fmt.Errorf("failed to send response for request %d: %w", req.ReqNum, err)
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

func (e *Endpoint) Serve(conn io.ReadWriteCloser) error {
	if e.closed.Load() {
		return ErrEndpointClosed
	}
	defer e.closed.Store(true)

	e.connWMux.Lock()
	e.connCloser = conn
	e.bufWriter = bufio.NewWriter(conn)
	e.enc = NewEncoder(e.bufWriter)
	e.connWMux.Unlock()

	// unblock all waiting client requests
	close(e.awaitConnC)

	bufReader := bufio.NewReader(conn)

	dec := NewDecoder(bufReader)
	if err := e.readMsgs(dec); err != nil {
		// net.ErrClose seems to be returned when connection has been closed on our side
		// io.EOF seems to be returned, when connection has been closed on the other side
		if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
			// TODO: this creates dependency on net package. not sure if it's wanted + doesn't solve pipes, etc...
			// TODO: maybe simply implement connection close handshake and get rid of this bit altogether?
			return ErrEndpointClosed
		}
		if errClose := e.connCloser.Close(); errClose != nil {
			return errors.Join(err, fmt.Errorf("conn.Close(): %w", err))
		}
		return err
	}

	return nil
}
