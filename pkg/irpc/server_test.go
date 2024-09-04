package irpc

import (
	"net"
	"testing"
	"time"
)

// creates a server and connects a client. checks that the client is removed from
// internal structures after close
func TestClientRemovalOnClose(t *testing.T) {
	// SERVER
	service := newTestIRpcService(testServiceImpl{skew: 0})
	t.Logf("starting server")
	server := NewServer(service)

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

	cEp := NewEndpoint(conn)
	t.Logf("registering client")
	client, err := NewMathIrpcClient(cEp)
	if err != nil {
		t.Fatalf("New client: %+v", err)
	}
	t.Log("making sure the client works")
	if res := client.Add(1, 2); res != 1+2 {
		t.Fatalf("unexpected result: %d", res)
	}
	// check the client is added to server's list
	server.clientsMux.Lock()
	if l := len(server.clients); l != 1 {
		t.Fatalf("wrong number of registered clients: %d", l)
	}
	server.clientsMux.Unlock()

	// close the endpoint. this should remove it from the list
	if err := cEp.Close(); err != nil {
		t.Fatalf("clientEndpoint.Close(): %+v", err)
	}

	time.Sleep(10 * time.Millisecond)
	server.clientsMux.Lock()
	if l := len(server.clients); l != 0 {
		t.Fatalf("wrong number of registered clients: %d", l)
	}
	server.clientsMux.Unlock()

	t.Log("closing server")
	if err := server.Close(); err != nil {
		t.Fatalf("server.Close(): %+v", err)
	}

	if err := <-serveErrC; err != ErrServerClosed {
		t.Fatalf("server returned: %+v", err)
	}
}
