package irpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"
)

// the interface that is base for our rpc
type Math interface {
	Add(a, b int) int
}

// the implementation of our function
type MathHandler struct {
	resultSkew int // skew is added to result, to distinguish different versions of math
}

func (mh MathHandler) Add(a, b int) int {
	return a + b + mh.resultSkew
}

var _ Math = MathHandler{}
var _ Service = &MathIRpcService{}

type MathIRpcService struct {
	impl Math
}

func newMathIRpcService(impl Math) *MathIRpcService { return &MathIRpcService{impl: impl} }

func (ms *MathIRpcService) CallFunc(funcName string, params []byte) ([]byte, error) {
	switch funcName {
	case "Add":
		return ms.callAdd(params)

	default:
		return nil, fmt.Errorf("function '%s' doesn't exist on service '%s'", funcName, ms.Id())
	}
}

func (ms *MathIRpcService) callAdd(params []byte) ([]byte, error) {
	var p addParams
	if err := p.deserialize(params); err != nil {
		return nil, fmt.Errorf("failed to deserialize addParams: %w", err)
	}
	rtnVal := ms.impl.Add(p.A, p.B)
	rtn := addRtnVals{Res: rtnVal}
	return rtn.serialize()
}

const mathIrpcServiceName = "MathService"

// todo: maybe not needed?
func (*MathIRpcService) Id() string {
	return mathIrpcServiceName
}

var _ Math = &MathIRpcClient{}

type MathIRpcClient struct {
	ep *Endpoint
}

// todo: maybe we request error return from rpc functions?
func (mc *MathIRpcClient) Add(a int, b int) int {
	var params = addParams{A: a, B: b}
	paramsBytes, err := params.serialize()
	if err != nil {
		// serialization errors should not happen in generated code (can generate test as well)
		panic(fmt.Sprintf("serialization of Math.Add params failed with: %+v", err))
	}
	rtnBytes, err := mc.ep.CallRemoteFuncRaw(mathIrpcServiceName, "Add", paramsBytes)
	if err != nil {
		panic(fmt.Sprintf("irpc call return error: %+v", err))
	}
	var rtn addRtnVals
	if err := rtn.deserialize(rtnBytes); err != nil {
		panic(fmt.Sprintf("failed to deserialize addRtnVals from data: %+v", rtnBytes))
	}

	return rtn.Res
}

type addParams struct {
	A int
	B int
}

func (p addParams) serialize() ([]byte, error) {
	return json.Marshal(p)
}

func (p *addParams) deserialize(data []byte) error {
	return json.Unmarshal(data, p)
}

type addRtnVals struct {
	Res int
}

func (v addRtnVals) serialize() ([]byte, error) {
	return json.Marshal(v)
}

func (v *addRtnVals) deserialize(data []byte) error {
	return json.Unmarshal(data, v)
}

func TestEndpointRemoteFunc(t *testing.T) {
	pA, pB := NewDoubleEndedPipe()

	serviceEndpoint := NewEndpoint()
	go func() { serviceEndpoint.Serve(pA) }()

	clientEndpoint := NewEndpoint()
	go func() { clientEndpoint.Serve(pB) }()

	mathServiceB := newMathIRpcService(MathHandler{})
	if err := serviceEndpoint.RegisterServices(mathServiceB); err != nil {
		t.Fatalf("failed to register service: %+v", err)
	}

	clientA := MathIRpcClient{ep: clientEndpoint}
	res := clientA.Add(1, 2)
	if res != 3 {
		t.Fatalf("expected result of 3, but got %d", res)
	}
}

// at this moment, it is possible to register clients and call them, before Serve(conn) is called
// calls client calls should block until the Serve() is called, upon which they should fire
func TestCallBeforeServe(t *testing.T) {
	clientEndpoint, serviceEndpoint := NewEndpoint(), NewEndpoint()

	if err := serviceEndpoint.RegisterServices(newMathIRpcService(MathHandler{})); err != nil {
		t.Fatalf("registering math service failed with: %+v", err)
	}

	clientA := MathIRpcClient{ep: clientEndpoint}

	resC := make(chan int)

	go func() { resC <- clientA.Add(1, 4) }()
	// hopefully the client call will be made during this sleep
	time.Sleep(200 * time.Millisecond)

	pClient, pB := NewDoubleEndedPipe()

	// first we connect the client endpoint, which should trigger write to connection. but it's not read yet
	go func() {
		clientEndpoint.Serve(pClient)
	}()

	// another sleep to make sure the write was performed
	time.Sleep(200 * time.Millisecond)

	// finally we start the service's endpoint Serve()
	// now the service should read, process and return result to client
	go func() {
		serviceEndpoint.Serve(pB)
	}()

	// wait for the client to obtain result
	res := <-resC
	if res != 5 {
		t.Fatalf("expected result 5, but got %d", res)
	}
}

func TestServeAfterClose(t *testing.T) {
	c1, c2 := NewDoubleEndedPipe()
	ep1, ep2 := NewEndpoint(), NewEndpoint()
	go func() { ep1.Serve(c1) }()
	go func() { ep2.Serve(c2) }()

	if err := ep1.Close(); err != nil {
		t.Fatalf("unexpected close err: %v", err)
	}
	if err2 := ep1.Serve(c1); !errors.Is(err2, ErrEndpointClosed) {
		t.Fatalf("unexpected error on second close: %v", err2)
	}
	// close after close
	if err3 := ep1.Close(); !errors.Is(err3, ErrEndpointClosed) {
		t.Fatalf("second close returned: %v", err3)
	}
}

// performs remote func call both A->B and B->A
func TestBothSidesRemoteCall(t *testing.T) {
	pA, pB := NewDoubleEndedPipe()

	endpointA := NewEndpoint()
	// a is skewed by 1
	endpointA.RegisterServices(newMathIRpcService(MathHandler{resultSkew: 1}))
	go func() { endpointA.Serve(pA) }()

	endpointB := NewEndpoint()
	// b is skewed by 2
	endpointB.RegisterServices(newMathIRpcService(MathHandler{resultSkew: 2}))
	go func() { endpointB.Serve(pB) }()

	clientA := MathIRpcClient{ep: endpointA}
	clientB := MathIRpcClient{ep: endpointB}

	resFromB := clientA.Add(1, 2)
	if resFromB != 5 {
		t.Fatalf("service B (skewed by 2) returned %d instead of 5", resFromB)
	}

	resFromA := clientB.Add(1, 2)
	if resFromA != 4 {
		t.Fatalf("service A (skewed by 1) returned %d instead of 4", resFromA)
	}
}

func TestRegisterServiceTwice(t *testing.T) {
	ep := NewEndpoint()

	if err := ep.RegisterServices(newMathIRpcService(nil)); err != nil {
		t.Fatalf("registration of first service failed")
	}

	err := ep.RegisterServices(newMathIRpcService(nil))
	if !errors.Is(err, errServiceAlreadyRegistered) {
		t.Fatalf("expected error %v, but got: %v", errServiceAlreadyRegistered, err)
	}
}

// calls local function on an endpoint
func TestEndpointLocalFunc(t *testing.T) {
	// pA, _ := DoubleEndedPipe()
	e := NewEndpoint()
	ts := newMathIRpcService(MathHandler{})

	if err := e.RegisterServices(ts); err != nil {
		t.Fatalf("failed to register test service: %+v", err)
	}

	pBytes, err := addParams{A: 1, B: 2}.serialize()
	if err != nil {
		t.Fatalf("failed to serialize add parameters: %+v", err)
	}
	rtnBytes, err := e.callLocalFunc("MathService", "Add", pBytes)
	if err != nil {
		t.Fatalf("failed to call local func: %+v", err)
	}
	var rtn addRtnVals
	rtn.deserialize(rtnBytes)

	if rtn.Res != 3 {
		t.Fatalf("addition is '%d' instead of '3'", rtn.Res)
	}
}
