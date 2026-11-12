package irpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/marben/irpc/irpcgen"
)

func TestPendingRequestsLimitedAndUnlimited(t *testing.T) {
	l := newOurPendingRequestsLog(2)
	resp := irpcgen.EmptyDeserializable{}

	// fill the limited slots
	a, err := l.addPendingRequest(t.Context(), resp, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.addPendingRequest(t.Context(), resp, true); err != nil {
		t.Fatal(err)
	}

	// a third limited request waits for a slot
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	if _, err := l.addPendingRequest(ctx, resp, true); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected the limited request to wait for a slot, got %v", err)
	}

	// unlimited requests don't wait, and get unique request numbers
	seen := map[reqNumT]bool{}
	for range 100 {
		pr, err := l.addPendingRequest(t.Context(), resp, false)
		if err != nil {
			t.Fatal(err)
		}
		if seen[pr.reqNum] {
			t.Fatalf("request number %d assigned twice", pr.reqNum)
		}
		seen[pr.reqNum] = true
	}

	// popping a limited request frees a slot
	if _, err := l.popPendingRequest(a.reqNum); err != nil {
		t.Fatal(err)
	}
	ctx2, cancel2 := context.WithTimeout(t.Context(), time.Second)
	defer cancel2()
	if _, err := l.addPendingRequest(ctx2, resp, true); err != nil {
		t.Fatalf("expected a free slot after pop, got %v", err)
	}
}

func TestPendingRequestsAllReqNumsUsed(t *testing.T) {
	l := newOurPendingRequestsLog(1)
	resp := irpcgen.EmptyDeserializable{}
	for range 1 << 16 {
		if _, err := l.addPendingRequest(t.Context(), resp, false); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := l.addPendingRequest(t.Context(), resp, false); !errors.Is(err, errTooManyPendingRequests) {
		t.Fatalf("expected errTooManyPendingRequests, got %v", err)
	}
	// the failed limited request must give its slot back
	if _, err := l.addPendingRequest(t.Context(), resp, true); !errors.Is(err, errTooManyPendingRequests) {
		t.Fatalf("expected errTooManyPendingRequests, got %v", err)
	}
	if len(l.limitedSlots) != 0 {
		t.Fatalf("limited slot leaked: %d taken", len(l.limitedSlots))
	}
}
