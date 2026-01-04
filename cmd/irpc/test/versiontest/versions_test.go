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
	defer func() {
		t.Logf("closing server connection")
		if err := tcpListener.Close(); err != nil {
			t.Fatalf("failed to close tcp listener: %v", err)
		}
	}()

	addr := tcpListener.Addr().String()
	t.Logf("listening on addr: %v", addr)

	irpcServer := irpc.NewServer(irpc.WithServices(api.NewApiIrpcService(api.ApiImpl{})))
	go func() {
		if err := irpcServer.Serve(tcpListener); err != nil {
			t.Logf("irpcServer.Serve(): %v", err)
		}
	}()

	// Client
	if err := api.ConnectToLatestApiAndMakeCall(addr); err != nil {
		t.Fatalf("connect: %+v", err)
	}
}
