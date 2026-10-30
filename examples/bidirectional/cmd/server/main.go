package main

import (
	"crypto/sha256"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/coder/websocket"
	"github.com/marben/irpc"
	"github.com/marben/irpc/examples/bidirectional"
)

// BarImpl is the concrete implementation of BarService.
type BarImpl struct{}

func (s *BarImpl) GetTime() (time.Time, error) {
	now := time.Now()
	barLog.Printf("client->server GetTime() => %s", now.Format(time.RFC3339))
	return now, nil
}

func (s *BarImpl) HashBytes(data []byte) ([]byte, error) {
	h := sha256.Sum256(data)
	barLog.Printf("client->server HashBytes(%d bytes) => %x…", len(data), h[:4])
	return h[:], nil
}

func (s *BarImpl) GetExampleData() bidirectional.ExampleData {
	data := randomExampleData()
	barLog.Printf("client->server GetExampleData() => %+v", data)
	return data
}

// BoltRealImplementation is the concrete implementation of BoltService.
type BoltRealImplementation struct{}

func (f *BoltRealImplementation) IsEven(number int) bool {
	even := number%2 == 0
	boltLog.Printf("client->server IsEven(%d) => %v", number, even)
	return even
}

var (
	barLog  = bidirectional.NewServiceLogger("Bar")
	boltLog = bidirectional.NewServiceLogger("Bolt")
	fooLog  = bidirectional.NewServiceLogger("Foo")
	fizzLog = bidirectional.NewServiceLogger("Fizz")
)

var barIrpcService = bidirectional.NewBarServiceIrpcService(&BarImpl{})
var boltIrpcService = bidirectional.NewBoltServiceIrpcService(&BoltRealImplementation{})
var contextIrpcService = bidirectional.NewContextCancellationServiceIrpcService(&bidirectional.ContextCancellationImpl{Direction: "client->server"})

func randomExampleData() bidirectional.ExampleData {
	return bidirectional.ExampleData{
		Name:    gofakeit.FirstName(),
		Surname: gofakeit.LastName(),
		Age:     gofakeit.Age(),
	}
}

func FooClientHandler(ep *irpc.Endpoint, wg *sync.WaitGroup) {
	defer wg.Done()

	checkDataTicker := time.NewTicker(3 * time.Second)
	defer checkDataTicker.Stop()

	pingTicker := time.NewTicker(1 * time.Second)
	defer pingTicker.Stop()

	foo, err := bidirectional.NewFooServiceIrpcClient(ep)
	if err != nil {
		fooLog.Printf("create FooService client: %v", err)
		return
	}

	for {
		select {
		case <-checkDataTicker.C:
			data := randomExampleData()
			err := foo.CheckDataValid(data)
			fooLog.Printf("server->client CheckDataValid(%+v) => %v", data, err)

		case <-pingTicker.C:
			msg, err := foo.Ping()
			fooLog.Printf("server->client Ping() => %q, %v", msg, err)

		case <-ep.Context().Done():
			return
		}
	}
}

func FizzClientHandler(ep *irpc.Endpoint, wg *sync.WaitGroup) {
	defer wg.Done()

	isOddTicker := time.NewTicker(1 * time.Second)
	defer isOddTicker.Stop()

	fizz, err := bidirectional.NewFizzServiceIrpcClient(ep)
	if err != nil {
		fizzLog.Printf("create FizzService client: %v", err)
		return
	}

	for {
		select {
		case <-isOddTicker.C:
			n := int(time.Now().UnixMicro())
			fizzLog.Printf("server->client IsOdd(%d) => %v", n, fizz.IsOdd(n))

		case <-ep.Context().Done():
			return
		}
	}
}

func ContextClientHandler(ep *irpc.Endpoint, wg *sync.WaitGroup) {
	defer wg.Done()

	ctxService, err := bidirectional.NewContextCancellationServiceIrpcClient(ep)
	if err != nil {
		log.Printf("create ContextCancellationService client: %v", err)
		return
	}

	bidirectional.CallLongWorkPeriodically(ep.Context(), ctxService, 8*time.Second, "server->client")
}

func handleConn(w http.ResponseWriter, r *http.Request) {
	wsConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// Accept connections from any origin. Fine for an example, don't do this in production.
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("websocket accept: %v", err)
		return
	}

	netConn := websocket.NetConn(r.Context(), wsConn, websocket.MessageBinary)
	// Services are registered at creation, so they are available before the client's first call.
	ep := irpc.NewEndpoint(netConn,
		irpc.WithEndpointServices(barIrpcService, boltIrpcService, contextIrpcService),
	)

	log.Printf("client connected: %s", r.RemoteAddr)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go FooClientHandler(ep, &wg)

	wg.Add(1)
	go FizzClientHandler(ep, &wg)

	wg.Add(1)
	go ContextClientHandler(ep, &wg)

	<-ep.Context().Done()
	wg.Wait()

	log.Printf("client disconnected: %s", r.RemoteAddr)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix(bidirectional.LogPrefix("Connection"))

	http.HandleFunc("/rpc", handleConn)

	log.Println("listening on ws://localhost:8080/rpc")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
