package main

import (
	"context"
	"errors"
	"log"
	"math"
	"time"
)

// sampleEvery paces Samples, so the flow can be followed in the logs.
const sampleEvery = 200 * time.Millisecond

// StreamingImpl implements stream_example.StreamingSer and does the actual work.
type StreamingImpl struct{}

func (StreamingImpl) Example(n int) error {
	log.Printf("server: Example(%d)", n)
	return nil
}

// StreamImpl implements stream_example.StreamingService. It contains no network code.
type StreamImpl struct{}

// ExampleStreamIn is blocking: it returns when it stops reading.
func (StreamImpl) ExampleStreamIn(ctx context.Context, inA <-chan int16, inB <-chan int) error {
	for inA != nil || inB != nil {
		select {
		case <-ctx.Done():
			log.Printf("server: ExampleStreamIn cancelled: %v", context.Cause(ctx))
			return context.Cause(ctx)

		case num, ok := <-inA:
			if !ok {
				inA = nil // a closed channel is always ready: set it to nil to stop selecting it
				continue
			}
			log.Printf("server: incoming number from A: %d", num)

		case num, ok := <-inB:
			if !ok {
				inB = nil
				continue
			}
			log.Printf("server: incoming number from B: %d", num)
		}
	}
	log.Printf("server: ExampleStreamIn done, both channels closed")
	return nil
}

// Samples returns immediately and produces a sine wave in its own goroutine, one sample every sampleEvery,
// for as long as the caller keeps the stream open (or until count samples have been produced).
func (StreamImpl) Samples(ctx context.Context, count int) (<-chan float64, error) {
	if count < 0 {
		return nil, errors.New("count must be >= 0")
	}
	out := make(chan float64)
	go func() {
		defer close(out)
		for i := range count {
			select {
			case out <- math.Sin(float64(i) / 10):
				log.Printf("server: sent sample %d", i)
			case <-ctx.Done():
				log.Printf("server: Samples stopped after %d samples: %v", i, context.Cause(ctx))
				return
			}
			time.Sleep(sampleEvery)
		}
		log.Printf("server: Samples produced all %d samples", count)
	}()
	return out, nil
}
