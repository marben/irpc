package main

import (
	"context"
	"log"
	"time"

	"github.com/marben/irpc/examples/stream_example"
)

// How long the demo keeps each stream open, and how fast the client produces.
const (
	// client -> server
	inAFor           = 3 * time.Second // inA is closed after this
	inBFor           = 4 * time.Second // inB is closed after this
	inAEvery         = 100 * time.Millisecond
	inBEvery         = 150 * time.Millisecond
	StreamCtxTimeOut = 5 * time.Second // upper bound for the whole ExampleStreamIn call

	// server -> client
	samplesCount = 25000
	readFor      = 3 * time.Second // the client stops reading the samples after this
)

// StreamInExample: CLIENT -> SERVER, two channels in the same call.
// Each channel is fed by its own goroutine and closed after inAFor/inBFor.
// ExampleStreamIn is blocking: it returns when the server has read both channels until they are closed.
func StreamInExample(ctx context.Context, streams *stream_example.StreamingServiceIrpcClient) {
	inA := make(chan int16, 500)
	inB := make(chan int, 500)

	go produce(ctx, inA, inAFor, inAEvery, func(i int) int16 { return int16(i) })
	go produce(ctx, inB, inBFor, inBEvery, func(i int) int { return i * 100 })

	streamCtx, cancel := context.WithTimeout(ctx, StreamCtxTimeOut)
	defer cancel()

	start := time.Now()
	log.Printf("ExampleStreamIn(): streaming to the server (inA for %v, inB for %v)...", inAFor, inBFor)
	if err := streams.ExampleStreamIn(streamCtx, inA, inB); err != nil {
		log.Printf("ExampleStreamIn() => %v, after %v", err, since(start))
		return
	}
	log.Printf("ExampleStreamIn() => ok, the stream was open for %v", since(start))
}

// SamplesExample: SERVER -> CLIENT, the client reads for readFor and then cancels the ctx.
// If the server finishes first, it closes the channel and the loop ends by itself.
// Either way the channel gets closed: ranging over it is enough.
func SamplesExample(ctx context.Context, streams *stream_example.StreamingServiceIrpcClient) {
	readCtx, cancel := context.WithTimeout(ctx, readFor)
	defer cancel()

	start := time.Now()
	samples, err := streams.Samples(readCtx, samplesCount)
	if err != nil {
		log.Printf("Samples(%d) => %v", samplesCount, err)
		return
	}

	n := 0
	for s := range samples {
		log.Printf("sample %2d: % .3f  (after %v)", n, s, since(start))
		n++
	}

	if err := readCtx.Err(); err != nil {
		log.Printf("Samples(%d) => received %d samples in %v, then we closed the stream: %v", samplesCount, n, since(start), err)
	} else {
		log.Printf("Samples(%d) => received %d samples in %v, closed by the server", samplesCount, n, since(start))
	}
}

// produce sends next(0), next(1), ... on ch, one every `every`, and closes ch after `duration`
// (or earlier, if ctx is cancelled). Closing the channel is how the client tells the server "no more values".
func produce[T any](ctx context.Context, ch chan<- T, duration, every time.Duration, next func(i int) T) {
	defer close(ch)

	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	tick := time.NewTicker(every)
	defer tick.Stop()

	for i := 0; ; i++ {
		select {
		case ch <- next(i):
		case <-ctx.Done():
			return
		}

		// wait on ctx too: with a buffered ch both cases above may be ready, and select would pick one at random
		select {
		case <-tick.C:
		case <-ctx.Done():
			return
		}
	}
}

func since(t time.Time) time.Duration {
	return time.Since(t).Round(time.Millisecond)
}
