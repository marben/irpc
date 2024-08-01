package irpc_test

import (
	"log"
	"net"
	"testing"

	"github.com/marben/irpc/pkg/irpc"
	irpctestpkg "github.com/marben/irpc/test"
)

func TestIrpcServer(t *testing.T) {
	s := irpc.NewServer()

	l, err := net.Listen("tcp", ":")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	// defer l.Close()

	skew := 8
	mathService := irpctestpkg.NewMathIRpcService(irpctestpkg.MathImpl{Skew: skew})
	s.RegisterService(mathService)

	localAddr := l.Addr().String()
	clientConn, err := net.Dial("tcp", localAddr)
	if err != nil {
		t.Fatalf("net.Dial(%s): %v", localAddr, err)
	}

	clientEp := irpc.NewEndpoint()
	go func() {
		if err := clientEp.Serve(clientConn); err != nil {
			log.Fatalf("clientEp.Serve(): %+v", err)
		}
	}()

	serveC := make(chan error)
	go func() {
		serveC <- s.Serve(l)
	}()

	client, err := irpctestpkg.NewMathIRpcClient(clientEp)
	if err != nil {
		t.Fatalf("NewMathIRpcClient(): %+v", err)
	}

	res, err := client.Add(1, 2)
	if err != nil {
		t.Fatalf("client.Add(1,2): %+v", err)
	}
	if res != 1+2+skew {
		t.Fatalf("unexpected reult: %d", res)
	}

	if err := s.Close(); err != nil {
		t.Fatalf("Server.Close() returned: %+v", err)
	}
	if err := serveC; err != nil {
		t.Fatalf("s.Serve(): %+v", err)
	}
}
