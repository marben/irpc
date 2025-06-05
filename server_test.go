package irpc_test

import (
	"context"
	"errors"
	"github.com/marben/irpc"
	"github.com/marben/irpc/cmd/irpc/test"
	"github.com/marben/irpc/cmd/irpc/test/testtools"
	"net"
	"testing"
	"time"
)

func TestServeOnMultipleListeners(t *testing.T) {
	// SERVER
	skew := 3
	service := testtools.NewTestServiceIRpcService(testtools.NewTestServiceImpl(skew))
	server := irpc.NewServer(service)

	l1, err := net.Listen("tcp", ":")
	if err != nil {
		t.Fatalf("failed to create server")
	}
	t.Logf("opened tcp listener1: %s", l1.Addr())
	serve1ErrC := make(chan error)
	t.Logf("running the server")
	go func() { serve1ErrC <- server.Serve(l1) }()

	l2, err := net.Listen("tcp", ":")
	if err != nil {
		t.Fatalf("failed to create server")
	}
	t.Logf("opened tcp listener2: %s", l2.Addr())
	serve2ErrC := make(chan error)
	t.Logf("running the server")
	go func() { serve2ErrC <- server.Serve(l2) }()

	// CLIENT 1
	t.Logf("net.Dial(%s)", l1.Addr())
	conn1, err := net.Dial("tcp", l1.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial(%s): %v", l1.Addr().String(), err)
	}

	c1Ep := irpc.NewEndpoint(conn1)

	client1, err := testtools.NewTestServiceIRpcClient(c1Ep)
	if err != nil {
		t.Fatalf("NewMathIrpcClient(): %v", err)
	}

	// CLIENT 2
	t.Logf("net.Dial(%s)", l2.Addr())
	conn2, err := net.Dial("tcp", l2.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial(%s): %v", l1.Addr().String(), err)
	}

	c2Ep := irpc.NewEndpoint(conn2)

	client2, err := testtools.NewTestServiceIRpcClient(c2Ep)
	if err != nil {
		t.Fatalf("NewMathIrpcClient(): %v", err)
	}

	t.Logf("making client1 call")
	res1, err := client1.DivCtxErr(context.Background(), 6, 3)
	if err != nil {
		t.Fatalf("client1.DivCtxErr(): %+v", err)
	}
	if res1 != 6/3+skew {
		t.Fatalf("unexpected client1 error: %d", res1)
	}

	t.Logf("making client2 call")
	res2, err := client2.DivCtxErr(context.Background(), 6, 2)
	if err != nil {
		t.Fatalf("client2.DivCtxErr(): %+v", err)
	}
	if res2 != 6/2+skew {
		t.Fatalf("unexpected client1 error: %d", res2)
	}

	t.Logf("closing listener1")
	if err := l1.Close(); err != nil {
		t.Fatalf("l1.Close(): %+v", err)
	}
	t.Logf("reading server.serve(l1) error")
	if err := <-serve1ErrC; !errors.Is(err, net.ErrClosed) {
		t.Fatal("server.serve(l1):", err)
	}

	t.Logf("trying a clinet1 call on closed listener, but still open connection")
	res12, err := client1.DivCtxErr(context.Background(), 6, 1)
	if err != nil {
		t.Fatalf("client1.DivCtxErr(): %+v", err)
	}
	if res12 != 6/1+skew {
		t.Fatalf("unexpected client1 error: %d", res1)
	}

	t.Log("closing connection 1")
	if err := conn1.Close(); err != nil {
		t.Fatal("conn1.Close():", err)
	}

	<-c1Ep.Done()
	t.Log("trying client call on closed connection")
	if _, err := client1.DivErr(2, 1); !errors.Is(err, irpc.ErrEndpointClosed) {
		t.Fatalf("client1.DivErr(): %+v", err)
	}

	t.Logf("testing client2 call on a still running connection")
	res22, err := client2.DivErr(8, 4)
	if err != nil {
		t.Fatalf("client2.DivErr(): %+v", err)
	}
	if res22 != 8/4+skew {
		t.Fatalf("unexpected res22: %d", res22)
	}

	t.Logf("closing server")
	if err := server.Close(); err != nil {
		t.Fatal("server.Close()", err)
	}

	<-c2Ep.Done()
	t.Log("testing client2 after server.Close()")
	// time.Sleep(5*time.Millisecond)
	if _, err := client2.DivErr(8, 4); !errors.Is(err, irpc.ErrEndpointClosed) {
		t.Fatal("client2.DivErr()", err)
	}
}

func TestTcpServerDialClose(t *testing.T) {
	// SERVER
	skew := 2
	service := testtools.NewTestServiceIRpcService(testtools.NewTestServiceImpl(skew))
	server := irpc.NewServer(service)

	l, err := net.Listen("tcp", ":")
	if err != nil {
		t.Fatalf("failed to create server")
	}
	t.Logf("opened tcp listener: %s", l.Addr())
	serveErrC := make(chan error)
	t.Logf("running the server")
	go func() { serveErrC <- server.Serve(l) }()

	// CLIENT
	t.Logf("net.Dial(%s)", l.Addr())
	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial(%s): %v", l.Addr().String(), err)
	}

	cEp := irpc.NewEndpoint(conn)

	client, err := testtools.NewTestServiceIRpcClient(cEp)
	if err != nil {
		t.Fatalf("NewMathIrpcClient(): %v", err)
	}

	// test call
	res, err := client.DivCtxErr(context.Background(), 6, 3)
	if err != nil {
		t.Fatalf("DivCtxErr(): %+v", err)
	}
	if res != 6/3+skew {
		t.Fatalf("unexpected result: %d", res)
	}

	t.Log("closing client endpoint")
	if err := cEp.Close(); err != nil {
		t.Fatalf("clientEp.Close(): %+v", err)
	}
}

func TestClientClosesOnServerClose(t *testing.T) {
	// SERVER
	skew := 2
	service := testtools.NewTestServiceIRpcService(testtools.NewTestServiceImpl(skew))
	server := irpc.NewServer(service)

	l, err := net.Listen("tcp", ":")
	if err != nil {
		t.Fatalf("failed to create server")
	}
	t.Logf("opened tcp listener: %s", l.Addr())
	serveErrC := make(chan error)
	t.Logf("running the server")
	go func() { serveErrC <- server.Serve(l) }()

	// CLIENT
	t.Logf("net.Dial(%s)", l.Addr())
	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial(%s): %v", l.Addr().String(), err)
	}

	cEp := irpc.NewEndpoint(conn)

	client, err := testtools.NewTestServiceIRpcClient(cEp)
	if err != nil {
		t.Fatalf("NewMathIrpcClient(): %v", err)
	}

	// test call
	res, err := client.DivCtxErr(context.Background(), 6, 3)
	if err != nil {
		t.Fatalf("DivCtxErr(): %+v", err)
	}
	if res != 6/3+skew {
		t.Fatalf("unexpected result: %d", res)
	}

	t.Logf("closing server")
	if err := server.Close(); err != nil {
		t.Fatalf("server.Close(): %+v", err)
	}

	<-cEp.Done()
	t.Log("making client call with closed server")
	if _, err := client.DivErr(6, 2); !errors.Is(err, irpc.ErrEndpointClosed) {
		t.Fatalf("unexpected error when calling client on closed server: %+v", err)
	}
}

func TestIrpcServerSimpleCall(t *testing.T) {
	l, err := net.Listen("tcp", ":")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	skew := 8
	mathService := irpctestpkg.NewMathIRpcService(irpctestpkg.MathImpl{Skew: skew})
	server := irpc.NewServer(mathService)

	localAddr := l.Addr().String()
	clientConn, err := net.Dial("tcp", localAddr)
	if err != nil {
		t.Fatalf("net.Dial(%s): %v", localAddr, err)
	}

	clientEp := irpc.NewEndpoint(clientConn)

	serveC := make(chan error)
	go func() { serveC <- server.Serve(l) }()

	client, err := irpctestpkg.NewMathIRpcClient(clientEp)
	if err != nil {
		t.Fatalf("NewMathIRpcClient(): %+v", err)
	}

	res, err := client.Add(1, 2)
	if err != nil {
		t.Fatalf("client.Add(1,2): %+v", err)
	}
	if res != 1+2+skew {
		t.Fatalf("unexpected result: %d", res)
	}

	if err := server.Close(); err != nil {
		t.Fatalf("Server.Close() returned: %+v", err)
	}
	if err := <-serveC; err != irpc.ErrServerClosed {
		t.Errorf("server.serve(): %+v", err)
	}

	time.Sleep(5 * time.Millisecond)
	if err := clientEp.Close(); !errors.Is(err, irpc.ErrEndpointClosed) {
		t.Fatalf("clientEp.Close(): %+v", err)
	}
}
