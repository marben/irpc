package irpc

import (
	"fmt"
	"net"
)

type Server struct {
	services []Service
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) RegisterService(services ...Service) {
	s.services = services
}

func (s *Server) Serve(lis net.Listener) error {
	for {
		conn, err := lis.Accept()
		if err != nil {
			return fmt.Errorf("listener.Accept(): %w", err)
		}

		ep := NewEndpoint()
		ep.RegisterServices(s.services...)
		go func() {
			// not sure what to do about errors (serve loop of http.Server doesn't seem to care, so we will follow suit for now)
			_ = ep.Serve(conn)
			// if err := ep.Serve(conn); err != nil {
			// 	log.Printf("ep.Serve(): %v", err)
			// }
		}()
	}
}

func (s *Server) Close() error {
	// todo: we need to close all listeners
	return fmt.Errorf("currently, Close is a no op") // todo: implement
}
