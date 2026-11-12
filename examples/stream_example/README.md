# IRPC streaming example

Channel streams on top of the existing irpc primitives, without any protocol change.
`stream.go` declares an interface with channels, and `go generate` produces `stream_irpc.go`.

## How it is generated

For each method with channels, the generator derives some "wire" methods without channels and generates
them like any other method (req/resp structs, dispatch, client calls):

| Wire method | When | Role |
|---|---|---|
| `streamOpen_M` | always | registers the session on the service, before anything else |
| `stream_M` | always | main call with the scalar parameters, lives as long as the session (keeps the method's funcId) |
| `streamPush_M_<ch>` / `streamClose_M_<ch>` | for each channel parameter | sends a batch of values / the caller closed the channel |
| `streamReady_M` | if the method returns channels | waits for the implementation, returns its scalar results |
| `streamPull_M_<ch>` | for each returned channel | long poll returning a batch of values |

On top of them it generates:
- a `streams` struct for the service side, with the sessions;
- the client method implementing the user's interface;
- `IsUnlimitedFunc`.

The generic pieces (sessions, pumps, batching) are in `irpcgen/stream.go`.

## Rules

- Only `<-chan T` is supported: as a parameter (client → server) and as a return value (server → client).
- A method with streams must take a `context.Context`.
- One goroutine per channel moves the values using normal RPC calls, in batches of up to 64 values.
  The next batch is sent only after the previous one is acknowledged.
- Each call has a session. Its ctx is the one passed to the implementation, and **the reader ends it**:
  - server reading (channel parameters): the implementation is blocking, and the session ends when it returns;
  - client reading (returned channels): the session ends when the server closes the channel, or when the client cancels the ctx passed to the call.
- Sessions belong to the connection that opened them: another connection can't use them, even knowing the id,
  and they are dropped when the connection closes.
- A connection can keep at most `irpc.DefaultMaxStreamSessions` (64) sessions open at the same time, configurable
  per endpoint with `irpc.WithMaxStreamSessions`. Further calls fail with `irpcgen.ErrStreamTooManySessions`.
- The service runs the main call of a session once (`irpcgen.ErrStreamSessionInUse`), and rejects overlapping
  calls on the same channel (`irpcgen.ErrStreamBusy`). The generated client never does either: they protect the
  server from a misbehaving peer, which would otherwise run the implementation twice or pile up goroutines
  on calls that don't take a worker slot.

## Limits

- **Read a returned channel until it is closed, or cancel the ctx.** The generated client writes the received
  values into the channel: if nobody reads it, that goroutine stays blocked and the session stays open until
  the connection closes. With the default limit, 64 forgotten streams block every further stream call on the connection
  (`irpcgen.ErrStreamTooManySessions`). Use `defer cancel()` on the ctx passed to the call.
- **Throughput depends on latency.** Push and Pull are synchronous: a channel moves at most 64 values per round trip.
  On a local network this is not noticeable, on a 100ms link it is about 640 values/s per channel.

## Parallel workers

By default every incoming RPC holds one of the endpoint's worker slots (`WithParallelWorkers`), and every
outgoing call holds one of the parallel calls slots (`WithParallelClientCalls`).
A stream's calls can last as long as the stream (main call, Push with backpressure, Pull as a long poll)
and they depend on each other: with the limits, once the main calls take all the slots, their pushes and
pulls never run, and you get a deadlock.

That is why the runtime has two optional interfaces, used by the generated code:

- `irpcgen.UnlimitedFuncService`: the service declares which functions don't take a worker slot (`IsUnlimitedFunc`).
  The server decides by itself, a client cannot force it.
- `irpcgen.UnlimitedFuncCaller`: `Endpoint.CallRemoteFuncUnlimited` is not counted against `parallelClientCalls`.

`TestManyStreamsWithMinimalLimits` runs 20 streams in parallel with 1 worker and 1 client call.

## Generator errors

The generator rejects with an explicit message: `chan<- T`, `chan T`, a missing `context.Context`, a missing final `error`,
channels of channels, and channels that are not direct parameters/results (e.g. `[]<-chan T`).

## Running

```bash
go generate ./stream.go  # regenerates stream_irpc.go
go test -race .          # end-to-end over net.Pipe, with goleak
go run ./cmd/server      # terminal 1
go run ./cmd/client      # terminal 2
```

The client (`cmd/client/streams.go`) runs two streams at the same time on the same connection
(the logs have millisecond timestamps):

- `StreamInExample` (client → server): two goroutines feed `inA` (a value every 100ms) and `inB` (every 150ms),
  and close them after 3s and 4s. The server logs every value as it arrives, and `ExampleStreamIn`
  returns once both channels are closed, after ~4s.
- `SamplesExample` (server → client): `Samples(25000)` produces a sample every 200ms, more than the client wants.
  The client reads for 3s, then cancels the ctx: the server sees the cancellation and stops producing.

The client doesn't serve anything: when both are done it closes the connection and exits.
Ctrl+C cancels the streams still open, and the server sees the cancellation on both.

```
14:50:34.933840 ExampleStreamIn(): streaming to the server (inA for 3s, inB for 4s)...
14:50:34.934205 sample  0:  0.000  (after 0s)
14:50:35.134288 sample  1:  0.100  (after 201ms)
...
14:50:37.742725 sample 14:  0.985  (after 2.809s)
14:50:37.934461 Samples(25000) => received 15 samples in 3.001s, then we closed the stream: context deadline exceeded
14:50:38.934553 ExampleStreamIn() => ok, the stream was open for 4.001s
14:50:38.934594 done
```

On the server side:

```
14:50:34.934209 server: incoming number from A: 0
14:50:34.934238 server: incoming number from B: 0
14:50:35.035051 server: incoming number from A: 1
14:50:35.084498 server: incoming number from B: 100
...
14:50:37.942759 server: Samples stopped after 15 samples: context deadline exceeded
...
14:50:38.834746 server: incoming number from B: 2600
14:50:38.934434 server: ExampleStreamIn done, both channels closed
```

With Ctrl+C after 1.5s:

```
14:50:08.826506 ExampleStreamIn() => interrupt signal received, after 1.5s
14:50:08.826542 Samples(25000) => received 8 samples in 1.5s, then we closed the stream: context canceled
14:50:08.826550 done
```

