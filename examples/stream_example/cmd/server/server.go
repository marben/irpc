package main

import (
	"log"
	"net"

	"github.com/marben/irpc"
	"github.com/marben/irpc/examples/stream_example"
)

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	irpcServer := irpc.NewServer(irpc.WithServices(
		stream_example.NewStreamingSerIrpcService(StreamingImpl{}),
		stream_example.NewStreamingServiceIrpcService(StreamImpl{}),
	))

	tcpListener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("error opening tcp listener: %v", err)
	}

	log.Printf("listening on port 8080")

	if err := irpcServer.Serve(tcpListener); err != nil {
		log.Println("irpcServer.Serve():", err)
	}
}
