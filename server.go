package irpc

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"

	"github.com/marben/irpc/irpcgen"
)

var ErrServerClosed error = errors.New("irpc: server closed")

type Server struct {
	services []irpcgen.Service // immutable

	clients    map[*Endpoint]struct{}
	clientsMux sync.Mutex
	clientsWg  sync.WaitGroup

	inShutdown atomic.Bool

	listeners    map[net.Listener]struct{} // todo: should we store pointers in a similar fashion std http server does?
	listenersMux sync.Mutex
	listenersWg  sync.WaitGroup
}

func NewServer(services ...irpcgen.Service) *Server {
	return &Server{
		listeners: make(map[net.Listener]struct{}),
		clients:   make(map[*Endpoint]struct{}),
		services:  services,
	}
}

func (s *Server) isShuttingDown() bool {
	return s.inShutdown.Load()
}

// returns error if server is shutting down
func (s *Server) addListener(l net.Listener) error {
	if s.isShuttingDown() {
		return ErrServerClosed
	}

	s.listenersMux.Lock()
	defer s.listenersMux.Unlock()

	s.listeners[l] = struct{}{}
	s.listenersWg.Add(1)

	return nil
}

func (s *Server) rmListener(l net.Listener) {
	s.listenersMux.Lock()
	defer s.listenersMux.Unlock()

	delete(s.listeners, l)
	s.listenersWg.Done()
}

// Serve always returns a non-nil error. After [Server.Close], the returned error is [ErrServerClosed]
func (s *Server) Serve(lis net.Listener) error {
	// log.Printf("irpc server: serving on port %v", lis.Addr())
	if err := s.addListener(lis); err != nil {
		return err
	}
	defer s.rmListener(lis)
	for {
		conn, err := lis.Accept()
		if err != nil {
			// log.Printf("Server accept err: %v", err)
			if s.isShuttingDown() {
				return ErrServerClosed
			}
			return fmt.Errorf("listener.Accept(): %w", err)
		}
		// log.Println("Accept: ", conn.LocalAddr())

		ep := NewEndpoint(conn, WithService(s.services...))

		s.clientsMux.Lock()
		s.clients[ep] = struct{}{}
		s.clientsMux.Unlock()

		s.clientsWg.Add(1)
		go func() {
			defer s.clientsWg.Done()

			<-ep.Ctx.Done()
			// log.Printf("endpoint ended with: %v", err)
			// not sure what to do about errors (serve loop of http.Server doesn't seem to care, so we will follow suit for now)
			s.clientsMux.Lock()
			defer s.clientsMux.Unlock()
			delete(s.clients, ep)
		}()
	}
}

// Close immediately closes all listeners and all connections/endpoints
// Close returns any errors returned by the underlying listeners
func (s *Server) Close() error {
	s.inShutdown.Store(true)
	var multiError error

	// close all listeners
	s.listenersMux.Lock()
	for l := range s.listeners {
		lErr := l.Close()
		multiError = errors.Join(multiError, lErr)
	}
	s.listenersMux.Unlock()

	s.clientsMux.Lock()
	for c := range s.clients {
		//c.Close()
		if err := c.Close(); err != nil {
			log.Fatalf("c.Close(): %+v", err)
		}
	}
	s.clientsMux.Unlock()

	s.listenersWg.Wait()
	s.clientsWg.Wait()

	return multiError
}
