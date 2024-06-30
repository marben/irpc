package irpc_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/marben/irpc/pkg/irpc"
	"github.com/marben/irpc/test/testtools"
)

// the interface that is base for our rpc
type Math interface {
	Add(a, b int) int
}

// the implementation of our function
type MathImpl struct {
	resultSkew int // skew is added to result, to distinguish different versions of math
}

func (mh MathImpl) Add(a, b int) int {
	return a + b + mh.resultSkew
}

var _ Math = MathImpl{}
var _ irpc.Service = &MathIRpcService{}

type MathIRpcService struct {
	impl Math
}

func newMathIRpcService(impl Math) *MathIRpcService { return &MathIRpcService{impl: impl} }

func (ms *MathIRpcService) GetFuncCall(funcId irpc.FuncId) (irpc.ArgDeserializer, error) {
	switch funcId {
	case mathIrpcFuncAddId:
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args addParams
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp addRtnVals
				resp.Res = ms.impl.Add(args.A, args.B)
				return resp
			}, nil
		}, nil
	default:
		return nil, fmt.Errorf("function '%v' doesn't exist on service '%s'", funcId, ms.Hash())
	}
}

var mathIrpcServiceHash = []byte("MathServiceHash")

func (*MathIRpcService) Hash() []byte {
	return mathIrpcServiceHash
}

var _ Math = &MathIRpcClient{}

const (
	mathIrpcFuncAddId irpc.FuncId = iota
)

type MathIRpcClient struct {
	ep *irpc.Endpoint
	id irpc.RegisteredServiceId
}

func NewMathIrpcClient(ep *irpc.Endpoint) (*MathIRpcClient, error) {
	id, err := ep.RegisterClient(mathIrpcServiceHash)
	if err != nil {
		return nil, fmt.Errorf("register failed: %w", err)
	}

	return &MathIRpcClient{ep, id}, nil
}

// todo: maybe we request error return from rpc functions?
func (mc *MathIRpcClient) Add(a int, b int) int {
	var params = addParams{A: a, B: b}
	var resp addRtnVals

	if err := mc.ep.CallRemoteFunc(mc.id, mathIrpcFuncAddId, params, &resp); err != nil {
		panic(fmt.Sprintf("callRemoteFunc failed: %v", err))
	}

	return resp.Res
}

type addParams struct {
	A int
	B int
}

func (p addParams) Serialize(w io.Writer) error {
	return json.NewEncoder(w).Encode(p)
}

func (p *addParams) Deserialize(d *irpc.Decoder) error {
	return json.NewDecoder(d.R).Decode(p)
}

type addRtnVals struct {
	Res int
}

func (v addRtnVals) Serialize(w io.Writer) error {
	return json.NewEncoder(w).Encode(v)
}

func (v *addRtnVals) Deserialize(d *irpc.Decoder) error {
	return json.NewDecoder(d.R).Decode(v)
}

func TestEndpointClientRegister(t *testing.T) {
	ep1, ep2, err := testtools.CreateLocalTcpEndpoints(t)
	if err != nil {
		t.Fatalf("create tcp: %v", err)
	}

	t.Logf("registering math service")
	if err := ep1.RegisterServices(newMathIRpcService(MathImpl{})); err != nil {
		t.Fatalf("service register: %v", err)
	}

	t.Logf("creating client")
	mathClient, err := NewMathIrpcClient(ep2)
	if err != nil {
		t.Fatalf("failed to create mathirpc client: %+v", err)
	}
	res := mathClient.Add(1, 2)
	if res != 3 {
		t.Fatalf("wrong result: %d", res)
	}
}

func TestEndpointRemoteFunc(t *testing.T) {
	pA, pB := testtools.NewDoubleEndedPipe()

	serviceEndpoint := irpc.NewEndpoint()
	go func() { serviceEndpoint.Serve(pA) }()

	clientEndpoint := irpc.NewEndpoint()
	go func() { clientEndpoint.Serve(pB) }()

	skew := 8
	mathServiceB := newMathIRpcService(MathImpl{resultSkew: skew})
	if err := serviceEndpoint.RegisterServices(mathServiceB); err != nil {
		t.Fatalf("failed to register service: %+v", err)
	}

	clientA, err := NewMathIrpcClient(clientEndpoint)
	if err != nil {
		t.Fatalf("failed to register client: %+v", err)
	}
	res := clientA.Add(1, 2)
	if res != 1+2+skew {
		t.Fatalf("expected result of 3, but got %d", res)
	}
}

// this blocks - for obvious reasons.
// todo: implement context cancelling
/*
func TestCallBeforeServe(t *testing.T) {
	clientEndpoint, serviceEndpoint := irpc.NewEndpoint(), irpc.NewEndpoint()

	if err := serviceEndpoint.RegisterServices(newMathIRpcService(MathImpl{})); err != nil {
		t.Fatalf("registering math service failed with: %+v", err)
	}

	clientA, err := NewMathIrpcClient(clientEndpoint)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	resC := make(chan int)

	go func() { resC <- clientA.Add(1, 4) }()
	// hopefully the client call will be made during this sleep
	time.Sleep(200 * time.Millisecond)

	pClient, pB := testtools.NewDoubleEndedPipe()

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
*/

func TestServeAfterClose(t *testing.T) {
	c1, c2 := testtools.NewDoubleEndedPipe()
	ep1, ep2 := irpc.NewEndpoint(), irpc.NewEndpoint()
	go func() { ep1.Serve(c1) }()
	go func() { ep2.Serve(c2) }()

	if err := ep1.Close(); err != nil {
		t.Fatalf("unexpected close err: %v", err)
	}
	if err2 := ep1.Serve(c1); !errors.Is(err2, irpc.ErrEndpointClosed) {
		t.Fatalf("unexpected error on second close: %v", err2)
	}
	// close after close
	if err3 := ep1.Close(); !errors.Is(err3, irpc.ErrEndpointClosed) {
		t.Fatalf("second close returned: %v", err3)
	}
}

// performs remote func call both A->B and B->A
func TestBothSidesRemoteCall(t *testing.T) {
	pA, pB := testtools.NewDoubleEndedPipe()

	endpointA := irpc.NewEndpoint()
	// a is skewed by 1
	endpointA.RegisterServices(newMathIRpcService(MathImpl{resultSkew: 1}))
	go func() { endpointA.Serve(pA) }()

	endpointB := irpc.NewEndpoint()
	// b is skewed by 2
	endpointB.RegisterServices(newMathIRpcService(MathImpl{resultSkew: 2}))
	go func() { endpointB.Serve(pB) }()

	clientA, err := NewMathIrpcClient(endpointA)
	if err != nil {
		t.Fatalf("new clientA: %+v", err)
	}
	clientB, err := NewMathIrpcClient(endpointB)
	if err != nil {
		t.Fatalf("new clientB: %+v", err)
	}

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
	ep := irpc.NewEndpoint()

	if err := ep.RegisterServices(newMathIRpcService(nil)); err != nil {
		t.Fatalf("registration of first service failed")
	}

	err := ep.RegisterServices(newMathIRpcService(nil))
	if !errors.Is(err, irpc.ErrServiceAlreadyRegistered) {
		t.Fatalf("expected error %v, but got: %v", irpc.ErrServiceAlreadyRegistered, err)
	}
}
