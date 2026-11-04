package versiontest

import (
	"log"
	"net"
	"testing"

	"github.com/marben/irpc"
	"github.com/marben/irpc/cmd/irpc/test/versiontest/api"
)

func TestOneRegisteredVersion(t *testing.T) {
	// Server
	tcpListener, err := net.Listen("tcp", ":")
	if err != nil {
		log.Fatalf("error opening tcp listener: %v", err)
	}
	addr := tcpListener.Addr().String()
	t.Logf("listening on addr: %v", addr)

	irpcServer := irpc.NewServer(irpc.WithServices(api.NewApiIrpcService(api.ApiImpl{})))
	serveC := make(chan error, 1)
	go func() {
		serveC <- irpcServer.Serve(tcpListener)
	}()
	// close the server and wait for Serve() to return before the test ends,
	// so the serving goroutine doesn't outlive the test
	defer func() {
		if err := irpcServer.Close(); err != nil {
			t.Errorf("failed to close server: %v", err)
		}
		if err := <-serveC; err != irpc.ErrServerClosed {
			t.Errorf("unexpected Serve() error: %v", err)
		}
	}()

	// Client
	if err := api.ConnectToLatestApiAndMakeCall(addr); err != nil {
		t.Fatalf("connect: %+v", err)
	}
}
