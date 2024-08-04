package irpc

import (
	"log"
	"net"
)

// returns a running endpoint
func DialTcp(addr string) (*Endpoint, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	ep := NewEndpoint()

	// todo: figure out err etc...
	go func() { log.Printf("endpoint Serve() finished with: %+v", ep.Serve(conn)) }()

	return ep, nil
}
