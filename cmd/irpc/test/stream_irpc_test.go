package irpctestpkg

import (
	"context"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/marben/irpc"
	"github.com/marben/irpc/cmd/irpc/test/testtools"
	"github.com/marben/irpc/irpcgen"
	"go.uber.org/goleak"
)

func newStreamTestClient(t *testing.T, opts ...irpc.EndpointOption) *streamTestApiIrpcClient {
	t.Helper()
	// other tests of this package leave their endpoints open: we only check the goroutines started from here.
	// registered first, so it runs after the endpoints are closed
	ignoreOthers := goleak.IgnoreCurrent()
	t.Cleanup(func() { goleak.VerifyNone(t, ignoreOthers) })

	// minimal limits: the streams must not depend on worker slots
	opts = append([]irpc.EndpointOption{irpc.WithParallelWorkers(1), irpc.WithParallelClientCalls(1)}, opts...)
	serviceEp, clientEp, err := testtools.CreateLocalTcpEndpoints(opts...)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		clientEp.Close()
		serviceEp.Close()
		<-clientEp.Context().Done()
		<-serviceEp.Context().Done()
	})
	serviceEp.RegisterService(newStreamTestApiIrpcService(streamTestImpl{}))

	client, err := newStreamTestApiIrpcClient(clientEp)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestStreamNormalMethod(t *testing.T) {
	client := newStreamTestClient(t)
	res, err := client.Add(2, 3)
	if err != nil || res != 5 {
		t.Fatalf("Add(2, 3) = %d, %v", res, err)
	}
}

func TestStreamSum(t *testing.T) {
	client := newStreamTestClient(t)

	in := make(chan int)
	go func() {
		defer close(in)
		for i := 1; i <= 1000; i++ {
			in <- i
		}
	}()
	sum, err := client.Sum(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}
	if sum != 500500 {
		t.Fatalf("sum: got %d, expected 500500", sum)
	}
}

func TestStreamSplit(t *testing.T) {
	client := newStreamTestClient(t)

	lines := make(chan string)
	go func() {
		defer close(lines)
		lines <- "a,b,c"
		lines <- "dd,ee"
	}()

	sepLen, words, lens, err := client.Split(t.Context(), ",", lines)
	if err != nil {
		t.Fatal(err)
	}
	if sepLen != 1 {
		t.Fatalf("sepLen: got %d", sepLen)
	}

	// the implementation writes both channels: they must be read concurrently
	var gotWords []string
	var gotLens []int
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for w := range words {
			gotWords = append(gotWords, w)
		}
	}()
	go func() {
		defer wg.Done()
		for l := range lens {
			gotLens = append(gotLens, l)
		}
	}()
	wg.Wait()

	if !slices.Equal(gotWords, []string{"a", "b", "c", "dd", "ee"}) {
		t.Fatalf("words: %q", gotWords)
	}
	if !slices.Equal(gotLens, []int{5, 5}) {
		t.Fatalf("lens: %v", gotLens)
	}
}

func TestStreamPoints(t *testing.T) {
	client := newStreamTestClient(t)

	points, err := client.Points(t.Context(), 300)
	if err != nil {
		t.Fatal(err)
	}
	i := 0
	for p := range points {
		if p != (streamPoint{X: i, Y: -i}) {
			t.Fatalf("point %d: %+v", i, p)
		}
		i++
	}
	if i != 300 {
		t.Fatalf("got %d points", i)
	}
}

func TestStreamPointsImplError(t *testing.T) {
	client := newStreamTestClient(t)

	points, err := client.Points(t.Context(), -1)
	if err == nil || err.Error() != errNegativePoints.Error() {
		t.Fatalf("expected %q, got %v", errNegativePoints, err)
	}
	if points != nil {
		t.Fatal("expected nil channel")
	}
}

func TestStreamPointsReaderCancels(t *testing.T) {
	client := newStreamTestClient(t)

	ctx, cancel := context.WithCancel(t.Context())
	points, err := client.Points(ctx, 1_000_000)
	if err != nil {
		t.Fatal(err)
	}
	for range 10 {
		<-points
	}
	cancel()
	for range points { // gets closed
	}
}

func TestStreamMaxSessions(t *testing.T) {
	client := newStreamTestClient(t, irpc.WithMaxStreamSessions(2))

	// a returned stream that isn't read keeps its session open
	ctx1, cancel1 := context.WithCancel(t.Context())
	defer cancel1()
	if _, err := client.Points(ctx1, 1_000_000); err != nil {
		t.Fatal(err)
	}
	ctx2, cancel2 := context.WithCancel(t.Context())
	defer cancel2()
	if _, err := client.Points(ctx2, 1_000_000); err != nil {
		t.Fatal(err)
	}

	_, err := client.Points(t.Context(), 1)
	if err == nil || !strings.Contains(err.Error(), irpcgen.ErrStreamTooManySessions.Error()) {
		t.Fatalf("expected %q, got %v", irpcgen.ErrStreamTooManySessions, err)
	}

	// closing a stream frees its session (asynchronously, on the service side)
	cancel1()
	deadline := time.Now().Add(5 * time.Second)
	for {
		points, err := client.Points(t.Context(), 3)
		if err == nil {
			n := 0
			for range points {
				n++
			}
			if n != 3 {
				t.Fatalf("got %d points", n)
			}
			break
		}
		if !strings.Contains(err.Error(), irpcgen.ErrStreamTooManySessions.Error()) {
			t.Fatal(err)
		}
		if time.Now().After(deadline) {
			t.Fatal("session not released after cancel")
		}
		time.Sleep(time.Millisecond)
	}
}

// A misbehaving peer sending the main call twice must not run the implementation twice
// (it used to panic closing the session's ready channel twice, crashing the whole process).
func TestStreamMainCallTwice(t *testing.T) {
	client := newStreamTestClient(t)

	id := irpcgen.NewStreamSessionId()
	if err := client.streamOpen_Points(t.Context(), id); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	errC := make(chan error, 2)
	for range 2 {
		go func() { errC <- client.stream_Points(ctx, id, 1_000_000) }()
	}

	// one of the two is rejected, the other keeps running until we cancel it
	err := <-errC
	if err == nil || !strings.Contains(err.Error(), irpcgen.ErrStreamSessionInUse.Error()) {
		t.Fatalf("expected %q, got %v", irpcgen.ErrStreamSessionInUse, err)
	}
	cancel()
	<-errC
}

// A misbehaving peer overlapping the calls of the same stream is rejected, instead of piling up goroutines.
func TestStreamOverlappingCalls(t *testing.T) {
	client := newStreamTestClient(t)

	id := irpcgen.NewStreamSessionId()
	if err := client.streamOpen_Points(t.Context(), id); err != nil {
		t.Fatal(err)
	}

	// without the main call the session is never ready: Pull blocks and keeps its slot.
	// One of the two is rejected, the other is released when the endpoints close.
	errC := make(chan error, 2)
	for range 2 {
		go func() {
			_, _, err := client.streamPull_Points_p0(t.Context(), id)
			errC <- err
		}()
	}
	err := <-errC
	if err == nil || !strings.Contains(err.Error(), irpcgen.ErrStreamBusy.Error()) {
		t.Fatalf("expected %q, got %v", irpcgen.ErrStreamBusy, err)
	}
}
