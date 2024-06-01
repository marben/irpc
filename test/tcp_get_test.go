package irpctestpkg

import (
	"fmt"
	"net"
	"testing"

	"github.com/marben/irpc/pkg/irpc"
)

func createLocalTcpConnPipe() (net.Conn, net.Conn, error) {
	l, err := net.Listen("tcp", ":")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create listener: %w", err)
	}
	defer l.Close()

	c1Ch := make(chan net.Conn)
	err1Ch := make(chan error)
	go func() {
		conn, err := l.Accept()
		if err != nil {
			err1Ch <- fmt.Errorf("failed to accept connection 1: %w", err)
		}
		c1Ch <- conn
	}()

	conn2, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to our listener: %w", err)
	}

	select {
	case err1 := <-err1Ch:
		return nil, nil, err1
	case conn1 := <-c1Ch:
		return conn1, conn2, nil
	}
}

func createLocalTcpEndpoints() (*irpc.Endpoint, *irpc.Endpoint, error) {
	c1, c2, err := createLocalTcpConnPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create local tcp pipe")
	}
	ep1 := irpc.NewEndpoint()
	ep2 := irpc.NewEndpoint()

	go func() { ep1.Serve(c1) }()
	go func() { ep2.Serve(c2) }()
	// go func() { log.Printf("serve ep1: %v", ep1.Serve(c1)) }()
	// go func() { log.Printf("serve ep2: %v", ep2.Serve(c2)) }()

	return ep1, ep2, nil
}

func TestTcpClientServer(t *testing.T) {
	ep1, ep2, err := createLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("failed to create local tcp endpoints: %v", err)
	}

	service := newTcpTestApiRpcService(tcpTestApiImpl{})
	ep1.RegisterServices(service)

	client := newTcpTestApiRpcClient(ep2)
	res, err := client.Div(4, 2)
	if err != nil {
		t.Fatalf("div failed: %v", err)
	}
	if res != 2 {
		t.Fatalf("wrong result: %f", res)
	}
	defer func() {
		if err := ep1.Close(); err != nil {
			t.Fatalf("ep1.Close(): %v", err)
		}
	}()
	defer func() {
		if err := ep2.Close(); err != nil {
			t.Fatalf("ep2.Close(): %v", err)
		}
	}()
}

func TestClosingConnection1(t *testing.T) {
	c1, c2, err := createLocalTcpConnPipe()
	if err != nil {
		t.Fatalf("failed to create local tcp pipe")
	}
	ep1 := irpc.NewEndpoint()
	ep2 := irpc.NewEndpoint()

	errC1 := make(chan error)
	errC2 := make(chan error)
	go func() { errC1 <- ep1.Serve(c1) }()
	go func() { errC2 <- ep2.Serve(c2) }()

	if err := c1.Close(); err != nil {
		t.Fatalf("failed to close connection c1: %v", err)
	}

	// both endpoint.Serve() errors out
	err1 := <-errC1
	err2 := <-errC2
	t.Logf("err1: %v", err1)
	t.Logf("err2: %v", err2)
}

func TestClosingConnection2(t *testing.T) {
	c1, c2, err := createLocalTcpConnPipe()
	if err != nil {
		t.Fatalf("failed to create local tcp pipe")
	}
	ep1 := irpc.NewEndpoint()
	ep2 := irpc.NewEndpoint()

	errC1 := make(chan error)
	errC2 := make(chan error)
	go func() { errC1 <- ep1.Serve(c1) }()
	go func() { errC2 <- ep2.Serve(c2) }()

	if err := c2.Close(); err != nil {
		t.Fatalf("failed to close connection c2: %v", err)
	}

	// both endpoint.Serve() errors out
	err1 := <-errC1
	err2 := <-errC2
	t.Logf("err1: %v", err1)
	t.Logf("err2: %v", err2)
}
