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

func CreateLocalTcpEndpoints() (*irpc.Endpoint, *irpc.Endpoint, error) {
	c1, c2, err := CreateLocalTcpConnPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create local tcp pipe")
	}
	ep1 := irpc.NewEndpoint(c1)
	ep2 := irpc.NewEndpoint(c2)
	return ep1, ep2, nil
}
