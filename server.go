package irpc

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"github.com/marben/irpc/irpcgen"
)

var ErrServerClosed error = errors.New("irpc: server closed")

type Server struct {
	services []irpcgen.Service // don't mutate after Serve()

	clients    map[*Endpoint]struct{}
	clientsMux sync.Mutex
	clientsWg  sync.WaitGroup

	onConnect func(*Endpoint)

	inShutdown atomic.Bool

	listeners    map[net.Listener]struct{} // todo: should we store pointers in a similar fashion std http server does?
	listenersMux sync.Mutex
	listenersWg  sync.WaitGroup
}

func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		listeners: make(map[net.Listener]struct{}),
		clients:   make(map[*Endpoint]struct{}),
	}
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Do not call after Serve
func (s *Server) AddService(svc ...irpcgen.Service) {
	s.services = append(s.services, svc...)
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
	// log.Printf("irpc server: serving %s on addr %s", lis.Addr().Network(), lis.Addr())
	if err := s.addListener(lis); err != nil {
		return err
	}
	defer s.rmListener(lis)
	for {
		conn, err := lis.Accept()
		if err != nil {
			if s.isShuttingDown() {
				return ErrServerClosed
			}
			return fmt.Errorf("listener.Accept(): %w", err)
		}
		// log.Println("Accept: ", conn.LocalAddr())

		if s.isShuttingDown() {
			conn.Close()
			return ErrServerClosed
		}

		ep := NewEndpoint(conn,
			WithEndpointServices(s.services...),
			WithLocalAddress(conn.LocalAddr()),
			WithRemoteAddress(conn.RemoteAddr()),
		)

		s.clientsMux.Lock()
		s.clients[ep] = struct{}{}
		s.clientsMux.Unlock()

		s.clientsWg.Add(1)
		go func() {
			defer s.clientsWg.Done()
			if s.onConnect != nil {
				s.onConnect(ep)
			}

			<-ep.ctx.Done()
			// not sure what to do about errors (serve loop of http.Server doesn't seem to care, so we will follow suit for now)
			s.clientsMux.Lock()
			delete(s.clients, ep)
			s.clientsMux.Unlock()

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
		if err := l.Close(); err != nil {
			multiError = errors.Join(multiError, err)
		}
	}
	s.listenersMux.Unlock()

	s.clientsMux.Lock()
	for c := range s.clients {
		if err := c.Close(); err != nil {
			multiError = errors.Join(multiError, err)
		}
	}
	s.clientsMux.Unlock()

	s.listenersWg.Wait()
	s.clientsWg.Wait()

	return multiError
}

type ServerOption func(*Server)

// OnConnect is called synchronously for each accepted connection.
func WithOnConnect(f func(ep *Endpoint)) ServerOption {
	return func(s *Server) {
		s.onConnect = f
	}
}

func WithServices(svcs ...irpcgen.Service) ServerOption {
	return func(s *Server) {
		s.services = append(s.services, svcs...)
	}
}
