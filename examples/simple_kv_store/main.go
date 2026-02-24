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
