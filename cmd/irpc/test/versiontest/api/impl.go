package api

import (
	"fmt"
	"log"
	"net"

	"github.com/marben/irpc"
)

var _ Api = ApiImpl{}

type ApiImpl struct{}

var version = "latest"

// ApiVersion implements Api.
func (a ApiImpl) ApiVersion() (string, error) {
	return version, nil
}

func ConnectToLatestApiAndMakeCall(serviceAddr string) error {
	tcpConn, err := net.Dial("tcp", serviceAddr)
	if err != nil {
		log.Fatalf("failed to connect to %q: %v", serviceAddr, err)
	}
	defer tcpConn.Close()

	ep := irpc.NewEndpoint(tcpConn)

	client, err := NewApiIrpcClient(ep)
	if err != nil {
		return fmt.Errorf("failed to connect client: %v", err)
	}

	v, err := client.ApiVersion()
	if err != nil {
		return fmt.Errorf("failed to get version error: %w", err)
	}

	if v != version {
		return fmt.Errorf("unexpected version %q. expected %q", v, version)
	}
	log.Printf("api version: %v", v)

	return nil
}
