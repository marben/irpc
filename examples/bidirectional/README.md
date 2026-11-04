# Bidirectional IRPC over WebSocket Example

This example demonstrates using [irpc](https://github.com/marben/irpc) over a single WebSocket connection to show:

1. **Bidirectional RPC**: both sides of the connection expose services and call each other's.
2. **Multiple services on one endpoint**: each side serves several services and calls several remote ones over the same connection.
3. **Context cancellation over RPC**: cancelling the caller's context cancels the context of the remote function.

## Services

The services are defined in `services.go`. The irpc generator creates a client proxy and a service adapter for each of them in `services_irpc.go`.

| Service                      | Served by         | Called by         |
|------------------------------|-------------------|-------------------|
| `BarService`                 | server            | client            |
| `BoltService`                | server            | client            |
| `FooService`                 | client            | server            |
| `FizzService`                | client            | server            |
| `ContextCancellationService` | client and server | client and server |

## Bidirectional RPC

In irpc there is no difference between client and server once the connection is established. Both sides wrap their `net.Conn` in an `irpc.Endpoint`, and an endpoint can serve local services and call remote ones at the same time.

In this example the client dials the server, but after that the server is just as free to call the client: `cmd/server/main.go` starts goroutines that call `FooService`, `FizzService` and `ContextCancellationService` on every connected client, while the client keeps calling `BarService`, `BoltService` and `ContextCancellationService` on the server.

The WebSocket connection is turned into a `net.Conn` with `websocket.NetConn`, so irpc doesn't need to know anything about WebSockets.

## Multiple Services on One Endpoint

Services are registered on the endpoint when it is created. `irpc.WithEndpointServices` accepts any number of services:

```go
ep := irpc.NewEndpoint(netConn,
	irpc.WithEndpointServices(barIrpcService, boltIrpcService, contextIrpcService),
)
```

Services can also be added later with `ep.RegisterService(...)`. Registering them at creation guarantees they are available before the remote side makes its first call.

To call remote services, create one client proxy per service on the same endpoint:

```go
bar, err := bidirectional.NewBarServiceIrpcClient(ep)
bolt, err := bidirectional.NewBoltServiceIrpcClient(ep)
```

All the calls share the same connection; irpc routes each one to the right service.

## Context Cancellation over RPC

`ContextCancellationService.LongWorkThatWillBeCancelled` takes a `context.Context` and needs `LongWorkDelay` (12s) to complete, unless its context is cancelled first.

The service is served and called by both sides, so the cancellation is shown in both directions. Both use the same implementation, `ContextCancellationImpl`, and the same calling code, `CallLongWorkPeriodically` in `services.go`, which every 8 seconds calls the remote side with a context that expires 2 seconds before the work is done:

```go
callCtx, cancel := context.WithTimeout(ctx, LongWorkDelay-2*time.Second)
err := svc.LongWorkThatWillBeCancelled(callCtx)
cancel()
```

When the caller's context expires, irpc notifies the remote side, and the `ctx` received by the remote implementation is cancelled too. The remote function stops right away and the call returns an error on both sides. For a call from the server to the client:

```
server: [ContextCancellation] server->client LongWorkThatWillBeCancelled() => context canceled
client: [ContextCancellation] server->client LongWorkThatWillBeCancelled() cancelled by the caller: context canceled
```

and the same happens, with `client->server`, for the calls from the client to the server.

The call blocks for several seconds, so the client makes it from a separate goroutine to not hold up its other calls. irpc handles concurrent calls on the same endpoint.

Note that the remote side sees a plain cancellation (`context.Canceled`), not the reason why the caller's context ended.

## Running the Example

Start the server in one terminal:
```bash
go run ./cmd/server
```

Run the client in another terminal:
```bash
go run ./cmd/client
```

The client connects to the server on `ws://localhost:8080/rpc`, and both sides start calling each other periodically.

Every log line is prefixed with the service group it belongs to (`[Bar]`, `[Foo]`, `[ContextCancellation]`, ...), and calls are marked with their direction: `client->server` for calls made by the client, `server->client` for calls made by the server. The same call shows up with the same text in both logs.

## Module and Code Generation

This example has its own `go.mod`, so its dependencies (`github.com/coder/websocket`, `github.com/brianvoe/gofakeit/v7`) don't end up in the irpc module. A `replace` directive points it at the irpc sources in this repository.

The RPC code in `services_irpc.go` is generated by running `go generate` in this directory. The irpc generator is declared as a `tool` in `go.mod`, so it is invoked with `go tool irpc`.
