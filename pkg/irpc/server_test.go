package irpc_test

import (
	"net"
	"testing"

	"github.com/marben/irpc/pkg/irpc"
	irpctestpkg "github.com/marben/irpc/test"
)

func TestTcpServerDial(t *testing.T) {

}

func TestIrpcServer(t *testing.T) {
	server := irpc.NewServer()

	l, err := net.Listen("tcp", ":")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	skew := 8
	mathService := irpctestpkg.NewMathIRpcService(irpctestpkg.MathImpl{Skew: skew})
	server.RegisterService(mathService)

	localAddr := l.Addr().String()
	clientConn, err := net.Dial("tcp", localAddr)
	if err != nil {
		t.Fatalf("net.Dial(%s): %v", localAddr, err)
	}

	clientEp := irpc.NewEndpoint()
	clientErrC := make(chan error)
	go func() { clientErrC <- clientEp.Serve(clientConn) }()

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
	if err := <-clientErrC; err != irpc.ErrEndpointClosed {
		t.Errorf("clientEp.Serve(): %+v", err)
	}
}
