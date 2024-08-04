package testtools

import (
	"fmt"
	"net"

	"github.com/marben/irpc/pkg/irpc"
)

func CreateLocalTcpConnPipe() (net.Conn, net.Conn, error) {
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

type testLogger interface {
	Logf(format string, args ...any)
}

func CreateLocalTcpEndpoints(l testLogger) (*irpc.Endpoint, *irpc.Endpoint, error) {
	c1, c2, err := CreateLocalTcpConnPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create local tcp pipe")
	}
	ep1 := irpc.NewEndpoint()
	ep2 := irpc.NewEndpoint()

	// todo: turn to Fatalf on wrong Serve() return
	go func() {
		if err := ep1.Serve(c1); err != nil {
			l.Logf("ep1.Serve(): %v\n", err)
		}
	}()
	go func() {
		if err := ep2.Serve(c2); err != nil {
			l.Logf("ep2.Serve(): %v\n", err)
		}
	}()

	return ep1, ep2, nil
}
