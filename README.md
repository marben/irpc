# iRPC
## Interface-Driven RPC Code Generator for Go

`irpc` is a light RPC code generator for Go.

From a Go interface, `irpc` generates:
- A client implementation that forwards calls over a connection
- A server adapter that dispatches requests to your implementation
- Binary serialization code for parameters and return values

The project focuses on type safety and readable generated code, without reflection or schema files.


## Features

- Generates RPC clients and server adapters from standard Go interfaces
- Supports bidirectional RPC on the same connection
- Works with any `io.ReadWriteCloser` implementation (for example TCP, pipes, or WebSockets)
- Does not rely on reflection
- Supports common Go types:
  - primitives
  - structs
  - slices and maps
  - `time.Time` or any type implementing `encoding.BinaryMarshaler`
  - `context.Context`
  - `error` and other simple interfaces
  - pointers


## Installation

Add the runtime dependency to your project:
```bash
go get github.com/marben/irpc
```

Install the `irpc` binary into your Go bin path (for example `$HOME/go/bin`):
```bash
go install github.com/marben/irpc/cmd/irpc@latest
```

## Usage

Generate code from an interface file:

```bash
irpc api.go
```

Output file:

```text
api_irpc.go
```

Alternatively, use `go generate` to build and run the `irpc` binary by adding a `go:generate` directive to your `api.go` interface definition file:
```go
//go:generate go run github.com/marben/irpc/cmd/irpc@latest $GOFILE
```
Then run:
```bash
go generate api.go
```


## Example: KV Store Using `net.Pipe`
A complete runnable example is available in [examples/simple_kv_store/](examples/simple_kv_store/).

`kv.go` (api definition):
```go
type KVStore interface {
	Put(key string, value []byte, ttl time.Duration) error
	Get(key string) ([]byte, error)
	Delete(key string) error
	ModifiedSince(since time.Time) ([]string, error)
}
```

`main.go`
```go
package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/marben/irpc"
)

func main() {
	clientConn, serverConn := net.Pipe()

	// Each side of the connection needs an Endpoint to handle RPC.
	serverEp := irpc.NewEndpoint(serverConn)
	defer serverEp.Close()
	clientEp := irpc.NewEndpoint(clientConn)
	defer clientEp.Close()

	// In-memory implementation of the KV store.
	kvStore := newKVMemory()

	// Wrap our implementation in a generated service adapter.
	service := NewKVStoreIrpcService(kvStore)

	// Registering service on server endpoint makes it available to the client endpoint
	// After this, the client endpoint can call KVStore methods remotely.
	serverEp.RegisterService(service)

	// Create the generated client-side proxy.
	// Client implements KVStore interface, so it can be passed around as such
	client, err := NewKVStoreIrpcClient(clientEp)
	if err != nil {
		log.Fatalf("NewKVStoreIrpcClient: %v", err)
	}

	fmt.Println("Putting key 'hello'")
	if err := client.Put("hello", []byte("world"), 2*time.Minute); err != nil {
		log.Fatalf("client.Put: %v", err)
	}

	value, err := client.Get("hello")
	if err != nil {
		log.Fatalf("client.Get: %v", err)
	}
	fmt.Println("Value: ", string(value))

	oneMinuteAgo := time.Now().Add(-1 * time.Minute)

	// Since we just wrote the key, this will always return ["hello"].
	keys, err := client.ModifiedSince(oneMinuteAgo)
	if err != nil {
		log.Fatalf("client.ModifiedSince: %v", err)
	}
	fmt.Println("Modified keys:", keys)
}


```

## Bidirectional RPC

Both sides of a connection can:
- register services
- call methods on the other side

This supports patterns such as distributed workers, callbacks, and push-style interfaces.

## Roadmap

The project is functional but still requires API finalization and versioning decisions.

## Contributing
Issues and pull requests are welcome.
