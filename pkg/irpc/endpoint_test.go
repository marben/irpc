package irpc_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

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

func (p addParams) Serialize(e *irpc.Encoder) error {
	return json.NewEncoder(e.W).Encode(p) // todo: remove json
}

func (p *addParams) Deserialize(d *irpc.Decoder) error {
	return json.NewDecoder(d.R).Decode(p)
}

type addRtnVals struct {
	Res int
}

func (v addRtnVals) Serialize(e *irpc.Encoder) error {
	return json.NewEncoder(e.W).Encode(v) // todo: remove json
}

func (v *addRtnVals) Deserialize(d *irpc.Decoder) error {
	return json.NewDecoder(d.R).Decode(v)
}

func TestEndpointClientRegister(t *testing.T) {
	ep1, ep2, err := testtools.CreateLocalTcpEndpoints()
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

	if err := ep1.Close(); err != nil {
		t.Fatalf("ep1.Close(): %+v", err)
	}

	if err := ep2.Close(); err != nil {
		t.Fatalf("ep2.Close(): %+v", err)
	}
}

func TestEndpointRemoteFunc(t *testing.T) {
	pA, pB := testtools.NewDoubleEndedPipe()

	serviceEndpoint := irpc.NewEndpoint(pA)

	clientEndpoint := irpc.NewEndpoint(pB)

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

	if err := clientEndpoint.Close(); err != nil {
		t.Fatalf("client.Close(): %+v", err)
	}

	// clientEndpoint notifies us about Close and immediately closes underlying connection
	// if err := serviceEndpoint.Close(); err != irpc.ErrEndpointClosed {
	// 	t.Fatalf("unexpected close error: %+v", err)
	// }
	<-serviceEndpoint.Ctx.Done()
}

// performs remote func call both A->B and B->A
func TestBothSidesRemoteCall(t *testing.T) {
	pA, pB := testtools.NewDoubleEndedPipe()

	// a is skewed by 1
	endpointA := irpc.NewEndpoint(pA, newMathIRpcService(MathImpl{resultSkew: 1}))

	// b is skewed by 2
	endpointB := irpc.NewEndpoint(pB, newMathIRpcService(MathImpl{resultSkew: 2}))
	// irpc.NewEndpoint(pB, newMathIRpcService(MathImpl{resultSkew: 2}))

	log.Println("creating client a")
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
	if err := endpointB.Close(); err != nil {
		t.Fatalf("enpointB.Close(): %+v", err)
	}
	<-endpointA.Ctx.Done()
}

func TestLocalEndpointClose(t *testing.T) {
	conn1, conn2, err := testtools.CreateLocalTcpConnPipe()
	if err != nil {
		t.Fatalf("new tcp pipe: %+v", err)
	}

	serviceImpl := testtools.NewTestServiceImpl(99)
	serviceBackend := testtools.NewTestServiceIRpcService(serviceImpl)
	epRemote := irpc.NewEndpoint(conn1, serviceBackend)
	defer epRemote.Close()

	epLocal := irpc.NewEndpoint(conn2)

	client, err := testtools.NewTestServiceIRpcClient(epLocal)
	if err != nil {
		t.Fatalf("NewClient(): %+v", err)
	}
	res, err := client.DivErr(6, 3)
	if err != nil {
		t.Fatalf("client.Err(): %+v", err)
	}
	if res != 6/3+serviceImpl.Skew {
		t.Fatalf("unexpected result: %d", res)
	}

	if err := epLocal.Close(); err != nil {
		t.Fatalf("epLocal.Close(): %+v", err)
	}

	if _, err := client.DivErr(6, 2); err != irpc.ErrEndpointClosed {
		t.Fatalf("unexpected error: %+v", err)
	}

	if _, err := client.DivErr(6, 2); !errors.Is(err, irpc.ErrEndpointClosed) {
		t.Fatalf("unexpected error: %+v", err)
	}
}

// tests if waiting func calls errors out after endpoint close
func TestClosingServiceEpWithWaitingFuncCalls(t *testing.T) {
	serviceEp, clientEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create local tcp endpints: %+v", err)
	}

	unlockC := make(chan struct{})
	insideDivC := make(chan struct{}, 1)

	service := testtools.NewTestServiceImpl(0)
	// blocking function
	service.DivErrFunc = func(a, b int) (int, error) {
		//inform about our
		insideDivC <- struct{}{}

		// wait for unlock
		log.Println("DivErr() waiting for unlock")
		<-unlockC
		log.Println("DivErr() unlocked and returning")
		return a / b, nil
	}

	serviceEp.RegisterServices(testtools.NewTestServiceIRpcService(service))

	client, err := testtools.NewTestServiceIRpcClient(clientEp)
	if err != nil {
		t.Fatalf("new client: %+v", err)
	}

	rtnC := make(chan error)
	go func() {
		// this blocks on 'unlock' channel
		_, err := client.DivErr(6, 2)
		// however should err out on encpoint close
		rtnC <- err
	}()

	// make sure the DivErr() is running
	<-insideDivC

	if err := serviceEp.Close(); err != nil {
		t.Fatalf("ep1.Close(): %+v", err)
	}

	// the blocked client.DivErr should now err out
	if err := <-rtnC; err != irpc.ErrEndpointClosed {
		t.Fatalf("DivErr(): %+v", err)
	}

	close(unlockC)
}

func TestClosingClientEpWithWaitingFuncCalls(t *testing.T) {
	serviceEp, clientEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create local tcp endpints: %+v", err)
	}

	unlockC := make(chan struct{})
	insideDivC := make(chan struct{}, 1)

	service := testtools.NewTestServiceImpl(0)
	// blocking function
	service.DivErrFunc = func(a, b int) (int, error) {
		//inform about our
		insideDivC <- struct{}{}

		// wait for unlock
		log.Println("DivErr() waiting for unlock")
		<-unlockC
		log.Println("DivErr() unlocked and returning")
		return a / b, nil
	}

	serviceEp.RegisterServices(testtools.NewTestServiceIRpcService(service))

	client, err := testtools.NewTestServiceIRpcClient(clientEp)
	if err != nil {
		t.Fatalf("new client: %+v", err)
	}

	rtnC := make(chan error)
	go func() {
		// this blocks on 'unlock' channel
		_, err := client.DivErr(6, 2)
		// however should err out on encpoint close
		rtnC <- err
	}()

	// make sure the DivErr() is running
	<-insideDivC

	if err := clientEp.Close(); err != nil {
		t.Fatalf("ep1.Close(): %+v", err)
	}

	// the blocked client.DivErr should now err out
	if err := <-rtnC; err != irpc.ErrEndpointClosed {
		t.Fatalf("DivErr(): %+v", err)
	}

	close(unlockC)
}

func TestMaxWorkersNumber(t *testing.T) {
	serviceEp, clientEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create local tcp endpints: %+v", err)
	}

	service := testtools.NewTestServiceImpl(0)

	unlockC := make(chan struct{})
	resC := make(chan int)
	service.DivFunc = func(a, b int) int {
		resC <- a + b
		<-unlockC
		return a + b
	}

	serviceEp.RegisterServices(testtools.NewTestServiceIRpcService(service))

	client, err := testtools.NewTestServiceIRpcClient(clientEp)
	if err != nil {
		t.Fatalf("new client: %+v", err)
	}

	// start max parrallel workers + 1 calls in parallel goroutines
	wg := sync.WaitGroup{}
	for i := range irpc.ParallelWorkers + 1 {
		wg.Add(1)
		go func() {
			if res := client.Div(i, 3); res != i+3 {
				log.Fatalf("unexpected result: %d", res)
			}
			wg.Done()
		}()
	}

	// only max parallel workers should get started and send to resC
	for range irpc.ParallelWorkers {
		<-resC
	}

	// we don't have any means (atm) of making sure rpc call has arrived
	// instead we wait some time to make sure the func doesn't get called
	select {
	case res := <-resC:
		t.Fatalf("unexpectedly obtained result: %d", res)
	case <-time.After(100 * time.Millisecond):
		break
	}
	// wait timed out (means the last goroutine didn't get called)
	// closing the wait channel should unlock all gouroutines and let the last one pass
	close(unlockC)
	<-resC
	wg.Wait() // no reason really, just to be sure
}

// todo: uncomment
// func TestRegisterServiceTwice(t *testing.T) {
// 	ep := irpc.NewEndpoint()

// 	if err := ep.RegisterServices(newMathIRpcService(nil)); err != nil {
// 		t.Fatalf("registration of first service failed")
// 	}

// 	err := ep.RegisterServices(newMathIRpcService(nil))
// 	if !errors.Is(err, irpc.ErrServiceAlreadyRegistered) {
// 		t.Fatalf("expected error %v, but got: %v", irpc.ErrServiceAlreadyRegistered, err)
// 	}
// }
