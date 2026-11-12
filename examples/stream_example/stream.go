package stream_example

import (
	"context"
)

//go:generate go run ../../cmd/irpc

// StreamingService is the interface as the user writes it: channels are allowed as
// parameters (client -> server streams) and as return values (server -> client streams).
type StreamingService interface {
	// ExampleStreamIn consumes two streams coming from the caller.
	// The implementation is blocking: it returns when it stops reading
	// (both channels closed, ctx done, or its own decision).
	ExampleStreamIn(ctx context.Context, inA <-chan int16, inB <-chan int) error

	// Samples produces a flow of count samples for the caller.
	// The implementation returns immediately and produces in its own goroutine,
	// closing the channel when done. ctx stays alive for the whole stream:
	// the producer must stop when ctx is done.
	Samples(ctx context.Context, count int) (<-chan float64, error)
}
