package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"

	"github.com/marben/irpc"
	"github.com/marben/irpc/examples/stream_example"
)

const serviceAddr = ":8080"

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	// Ctrl+C cancels appCtx, and with it the streams still open
	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log.Println("connecting to", serviceAddr)
	tcpConn, err := net.Dial("tcp", serviceAddr)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer tcpConn.Close()

	ep := irpc.NewEndpoint(tcpConn)

	// plain RPC
	client, err := stream_example.NewStreamingSerIrpcClient(ep)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	if err := client.Example(42); err != nil {
		log.Fatalf("Example(): %v", err)
	}
	log.Printf("Example(42) => ok")

	// streams: the two examples run at the same time, on the same connection
	streams, err := stream_example.NewStreamingServiceIrpcClient(ep)
	if err != nil {
		log.Fatalf("failed to create streaming client: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		StreamInExample(appCtx, streams)
	}()
	go func() {
		defer wg.Done()
		SamplesExample(appCtx, streams)
	}()

	// the client doesn't serve anything: when the streams are done we close the connection and exit
	wg.Wait()
	log.Printf("done")
}
