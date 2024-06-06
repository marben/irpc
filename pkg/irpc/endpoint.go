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

var errServiceAlreadyRegistered = errors.New("service already registered")
var ErrEndpointClosed = errors.New("endpoint is closed")

const DefaultMaxMsgLength = 10 * 1024 * 1024 // 10 MiB for now

type Service interface {
	Id() string
	CallFunc(funcName string, params []byte) ([]byte, error)
}

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
	services    map[string]Service
	servicesMux sync.RWMutex

	closed atomic.Bool

	conn     io.ReadWriteCloser
	connWMux sync.Mutex

	reqNum    uint16 // serial number of our next request	// todo: sort out overflow (and test?)
	reqNumMux sync.Mutex

	awaitConnC chan struct{} // read from this channel will unblock once Endpoint.Serve() was called, meaning we are sending/receiving (and conn != nil)

	ourPendingRequests    map[uint16]chan responsePacket // pending request function awaiting response on given channel
	ourPendingRequestsMux sync.Mutex

	MaxMsgLen int
}

func NewEndpoint() *Endpoint {
	return &Endpoint{
		services:           make(map[string]Service),
		ourPendingRequests: make(map[uint16]chan responsePacket),
		awaitConnC:         make(chan struct{}),
		MaxMsgLen:          DefaultMaxMsgLength,
	}
}

// immediately stops serving any requests
// Close closes the underlying connection, if it is io.Closer in which case it return's it's return error
func (e *Endpoint) Close() error {
	if e.closed.Load() {
		return ErrEndpointClosed
	}

	e.connWMux.Lock()
	defer e.connWMux.Unlock()

	e.closed.Store(true)

	if closer, ok := e.conn.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}

func (e *Endpoint) getService(serviceId string) (s Service, found bool) {
	e.servicesMux.RLock()
	defer e.servicesMux.RUnlock()

	s, found = e.services[serviceId]
	return
}

func (e *Endpoint) callLocalFunc(serviceId, funcName string, params []byte) ([]byte, error) {
	service, found := e.getService(serviceId)
	if !found {
		return nil, fmt.Errorf("could not find service '%s'", serviceId)
	}

	rtn, err := service.CallFunc(funcName, params)
	if err != nil {
		return nil, fmt.Errorf("failed to call function '%s/%s': %w", serviceId, funcName, err)
	}

	return rtn, nil
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

func (e *Endpoint) CallRemoteFunc(serviceName, funcName string, req Serializable, resp Deserializable) error {
	if e.closed.Load() {
		return ErrEndpointClosed
	}

	b := bytes.NewBuffer(nil)
	err := req.Serialize(b)
	if err != nil {
		return fmt.Errorf("failed to serialize request: %w", err)
	}

	respBytes, err := e.CallRemoteFuncRaw(serviceName, funcName, b.Bytes())
	if err != nil {
		return fmt.Errorf("remote call failed: %w", err)
	}

	r := bytes.NewBuffer(respBytes)
	if err := resp.Deserialize(r); err != nil {
		return fmt.Errorf("response deserialize() failed: %w", err)
	}

	return nil
}

// same as call remote func, but doesn't do the serialization/deserialization for you
// TODO: eventually make non-public (i guess)
func (e *Endpoint) CallRemoteFuncRaw(serviceName, funcName string, params []byte) ([]byte, error) {
	if e.closed.Load() { // todo: should not be necessary, once it's not public
		return nil, ErrEndpointClosed
	}
	// log.Printf("irpc: sending reqest num %d for func %s", e.reqNum, funcName)

	reqNum := e.genNewRequestNumber()

	request := requestPacket{
		ReqNum:     reqNum,
		ServiceId:  serviceName,
		FuncNameId: funcName,
		Data:       params,
	}

	header := packetHeader{
		typ: rpcRequest,
	}

	ch := e.registerNewPendingRequestRtnChannel(reqNum)

	if err := e.serializeToConn(header, request); err != nil {
		e.popPendingRequestRtnChannel(reqNum)
		return nil, fmt.Errorf("failed to write request to the connection")
	}

	// log.Println("waiting for response")
	resp := <-ch
	// log.Printf("obtained response from a channel! :%+v", resp)
	if resp.Err != "" {
		return resp.Data, fmt.Errorf("response error: %s", resp.Err)
	}

	return resp.Data, nil
}

// todo: really variadic?
func (e *Endpoint) RegisterServices(services ...Service) error {
	e.servicesMux.Lock()
	defer e.servicesMux.Unlock()

	for _, service := range services {
		if _, found := e.services[service.Id()]; found {
			return fmt.Errorf("%s: %w", service.Id(), errServiceAlreadyRegistered)
		}

		e.services[service.Id()] = service
	}

	return nil
}

func (e *Endpoint) RegisteredServices() []string {
	e.servicesMux.RLock()
	defer e.servicesMux.RUnlock()

	ss := []string{}
	for _, s := range e.services {
		ss = append(ss, s.Id())
	}

	sort.Strings(ss)

	return ss
}

func (e *Endpoint) serializeToConn(header packetHeader, data Serializable) error {
	// make sure we have a connection
	<-e.awaitConnC

	e.connWMux.Lock()
	defer e.connWMux.Unlock()

	if err := header.Serialize(e.conn); err != nil {
		return fmt.Errorf("header.Serialize: %w", err)
	}

	if err := data.Serialize(e.conn); err != nil {
		return fmt.Errorf("data.Serialize(conn): %w", err)
	}

	return nil
}

// respData is the actual serialized return data from the function
func (e *Endpoint) sendResponse(reqNum uint16, respData []byte, err error) error {
	var errStr string
	if err != nil {
		errStr = err.Error()
	}
	resp := responsePacket{
		ReqNum: reqNum,
		Data:   respData,
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

// todo: pass in context to exit on
func (e *Endpoint) processIncomingRequest(req requestPacket) error {
	resp, err := e.callLocalFunc(req.ServiceId, req.FuncNameId, req.Data)
	if err != nil {
		if errResp := e.sendResponse(req.ReqNum, nil, err); errResp != nil {
			return fmt.Errorf("failed to respond with error '%s': %v", err, errResp)
		}
		return nil
	}
	if err := e.sendResponse(req.ReqNum, resp, nil); err != nil {
		return fmt.Errorf("failed to send response for request %d: %w", req.ReqNum, err)
	}

	return nil
}

func (e *Endpoint) readMsgs() error {
	for {
		// read the header
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
			go func() {
				if err := e.processIncomingRequest(req); err != nil {
					log.Printf("processing request %d failed with %+v", req.ReqNum, err)
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

	e.conn = conn

	// unblock all waiting client requests
	close(e.awaitConnC)

	errC := make(chan error, 1)
	go func() {
		if err := e.readMsgs(); err != nil {
			errC <- err
		}
		errC <- nil
	}()

	return <-errC
}
