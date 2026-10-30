package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/coder/websocket"
	"github.com/marben/irpc"
	"github.com/marben/irpc/examples/bidirectional"
)

// fooRealImplementation is the concrete implementation of FooService.
// The server calls Ping() and CheckDataValid() on us autonomously.
type fooRealImplementation struct{}

func (c *fooRealImplementation) Ping() (string, error) {
	msg := fmt.Sprintf("pong from client at %s", time.Now().Format(time.Kitchen))
	fooLog.Printf("server->client Ping() => %q", msg)
	return msg, nil
}

func (c *fooRealImplementation) CheckDataValid(data bidirectional.ExampleData) error {
	// Always Okay
	fooLog.Printf("server->client CheckDataValid(%+v) => <nil>", data)
	return nil
}

// fizzRealImplementation is the concrete implementation of FizzService.
type fizzRealImplementation struct{}

func (f *fizzRealImplementation) IsOdd(number int) bool {
	odd := number%2 != 0
	fizzLog.Printf("server->client IsOdd(%d) => %v", number, odd)
	return odd
}

var (
	barLog  = bidirectional.NewServiceLogger("Bar")
	boltLog = bidirectional.NewServiceLogger("Bolt")
	fooLog  = bidirectional.NewServiceLogger("Foo")
	fizzLog = bidirectional.NewServiceLogger("Fizz")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix(bidirectional.LogPrefix("Connection"))

	serverAddr := "ws://localhost:8080/rpc"
	log.Printf("connecting to %s", serverAddr)

	wsConn, _, err := websocket.Dial(context.Background(), serverAddr, nil)
	if err != nil {
		log.Fatalf("websocket.Dial: %v", err)
	}
	defer wsConn.CloseNow()

	netConn := websocket.NetConn(context.Background(), wsConn, websocket.MessageBinary)

	// The client endpoint serves these services so the server can call back.
	irpcFooService := bidirectional.NewFooServiceIrpcService(&fooRealImplementation{})
	irpcFizzService := bidirectional.NewFizzServiceIrpcService(&fizzRealImplementation{})
	irpcContextService := bidirectional.NewContextCancellationServiceIrpcService(&bidirectional.ContextCancellationImpl{Direction: "server->client"})

	ep := irpc.NewEndpoint(netConn,
		irpc.WithEndpointServices(irpcFooService, irpcFizzService, irpcContextService),
	)

	// Create proxies for the services exposed by the server.
	bar, err := bidirectional.NewBarServiceIrpcClient(ep)
	if err != nil {
		log.Fatalf("NewBarServiceIrpcClient: %v", err)
	}

	bolt, err := bidirectional.NewBoltServiceIrpcClient(ep)
	if err != nil {
		log.Fatalf("NewBoltServiceIrpcClient: %v", err)
	}

	ctxService, err := bidirectional.NewContextCancellationServiceIrpcClient(ep)
	if err != nil {
		log.Fatalf("NewContextCancellationServiceIrpcClient: %v", err)
	}

	log.Println("connected. starting autonomous calls (Ctrl-C to stop)")

	// LongWorkThatWillBeCancelled blocks for several seconds, so it gets its own goroutine to not hold up the other calls.
	go bidirectional.CallLongWorkPeriodically(ep.Context(), ctxService, 8*time.Second, "client->server")

	getTimeTicker := time.NewTicker(2 * time.Second)
	defer getTimeTicker.Stop()

	getHashTicker := time.NewTicker(5 * time.Second)
	defer getHashTicker.Stop()

	getDataTicker := time.NewTicker(3 * time.Second)
	defer getDataTicker.Stop()

	for {
		select {
		case <-getTimeTicker.C:
			// Ask the server for the current time.
			t, err := bar.GetTime()
			if err != nil {
				log.Fatalf("GetTime: %v", err)
			}
			barLog.Printf("client->server GetTime() => %s", t.Format(time.RFC3339))

			n := int(t.UnixMicro())
			boltLog.Printf("client->server IsEven(%d) => %v", n, bolt.IsEven(n))

		case <-getHashTicker.C:
			// Send 32 random bytes and get a SHA-256 hash back.
			buf := make([]byte, 32)
			if _, err := rand.Read(buf); err != nil {
				log.Fatalf("rand.Read: %v", err)
			}
			h, err := bar.HashBytes(buf)
			if err != nil {
				log.Fatalf("HashBytes: %v", err)
			}
			barLog.Printf("client->server HashBytes(%d bytes) => %x", len(buf), h)

		case <-getDataTicker.C:
			barLog.Printf("client->server GetExampleData() => %+v", bar.GetExampleData())

		case <-ep.Context().Done():
			log.Printf("connection closed: %v", context.Cause(ep.Context()))
			return
		}
	}
}
