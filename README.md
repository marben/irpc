# iRPC - Interface-based RPC Code Generator for Go

`irpc` is a small, dependency-free RPC generator for Go.  
You write a Go interface, and `irpc` generates:

- A **client** that implements the same interface and forwards calls over a connection
- A **server adapter** that dispatches incoming requests to your implementation
- **Binary serialization** code for method parameters and return values

iRPC is built around **simplicity**, **type safety**, and **readable generated code**.  
No reflection. No schema files.


---

## ✨ Features

- Generate RPC clients and servers directly from plain Go interfaces
- Duplex (bidirectional) RPC — both sides can call each other
- Works with any `io.ReadWriteCloser` (TCP, pipes, WebSockets ...)
- No reflection, no hidden magic
- Supports:
  - primitives
  - structs
  - slices, maps
  - `time.Time`
  - context.Context
  - some interfaces (`error`)

---

## 🚀 Getting Started

Install:

```bash
go install github.com/marben/irpc/cmd/irpc@latest
```

Generate code:
``` bash
irpc api.go
```

Generated file:
```
api_irpc.go
```


## 📘 Example: KV Store Using net.Pipe

Here is a minimal, self-contained example demonstrating:
- how to create endpoints
- how to register a service
- how to call generated client methods
- how RPC works over an in-memory pipe (no TCP needed)

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

// --------------------------------------------------------------
// Example usage of the generated RPC client & service
// --------------------------------------------------------------
func main() {
	// For a self-contained example, we use net.Pipe instead
	// of actual TCP: no ports, no concurrency issues.
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

## 🔁 Bidirectional RPC
Both sides of a connection can:
- register services
- call each other’s methods

This allows patterns like:
- distributed workers
- server-initiated callbacks
- remote control interfaces
- push notifications
- multi-node computation

## 📅 Roadmap
- Project is in failrly well working state, but needs api finalization and version definition

## 🤝 Contributing
PRs and issues are welcome!
If you build something cool with IRPC, please share it.
