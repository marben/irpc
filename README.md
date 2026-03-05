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
- High performance (comparable to or faster than gRPC in internal benchmarks)
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


## Example 1: KV Store Using `net.Pipe`
A complete runnable example is available in [examples/simple_kv_store/](examples/simple_kv_store/).

`kv.go` (API definition):
```go
type KVStore interface {
	Put(key string, value []byte, ttl time.Duration) error
	Get(key string) ([]byte, error)
	Delete(key string) error
	ModifiedSince(since time.Time) ([]string, error)
}
```

The full `main.go` walkthrough is in the example README: [examples/simple_kv_store/README.md](examples/simple_kv_store/README.md).

## Example 2: Distributed Mandelbrot set rendering done by cli and web clients
This comprehensive example using `irpc` over `tcp` and `websocket` can be found at [github.com/marben/irpc_dist_mandel](https://github.com/marben/irpc_dist_mandel)

## Bidirectional RPC

Both sides of a connection can:
- register services
- call methods on the other side

This supports patterns such as distributed workers, callbacks, and push-style interfaces.

## Context Cancellation

iRPC supports `context.Context` as the first method parameter. It listens for `Done()` on the caller side and propagates cancellation to the corresponding peer-side context.  
This does not abort the RPC call itself: the client still waits for the remote handler to finish and then returns the produced results.

## Errors and Interfaces

For `error` values, iRPC materializes a generated struct that implements `error` (`interface { Error() string }`).  
On send, it calls `Error()` on the original value and serializes the returned string. On receive, it reconstructs the generated error implementation so `Error()` returns the same string.

The same pattern is used for other simple interfaces: iRPC serializes the interface method outputs needed by the contract and recreates a generated implementation on the other side.

For transport/network failures during an RPC call, iRPC injects its own error into the call result when the method returns an `error`.  
If a method does not return an `error`, generated client code cannot surface the failure through return values, so it panics instead.

## Versioning Strategy

Each generated `_irpc.go` contains a hash of the exact generated code, and that hash is used as the service ID.  
Because of that, any API change produces a new ID (a new version).

To keep backward compatibility, keep the previously generated `_irpc.go` (for example in a `v1/` package), generate a new one for the updated API, and register both services.  
An endpoint can register multiple service IDs, so old and new API versions can coexist for as long as needed.

Example layout:
```text
api/
  v1/
    kv.go
    kv_irpc.go
  v2/
    kv.go
    kv_irpc.go
```

Typical flow:
- Keep `api/v1/kv_irpc.go` unchanged for existing clients.
- Introduce API changes in `api/v2/kv.go` and generate `api/v2/kv_irpc.go`.
- Register both `v1` and `v2` services on the endpoint during migration.

## Roadmap

The project is functional but still requires API finalization.

## Contributing
Issues and pull requests are welcome.
