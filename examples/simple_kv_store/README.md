# Simple KV Store Example

This is a simple example demonstrating the use of the [iRPC](https://github.com/marben/irpc) library to create RPC services for a basic key-value store.

## Overview

The example defines a `KVStore` interface with methods like `Put`, `Get`, `Delete`, and `ModifiedSince`. The iRPC code generator automatically creates the RPC client and server implementations from this interface.

## Why net.Pipe?

Instead of using TCP connections, this example uses `net.Pipe()` to create a self-contained demonstration. This avoids the need for network ports, making it easy to run without any setup or concurrency concerns. In a real application, you would replace `net.Pipe()` with actual TCP or other transport mechanisms. All iRPC needs is an `io.ReadWriteCloser`.

## Running the Example

Run `go run .` in this directory.

The program will simulate a client-server interaction using the in-memory KV store, demonstrating RPC calls over the pipe connection.

`main.go`:
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

	// Registering service on server endpoint makes it available to the client endpoint.
	// After this, the client endpoint can call KVStore methods remotely.
	serverEp.RegisterService(service)

	// Create the generated client-side proxy.
	// Client implements KVStore interface, so it can be passed around as such.
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
	fmt.Println("Value:", string(value))

	oneMinuteAgo := time.Now().Add(-1 * time.Minute)

	// Since we just wrote the key, this will always return ["hello"].
	keys, err := client.ModifiedSince(oneMinuteAgo)
	if err != nil {
		log.Fatalf("client.ModifiedSince: %v", err)
	}
	fmt.Println("Modified keys:", keys)
}
```

## Code Generation
API definition is in file [kv.go](kv.go).
The RPC code in [kv_irpc.go](kv_irpc.go) is generated using the `//go:generate` directive by running `go generate kv.go`. This creates the necessary serialization, client proxy, and server adapter code automatically.
