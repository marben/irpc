package irpc_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/marben/irpc/pkg/irpc"
	irpctestpkg "github.com/marben/irpc/test"
	"github.com/marben/irpc/test/testtools"
)

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
		t.Errorf("server.Serve(): %+v", err)
	}

	time.Sleep(5 * time.Millisecond)
	if err := clientEp.Close(); !errors.Is(err, irpc.ErrEndpointClosed) {
		t.Fatalf("clientEp.Close(): %+v", err)
	}
}
