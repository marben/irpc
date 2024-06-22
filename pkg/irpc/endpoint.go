package irpc

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"sort"
	"sync"
	"sync/atomic"
)

var ErrServiceAlreadyRegistered = errors.New("service already registered")
var ErrEndpointClosed = errors.New("endpoint is closed")
var ErrServiceNotFound = errors.New("service not found")

// const DefaultMaxMsgLength = 10 * 1024 * 1024 // 10 MiB for now

// client registers to a service (by hash). upon registration, service is given 'RegisteredServiceId'
type RegisteredServiceId uint16
type FuncId uint16

const (
	// clientRegistrationService is used to register other services (and give them their ids)
	// 0 is not used, so that uninitialized clients (with service id = 0), errors out
	clientRegistrationService RegisteredServiceId = iota + 1
)

type Service interface {
	Hash() []byte // unique hash of the service
	GetFuncCall(funcId FuncId) (ArgDeserializer, error)
}

type ArgDeserializer func(r io.Reader) (FuncExecutor, error)
type FuncExecutor func() (Serializable, error)

// our generated code function calls' arguments implements Serializable/Deserializble
type Serializable interface {
	Serialize(w io.Writer) error
}
type Deserializable interface {
	Deserialize(r io.Reader) error
}

// Endpoint represents one side of a connection
// there needs to be a serving endpoint on both sides of connection for communication to work
type Endpoint struct {
	// maps registered service's hash to it's given id
	// todo: rename to serviceHashes
	services        map[string]RegisteredServiceId
	clientsServices map[RegisteredServiceId]Service
	nextServiceId   RegisteredServiceId
	servicesMux     sync.RWMutex

	closed atomic.Bool

	conn     io.ReadWriteCloser
	connWMux sync.Mutex

	reqNum    uint16 // serial number of our next request	// todo: sort out overflow (and test?)
	reqNumMux sync.Mutex

	awaitConnC chan struct{} // read from this channel will unblock once Endpoint.Serve() was called, meaning we are sending/receiving (and conn != nil)

	ourPendingRequests    map[uint16]chan responsePacket // pending request function awaiting response on given channel
	ourPendingRequestsMux sync.Mutex
}

func NewEndpoint() *Endpoint {
	ep := &Endpoint{
		services:           make(map[string]RegisteredServiceId),
		clientsServices:    make(map[RegisteredServiceId]Service),
		ourPendingRequests: make(map[uint16]chan responsePacket),
		awaitConnC:         make(chan struct{}),
		nextServiceId:      clientRegistrationService + 1,
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

	if e.conn == nil {
		// log.Println("close called before serve")
		// close was called before serve
		return nil
	}

	return e.conn.Close()
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

func (e *Endpoint) genNewRequestNumber() uint16 {
	e.reqNumMux.Lock()
	defer e.reqNumMux.Unlock()

	reqNum := e.reqNum
	e.reqNum += 1
	return reqNum
}

func (e *Endpoint) registerNewPendingRequestRtnChannel(reqNum uint16) chan responsePacket {
	c := make(chan responsePacket, 1)

	e.ourPendingRequestsMux.Lock()
	defer e.ourPendingRequestsMux.Unlock()

	e.ourPendingRequests[reqNum] = c
	return c
}

func (e *Endpoint) popPendingRequestRtnChannel(reqNum uint16) (chan responsePacket, error) {
	e.ourPendingRequestsMux.Lock()
	defer e.ourPendingRequestsMux.Unlock()

	c, found := e.ourPendingRequests[reqNum]
	if !found {
		return nil, fmt.Errorf("couldnt find pending request with reqNum '%d'", reqNum)
	}
	delete(e.ourPendingRequests, reqNum)

	return c, nil
}

func (e *Endpoint) CallRemoteFunc(serviceId RegisteredServiceId, funcId FuncId, reqData Serializable, resp Deserializable) error {
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

	ch := e.registerNewPendingRequestRtnChannel(reqNum)

	if err := e.serializeToConn(reqType, requestHeader, reqData); err != nil {
		e.popPendingRequestRtnChannel(reqNum)
		return fmt.Errorf("failed to write request to the connection")
	}

	respPacket := <-ch
	if respPacket.Err != "" {
		return fmt.Errorf("response error: %s", respPacket.Err)
	}

	r := bytes.NewBuffer(respPacket.Data)
	if err := resp.Deserialize(r); err != nil {
		return fmt.Errorf("response deserialize() failed: %w", err)
	}

	return nil
}

// todo: variadic is not such a great idea - what if only one error fails to register? what to return?
func (e *Endpoint) RegisterServices(services ...Service) error {
	e.servicesMux.Lock()
	defer e.servicesMux.Unlock()

	for _, service := range services {
		if _, found := e.services[string(service.Hash())]; found {
			return fmt.Errorf("%s: %w", service.Hash(), ErrServiceAlreadyRegistered)
		}

		id := e.nextServiceId
		e.nextServiceId++

		e.services[string(service.Hash())] = id
		e.clientsServices[id] = service
	}

	return nil
}

// prints hashes of all registered services
func (e *Endpoint) RegisteredServices() []string {
	e.servicesMux.RLock()
	defer e.servicesMux.RUnlock()

	ss := []string{}
	for s := range e.services {
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
		if err := s.Serialize(e.conn); err != nil {
			return fmt.Errorf("Serialize: %w", err)
		}
	}

	return nil
}

// respData is the actual serialized return data from the function
func (e *Endpoint) sendResponse(reqNum uint16, respData Serializable, err error) error {
	var errStr string
	if err != nil {
		errStr = err.Error()
	}

	// todo: serialize directly to the connection
	buf := bytes.NewBuffer(nil)
	if err := respData.Serialize(buf); err != nil {
		return fmt.Errorf("respData.Serialize(): %w", err)
	}
	resp := responsePacket{
		ReqNum: reqNum,
		Data:   buf.Bytes(),
		Err:    errStr,
	}

	header := packetHeader{
		typ: rpcResponse,
	}

	if err := e.serializeToConn(header, resp); err != nil {
		return fmt.Errorf("failed to serialize response to connection: %w", err)
	}

	return nil
}

func (e *Endpoint) readMsgs() error {
	errC := make(chan error)
	for {
		select {
		case err := <-errC:
			return err
		default:
		}

		var h packetHeader
		if err := h.Deserialize(e.conn); err != nil {
			return fmt.Errorf("failed to read header: %w", err)
		}

		switch h.typ {
		case rpcRequest:
			var req requestPacket
			if err := req.Deserialize(e.conn); err != nil {
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
			exe, err := argDeser(e.conn)
			if err != nil {
				return fmt.Errorf("argDeserialize: %w", err)
			}
			go func() {
				resp, err := exe()
				if err != nil {
					errC <- fmt.Errorf("func exec for req %v: %w", req, err)
					return
				}

				if err := e.sendResponse(req.ReqNum, resp, err); err != nil {
					errC <- fmt.Errorf("failed to send response for request %d: %w", req.ReqNum, err)
					return
				}
			}()

		case rpcResponse:
			var resp responsePacket
			if err := resp.Deserialize(e.conn); err != nil {
				return fmt.Errorf("failed to read response data:%w", err)
			}
			ch, err := e.popPendingRequestRtnChannel(resp.ReqNum)
			if err != nil {
				log.Printf("skipping response num %d because we failed to find corresponding request: %v", resp.ReqNum, err)
				continue
			}
			// a blocking write, but channels should be buffered
			ch <- resp

		default:
			return fmt.Errorf("unexpected msg type: %d", h.typ)
		}
	}
}

func (e *Endpoint) Serve(conn io.ReadWriteCloser) error {
	if e.closed.Load() {
		return ErrEndpointClosed
	}

	e.connWMux.Lock()
	e.conn = conn
	e.connWMux.Unlock()

	// unblock all waiting client requests
	close(e.awaitConnC)

	errC := make(chan error)
	go func() {
		errC <- e.readMsgs()
	}()

	return <-errC
}

var _ Service = &clientRegisterService{}

// service accomodating the client's registration
// only for endpoint's purposes
type clientRegisterService struct {
	ep *Endpoint
}

// GetFuncCall implements Service.
func (c *clientRegisterService) GetFuncCall(funcId FuncId) (ArgDeserializer, error) {
	switch funcId {
	case 0:
		return func(r io.Reader) (FuncExecutor, error) {
			var args clientRegisterReq
			if err := args.Deserialize(r); err != nil {
				return nil, err
			}
			return func() (Serializable, error) {
				c.ep.servicesMux.RLock()
				defer c.ep.servicesMux.RUnlock()

				var resp clientRegisterResp
				serviceId, found := c.ep.services[string(args.ServiceHash)]
				if !found {
					resp.Err = fmt.Errorf("couldn't find service hash %q", string(args.ServiceHash)).Error()
				} else {
					resp.ServiceId = serviceId
				}

				return resp, nil
			}, nil
		}, nil
	default:
		return nil, fmt.Errorf("registerService has no function %d", funcId)
	}
}

// Hash implements Service.
func (c *clientRegisterService) Hash() []byte {
	// currently, client register service is not versioned and it's hash is not used at all
	return nil
}
