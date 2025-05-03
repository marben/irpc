package irpc_test

import (
	"context"
	"errors"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/marben/irpc/pkg/irpc"
	"github.com/marben/irpc/test/testtools"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func TestEndpointClientRegister(t *testing.T) {
	ep1, ep2, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create tcp: %v", err)
	}

	t.Logf("registering test service")
	service := testtools.NewTestServiceIRpcService(testtools.NewTestServiceImpl(0))
	if err := ep1.RegisterServices(service); err != nil {
		t.Fatalf("service register: %v", err)
	}

	t.Logf("creating client")
	client, err := testtools.NewTestServiceIRpcClient(ep2)
	if err != nil {
		t.Fatalf("failed to create testservice client: %+v", err)
	}

	t.Logf("calling client func")
	res := client.Div(6, 2)
	if res != 3 {
		t.Fatalf("wrong result: %d", res)
	}

	t.Logf("%p: ep1 closing", ep1)
	if err := ep1.Close(); err != nil {
		t.Fatalf("ep1.Close(): %+v", err)
	}

	time.Sleep(100 * time.Millisecond) // wait for the close signal to arrive
	t.Log("ep2 cosing")
	if err := ep2.Close(); !errors.Is(err, irpc.ErrEndpointClosedByCounterpart) {
		t.Fatalf("ep2.Close(): %+v", err)
	}
}

func TestEndpointRemoteFunc(t *testing.T) {
	pA, pB := testtools.NewDoubleEndedPipe()

	serviceEndpoint := irpc.NewEndpoint(pA)

	clientEndpoint := irpc.NewEndpoint(pB)

	skew := 8
	service := testtools.NewTestServiceIRpcService(testtools.NewTestServiceImpl(skew))
	if err := serviceEndpoint.RegisterServices(service); err != nil {
		t.Fatalf("service register: %v", err)
	}

	clientA, err := testtools.NewTestServiceIRpcClient(clientEndpoint)
	if err != nil {
		t.Fatalf("failed to register client: %+v", err)
	}
	res := clientA.Div(4, 2)
	if res != 4/2+skew {
		t.Fatalf("expected %d, but got %d", 4/2+skew, res)
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

	skewA := 1
	serviceA := testtools.NewTestServiceIRpcService(testtools.NewTestServiceImpl(skewA))
	endpointA := irpc.NewEndpoint(pA, serviceA)

	skewB := 2
	serviceB := testtools.NewTestServiceIRpcService(testtools.NewTestServiceImpl(skewB))
	endpointB := irpc.NewEndpoint(pB, serviceB)

	t.Log("creating client a")
	clientA, err := testtools.NewTestServiceIRpcClient(endpointA)
	if err != nil {
		t.Fatalf("new clientA: %+v", err)
	}

	clientB, err := testtools.NewTestServiceIRpcClient(endpointB)
	if err != nil {
		t.Fatalf("new clientB: %+v", err)
	}

	resFromB := clientA.Div(4, 2)
	if resFromB != 4/2+skewB {
		t.Fatalf("service B returned %d", resFromB)
	}

	resFromA := clientB.Div(8, 4)
	if resFromA != 8/4+skewA {
		t.Fatalf("service A returned %d", resFromA)
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
		// log.Println("DivErr() waiting for unlock")
		<-unlockC
		// log.Println("DivErr() unlocked and returning")
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
	if err := <-rtnC; !errors.Is(err, irpc.ErrEndpointClosed) || !errors.Is(err, irpc.ErrEndpointClosedByCounterpart) {
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
		<-unlockC
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
	defer func() {
		if err := serviceEp.Close(); err != nil {
			t.Fatalf("failed to close service ep: %+v", err)
		}
	}()

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

// makes sure that context cancelation on clinet's side gets propagated to the service
func TestContextCancel(t *testing.T) {
	serviceEp, clientEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create local tcp endpoints: %+v", err)
	}

	service := testtools.NewTestServiceImpl(0)

	serviceEp.RegisterServices(testtools.NewTestServiceIRpcService(service))

	client, err := testtools.NewTestServiceIRpcClient(clientEp)
	if err != nil {
		t.Fatalf("new client: %+v", err)
	}

	// test normal operation
	res, err := client.DivCtxErr(context.Background(), 6, 3)
	if err != nil {
		t.Fatalf("client.Div.CtxErr: %+v", err)
	}
	if res != 6/3 {
		t.Fatalf("unexpected result: %d", res)
	}

	// now we will test cancelling context
	serviceErrC := make(chan error, 1)

	clientCtx, cancelClientCtx := context.WithCancelCause(context.Background())

	workerStartedC := make(chan struct{}, 1)
	// service function implementation
	// bloack until context is canceled.
	service.DivCtxErrFunc = func(ctx context.Context, a, b int) (int, error) {
		workerStartedC <- struct{}{}

		// block until ctx is canceled
		<-ctx.Done()
		causeErr := context.Cause(ctx)
		t.Logf("service div func: context canceled with err: '%s' and cause '%s'", ctx.Err(), causeErr)
		serviceErrC <- causeErr
		return 777, causeErr // errs out, but still returns val
	}

	clientErrC := make(chan error)
	go func() {
		// block until ctx is canceled
		res, err = client.DivCtxErr(clientCtx, 8, 2)
		if res != 777 { // we actually waited for the runner to spin up, so we should get proper result
			log.Fatalf("unexpected div result: %d", res)
		}

		// log.Printf("client obtained error: %v", err)
		clientErrC <- err
	}()

	clientCancelErr := errors.New("canceled from client side")
	<-workerStartedC
	// canceling client context should propagate to the service function
	cancelClientCtx(clientCancelErr)

	select {
	case <-time.After(2 * time.Second):
		t.Fatalf("waiting for client to cancel on context timed out")
	case clientErr := <-clientErrC:
		if clientErr != clientCancelErr {
			if clientErr.Error() != clientCancelErr.Error() {
				t.Fatalf("unexpected client context error: %+v", err)
			}
		}
	}

	// the service side of function should unblock too
	select {
	case <-time.After(1 * time.Second):
		t.Fatalf("waiting for service to cancel context timed out")
	case err := <-serviceErrC:
		if err.Error() != clientCancelErr.Error() {
			t.Fatalf("unexpected service context error: %+v", err)
		}
	}

	if err := serviceEp.Close(); err != nil {
		t.Fatalf("serviceEp.Close(): %+v", err)
	}
}

func TestServiceEndpointClosingEndsRunningWorkers(t *testing.T) {
	serviceEp, clientEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create local tcp endpoints: %+v", err)
	}

	service := testtools.NewTestServiceImpl(0)

	serviceEp.RegisterServices(testtools.NewTestServiceIRpcService(service))

	client, err := testtools.NewTestServiceIRpcClient(clientEp)
	if err != nil {
		t.Fatalf("new client: %+v", err)
	}

	serviceErrC := make(chan error, 1)
	serviceFuncStartC := make(chan struct{})

	// clientCtx, cancelClientCtx := context.WithCancelCause(context.Background())

	// service function implementation
	// bloack until context is canceled.
	service.DivCtxErrFunc = func(ctx context.Context, a, b int) (int, error) {
		serviceFuncStartC <- struct{}{}
		// block until ctx is canceled
		<-ctx.Done()
		causeErr := context.Cause(ctx)
		// t.Logf("service div func: context canceled with err: '%s' and cause '%s'", ctx.Err(), causeErr)
		serviceErrC <- causeErr
		return 777, causeErr // errs out, but still returns val
	}

	workersNumber := irpc.ParallelWorkers
	clientErrC := make(chan error)

	// all following workers should block
	for range workersNumber {
		go func() {
			// clients context never expires, but remote endpoint should still
			// cancel the workers on Close() call
			res, err := client.DivCtxErr(context.Background(), 8, 2)
			if res != 0 { // we should return the nil value(provided by client) as close dowsn't wait for the work runners to return values
				log.Fatalf("unexpected div result: %d", res)
			}

			// log.Printf("client obtained error: %v", err)
			clientErrC <- err
		}()
	}

	// make sure all service workers are running
	for range workersNumber {
		<-serviceFuncStartC
	}

	// trigger remote endpoint to close - should trigger close of everything
	if err := serviceEp.Close(); err != nil {
		t.Fatalf("serviceEp.Close(): %+v", err)
	}

	// client's should err out
	for range workersNumber {
		select {
		case <-time.After(2 * time.Second):
			t.Fatalf("waiting for client to cancel on context timed out")
		case clientErr := <-clientErrC:
			if !errors.Is(clientErr, irpc.ErrEndpointClosed) || !errors.Is(clientErr, irpc.ErrEndpointClosedByCounterpart) {
				t.Fatalf("unexpected client error: %v", clientErr)
			}
		}
	}

	// the service side of function should unblock too
	for range workersNumber {
		select {
		case <-time.After(1 * time.Second):
			t.Fatalf("waiting for service to cancel context timed out")
		case err := <-serviceErrC:
			if err != irpc.ErrEndpointClosed {
				t.Fatalf("unexpected service error: %v", err)
			}
		}
	}
}

func TestClientEndpointClosingEndsRunningWorkers(t *testing.T) {
	serviceEp, clientEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create local tcp endpoints: %+v", err)
	}

	service := testtools.NewTestServiceImpl(0)

	serviceEp.RegisterServices(testtools.NewTestServiceIRpcService(service))

	client, err := testtools.NewTestServiceIRpcClient(clientEp)
	if err != nil {
		t.Fatalf("new client: %+v", err)
	}

	serviceErrC := make(chan error, 1)
	serviceFuncStartC := make(chan struct{})

	// service function implementation
	// bloack until context is canceled.
	service.DivCtxErrFunc = func(ctx context.Context, a, b int) (int, error) {
		serviceFuncStartC <- struct{}{}
		// block until ctx is canceled
		<-ctx.Done()
		causeErr := context.Cause(ctx)
		// t.Logf("service div func: context canceled with err: '%s' and cause '%s'", ctx.Err(), causeErr)
		serviceErrC <- causeErr
		return 777, causeErr // errs out, but still returns val
	}

	workersNumber := irpc.ParallelWorkers
	clientErrC := make(chan error)

	// all following workers should block
	for range workersNumber {
		go func() {
			// clients context never expires, but remote endpoint should still
			// cancel the workers on Close() call
			res, err := client.DivCtxErr(context.Background(), 8, 2)
			if res != 0 { // we should return the nil value(provided by client) as close dowsn't wait for the work runners to return values
				log.Fatalf("unexpected div result: %d", res)
			}

			// log.Printf("client obtained error: %v", err)
			clientErrC <- err
		}()
	}

	// make sure all service workers are running
	for range workersNumber {
		<-serviceFuncStartC
	}

	// trigger local endpoint to close - should trigger close of everything
	if err := clientEp.Close(); err != nil {
		t.Fatalf("clientEp.Close(): %+v", err)
	}

	// client's should err out
	for range workersNumber {
		select {
		case <-time.After(2 * time.Second):
			t.Fatalf("waiting for client to cancel on context timed out")
		case clientErr := <-clientErrC:
			if clientErr != irpc.ErrEndpointClosed {
				t.Fatalf("unexpected client error: %v", err)
			}
		}
	}

	// the service side of function should unblock too
	for range workersNumber {
		select {
		case <-time.After(1 * time.Second):
			t.Fatalf("waiting for service to cancel context timed out")
		case err := <-serviceErrC:
			if !errors.Is(err, irpc.ErrEndpointClosed) || !errors.Is(err, irpc.ErrEndpointClosedByCounterpart) {
				t.Fatalf("unexpected service error: %v", err)
			}
		}
	}
}

// blocks available workers and then makes one more call, waiting for mutex to write/read
// makes sure expiration of client side context actually quits the mutex wait
func TestWaitingClientCallGetsCanceledOnContextTimeout(t *testing.T) {
	serviceEp, clientEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create local tcp endpoints: %+v", err)
	}

	service := testtools.NewTestServiceImpl(0)

	serviceEp.RegisterServices(testtools.NewTestServiceIRpcService(service))

	client, err := testtools.NewTestServiceIRpcClient(clientEp)
	if err != nil {
		t.Fatalf("new client: %+v", err)
	}

	// serviceErrC := make(chan error, 1)
	// serviceFuncStartC := make(chan struct{})

	serviceUnblockCtx, serviceUnblock := context.WithCancel(context.Background())

	service.DivCtxErrFunc = func(ctx context.Context, a, b int) (int, error) {
		<-serviceUnblockCtx.Done()
		return a / b, nil
	}

	// allocate all available workers
	// and one more, to block the readMsg loop on service side
	workersNumber := irpc.ParallelClientCalls + 1
	clientResC := make(chan int)
	for range workersNumber {
		go func() {
			// clients context never expires, but remote endpoint should still
			// cancel the workers on Close() call
			res, err := client.DivCtxErr(context.Background(), 8, 2)
			if err != nil {
				log.Fatalf("unexpected error of a properly made div")
			}
			if res != 4 {
				log.Fatalf("unexpected div result: %d", res)
			}

			// log.Printf("client obtained error: %v", err)
			clientResC <- res
		}()
	}
	time.Sleep(10 * time.Millisecond) // wait so that all worker are hopefully running
	// do the one call, that should not get to service and should time out on our side
	extraCallErr := make(chan struct{})
	go func() {
		timeoutCtx, _ := context.WithTimeout(context.Background(), 50*time.Millisecond)
		_, err := client.DivCtxErr(timeoutCtx, 8, 2)
		if err != context.DeadlineExceeded {
			log.Fatalf("unexpected error from blocked function")
		}
		extraCallErr <- struct{}{}
	}()
	select {
	case <-time.After(300 * time.Millisecond):
		t.Fatalf("wait timed out, but context should have canceled the client")
	case <-extraCallErr:
		t.Logf("cancelation on client side succeeded")
	}

	// just a cleanup + check all works as expected
	serviceUnblock()
	for range workersNumber {
		select {
		case res := <-clientResC:
			if res != 4 {
				t.Fatalf("unexpected result: %d", res)
			}
		case <-time.After(10 * time.Millisecond):
			t.Fatalf("wait expired")
		}
	}
}

// tests, whether dropped connection correctly closes both endpoints
func TestOutsideConnectionClose(t *testing.T) {
	c1, c2, err := testtools.CreateLocalTcpConnPipe()
	if err != nil {
		t.Fatalf("failed to create local tcp pipe")
	}
	ep1 := irpc.NewEndpoint(c1)
	ep2 := irpc.NewEndpoint(c2)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	serviceCallStartedC := make(chan struct{})
	callBlockC := make(chan struct{})

	t.Log("starting service 1")
	service1 := testtools.NewTestServiceImpl(1)
	service1.DivCtxErrFunc = func(ctx context.Context, a, b int) (int, error) {
		serviceCallStartedC <- struct{}{}
		<-callBlockC
		return 0, errors.New("service1 shouldn't have returned")
	}
	ep1.RegisterServices(testtools.NewTestServiceIRpcService(service1))

	t.Log("starting service 2")
	service2 := testtools.NewTestServiceImpl(2)
	service2.DivCtxErrFunc = func(ctx context.Context, a, b int) (int, error) {
		serviceCallStartedC <- struct{}{}
		<-callBlockC
		return 0, errors.New("service2 shouldn't have returned")
	}
	ep2.RegisterServices(testtools.NewTestServiceIRpcService(service2))

	clientErrs := make(chan error)

	t.Log("starting client1")
	client1, err := testtools.NewTestServiceIRpcClient(ep1)
	if err != nil {
		t.Fatalf("new client1: %v", err)
	}
	go func() {
		_, err := client1.DivCtxErr(ctx, 6, 3)
		clientErrs <- err
	}()

	t.Log("starting client2")
	client2, err := testtools.NewTestServiceIRpcClient(ep2)
	if err != nil {
		t.Fatalf("new client2: %v", err)
	}
	go func() {
		_, err := client2.DivCtxErr(ctx, 6, 3)
		clientErrs <- err
	}()

	t.Log("making sure both workers are running")
	<-serviceCallStartedC
	<-serviceCallStartedC
	t.Log("both workers are running")

	t.Log("closing connection")
	if err := c1.Close(); err != nil {
		t.Fatalf("c1.Close(): %+v", err)
	}

	t.Log("testing client's errors")
	if err := <-clientErrs; !errors.Is(err, irpc.ErrEndpointClosed) {
		t.Fatalf("unexpected client error: %v", err)
	}
	if err := <-clientErrs; !errors.Is(err, irpc.ErrEndpointClosed) {
		t.Fatalf("unexpected client error: %v", err)
	}

	t.Log("trying to close endpoints that were already closed")
	if err := ep1.Close(); !errors.Is(err, irpc.ErrEndpointClosed) {
		t.Fatalf("ep1.Close(): %v", err)
	}
	if err := ep2.Close(); !errors.Is(err, irpc.ErrEndpointClosed) {
		t.Fatalf("ep2.Close(): %v", err)
	}

	t.Log("trying to make a call on client")
	if _, err := client1.DivCtxErr(ctx, 1, 2); !errors.Is(err, irpc.ErrEndpointClosed) {
		t.Fatalf("client1.DivCtxErr(): %v", err)
	}
	if _, err := client2.DivCtxErr(ctx, 1, 2); !errors.Is(err, irpc.ErrEndpointClosed) {
		t.Fatalf("client2.DivCtxErr(): %v", err)
	}
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
