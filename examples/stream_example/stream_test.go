package stream_example

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/marben/irpc"
	"go.uber.org/goleak"
)

type testImpl struct {
	mu   sync.Mutex
	gotA []int16
	gotB []int

	samplesStopped chan error // receives ctx's cause when the Samples producer stops early
}

func (t *testImpl) ExampleStreamIn(ctx context.Context, inA <-chan int16, inB <-chan int) error {
	for inA != nil || inB != nil {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case v, ok := <-inA:
			if !ok {
				inA = nil
				continue
			}
			t.mu.Lock()
			t.gotA = append(t.gotA, v)
			t.mu.Unlock()
		case v, ok := <-inB:
			if !ok {
				inB = nil
				continue
			}
			t.mu.Lock()
			t.gotB = append(t.gotB, v)
			t.mu.Unlock()
		}
	}
	return nil
}

var errNegativeCount = errors.New("negative count")

func (t *testImpl) Samples(ctx context.Context, count int) (<-chan float64, error) {
	if count < 0 {
		return nil, errNegativeCount
	}
	out := make(chan float64)
	go func() {
		defer close(out)
		for i := range count {
			select {
			case out <- float64(i) / 2:
			case <-ctx.Done():
				if t.samplesStopped != nil {
					t.samplesStopped <- context.Cause(ctx)
				}
				return
			}
		}
	}()
	return out, nil
}

func setup(t *testing.T, impl *testImpl, opts ...irpc.EndpointOption) *StreamingServiceIrpcClient {
	t.Helper()

	c1, c2 := net.Pipe()
	serverEp := irpc.NewEndpoint(c1, append(opts, irpc.WithEndpointServices(NewStreamingServiceIrpcService(impl)))...)
	clientEp := irpc.NewEndpoint(c2, opts...)
	t.Cleanup(func() {
		clientEp.Close()
		serverEp.Close()
		<-clientEp.Context().Done()
		<-serverEp.Context().Done()
	})

	client, err := NewStreamingServiceIrpcClient(clientEp)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestStreamIn(t *testing.T) {
	t.Cleanup(func() { goleak.VerifyNone(t) }) // registered first, so it runs after the endpoints are closed
	impl := &testImpl{}
	client := setup(t, impl)

	const n = 1000
	inA := make(chan int16)
	inB := make(chan int)
	go func() {
		defer close(inA)
		for i := range n {
			inA <- int16(i)
		}
	}()
	go func() {
		defer close(inB)
		for i := range n {
			inB <- -i
		}
	}()

	if err := client.ExampleStreamIn(t.Context(), inA, inB); err != nil {
		t.Fatalf("ExampleStreamIn: %v", err)
	}

	if len(impl.gotA) != n || len(impl.gotB) != n {
		t.Fatalf("got %d/%d values, expected %d", len(impl.gotA), len(impl.gotB), n)
	}
	for i := range n {
		if impl.gotA[i] != int16(i) || impl.gotB[i] != -i {
			t.Fatalf("wrong value/order at %d: %d %d", i, impl.gotA[i], impl.gotB[i])
		}
	}
}

func TestSamples(t *testing.T) {
	t.Cleanup(func() { goleak.VerifyNone(t) }) // registered first, so it runs after the endpoints are closed
	client := setup(t, &testImpl{})

	const n = 1000
	ch, err := client.Samples(t.Context(), n)
	if err != nil {
		t.Fatal(err)
	}
	i := 0
	for v := range ch {
		if v != float64(i)/2 {
			t.Fatalf("sample %d: got %v", i, v)
		}
		i++
	}
	if i != n {
		t.Fatalf("got %d samples, expected %d", i, n)
	}
}

func TestSamplesImplError(t *testing.T) {
	t.Cleanup(func() { goleak.VerifyNone(t) }) // registered first, so it runs after the endpoints are closed
	client := setup(t, &testImpl{})

	ch, err := client.Samples(t.Context(), -1)
	if err == nil || err.Error() != errNegativeCount.Error() {
		t.Fatalf("expected %q, got %v", errNegativeCount, err)
	}
	if ch != nil {
		t.Fatal("expected nil channel")
	}
}

// the reader stops early by cancelling its ctx: the producer on the server must see its ctx done
func TestSamplesReaderCancels(t *testing.T) {
	t.Cleanup(func() { goleak.VerifyNone(t) }) // registered first, so it runs after the endpoints are closed
	impl := &testImpl{samplesStopped: make(chan error, 1)}
	client := setup(t, impl)

	ctx, cancel := context.WithCancel(t.Context())
	ch, err := client.Samples(ctx, 1_000_000)
	if err != nil {
		t.Fatal(err)
	}
	for range 10 {
		<-ch
	}
	cancel()

	// the client channel gets closed
	for range ch {
	}

	select {
	case <-impl.samplesStopped:
	case <-time.After(5 * time.Second):
		t.Fatal("server producer didn't stop")
	}
}

// with a single worker and a single client call, many streams run at the same time:
// the stream calls don't take slots, so they can't deadlock each other
func TestManyStreamsWithMinimalLimits(t *testing.T) {
	t.Cleanup(func() { goleak.VerifyNone(t) })
	impl := &testImpl{}
	client := setup(t, impl, irpc.WithParallelWorkers(1), irpc.WithParallelClientCalls(1))

	const streams = 10
	const n = 100
	var wg sync.WaitGroup
	errC := make(chan error, 2*streams)
	for range streams {
		wg.Add(2)
		go func() {
			defer wg.Done()
			inA := make(chan int16)
			inB := make(chan int)
			go func() {
				defer close(inA)
				defer close(inB)
				for i := range n {
					inA <- int16(i)
					inB <- i
				}
			}()
			errC <- client.ExampleStreamIn(t.Context(), inA, inB)
		}()
		go func() {
			defer wg.Done()
			ch, err := client.Samples(t.Context(), n)
			if err != nil {
				errC <- err
				return
			}
			got := 0
			for range ch {
				got++
			}
			if got != n {
				errC <- fmt.Errorf("got %d samples, expected %d", got, n)
				return
			}
			errC <- nil
		}()
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("deadlock: streams didn't complete")
	}
	close(errC)
	for err := range errC {
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(impl.gotA) != streams*n || len(impl.gotB) != streams*n {
		t.Fatalf("server received %d/%d values, expected %d", len(impl.gotA), len(impl.gotB), streams*n)
	}
}
