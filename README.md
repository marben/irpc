# iRPC
## Interface-defined RPC Code Generator for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/marben/irpc.svg)](https://pkg.go.dev/github.com/marben/irpc)

`iRPC` is a light RPC code generator for Go.

From a Go interface, `iRPC` generates:
- A client implementation that forwards calls over a connection
- A server adapter that dispatches requests to your implementation
- Binary serialization code for parameters and return values

The project focuses on type safety and readable generated code, without reflection or schema files.


## Features

- Generates RPC clients and server adapters from standard Go interfaces
- Supports bidirectional RPC on the same connection
- Supports streaming with Go channels (`<-chan T` as parameters and results)
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

The full `main.go` walkthrough is in the example README.

## Example 2: Bidirectional RPC over WebSocket
A runnable client/server example is available in [examples/bidirectional/](examples/bidirectional/).

It shows both sides of a single WebSocket connection serving and calling each other's services, multiple services registered on one endpoint, and `context.Context` cancellation propagated across the connection in both directions.

## Example 3: Streaming with channels
A runnable client/server example is available in [examples/stream_example/](examples/stream_example/).

It streams values from the client to the server on two channels of the same call, and reads a flow of samples
produced by the server, stopping it early by cancelling the `context.Context`.
Its README also explains how the generated code works.

## Example 4: Distributed Mandelbrot set rendering done by cli and web clients
This comprehensive example using `iRPC` over `TCP` and `WebSocket` can be found at [github.com/marben/irpc_dist_mandel](https://github.com/marben/irpc_dist_mandel)

<a href="https://github.com/marben/irpc_dist_mandel">
  <img src="https://raw.githubusercontent.com/marben/irpc_dist_mandel/main/webclient.png" width="400">
</a>

## Bidirectional RPC

Both sides of a connection can:
- register services
- call methods on the other side

This supports patterns such as distributed workers, callbacks, and push-style interfaces.  
See [examples/bidirectional/](examples/bidirectional/) for a working example.

## Context Cancellation

iRPC supports `context.Context` as the first method parameter. It listens for `Done()` on the caller side and propagates cancellation to the corresponding peer-side context.  
This does not abort the RPC call itself: the client still waits for the remote handler to finish and then returns the produced results.

For methods with channels (see below), cancelling the `context.Context` does end the streams on both sides.

## Streaming with Channels

A method can take channels as parameters (client → server streams) and return channels (server → client streams):

```go
type StreamingService interface {
	// the implementation reads inA and inB, and returns when it stops reading
	ExampleStreamIn(ctx context.Context, inA <-chan int16, inB <-chan int) error

	// the implementation returns immediately and produces in its own goroutine, closing the channel when done
	Samples(ctx context.Context, count int) (<-chan float64, error)
}
```

On the client side the generated method is used like a local one:

```go
samples, err := client.Samples(ctx, 100)
if err != nil {
	return err
}
for s := range samples { // closed when the server closes it, or when ctx is cancelled
	fmt.Println(s)
}
```

Rules, checked by the generator:
- only receive-only channels (`<-chan T`) are supported, directly as parameters or results;
- a method with channels must take a `context.Context` and return `error` as its last result;
- the element type can be any supported type, except channels and `context.Context`.

How it works:
- no protocol change: the generator turns each method with channels into a few ordinary RPC calls
  that open a session, push and pull batches of values (up to 64), and close it;
- the context is the lifetime of the streams: the implementation must stop when it is done,
  and the caller cancels it to stop reading early;
- sessions belong to the connection that opened them, and are dropped when it closes.
  Each endpoint accepts at most `irpc.DefaultMaxStreamSessions` (64) open sessions, configurable with `irpc.WithMaxStreamSessions`;
- the calls that can last as long as the stream don't take the endpoint's worker slots
  (`WithParallelWorkers`, `WithParallelClientCalls`), so many open streams can't deadlock the endpoint.

Limitations:
- **A returned channel must be read until it is closed, or the `context.Context` cancelled.**
  If the caller just stops reading, the generated code stays blocked writing into the channel and the session
  stays open until the connection closes. Once `WithMaxStreamSessions` sessions are stuck, further stream calls
  fail with `irpcgen.ErrStreamTooManySessions`. As with any Go API returning a channel, stop with `cancel()`:
  ```go
  ctx, cancel := context.WithCancel(ctx)
  defer cancel() // stops the stream even if we return before the channel is closed
  samples, err := client.Samples(ctx, 100)
  ```
- **Throughput depends on latency.** Each channel waits for the round trip of a batch (up to 64 values)
  before sending the next one, so a single channel moves at most 64 values per round trip:
  about 640 values/s on a 100ms link. Values are not lost or reordered, only slowed down.

See [examples/stream_example/](examples/stream_example/) for the details.

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
