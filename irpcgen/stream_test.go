package irpcgen

import (
	"context"
	"errors"
	"testing"
	"time"
)

type testSession struct{ name string }

func ownerCtx(t *testing.T, maxSessions int) (context.Context, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return WithStreamOwner(ctx, NewStreamOwner(ctx, maxSessions)), cancel
}

func TestStreamSessionsOwnerIsolation(t *testing.T) {
	var reg StreamSessions[testSession]
	ctxA, _ := ownerCtx(t, 0)
	ctxB, _ := ownerCtx(t, 0)

	a := &testSession{"a"}
	if err := reg.Add(ctxA, 1, a); err != nil {
		t.Fatal(err)
	}
	if _, err := reg.Get(ctxB, 1); !errors.Is(err, ErrStreamUnknownSession) {
		t.Fatalf("Get from another owner: got %v, expected %v", err, ErrStreamUnknownSession)
	}
	// the same id can be used by another owner
	b := &testSession{"b"}
	if err := reg.Add(ctxB, 1, b); err != nil {
		t.Fatal(err)
	}
	if s, err := reg.Get(ctxA, 1); err != nil || s != a {
		t.Fatalf("Get(A): got %v, %v", s, err)
	}
	if s, err := reg.Get(ctxB, 1); err != nil || s != b {
		t.Fatalf("Get(B): got %v, %v", s, err)
	}
	// Del from another owner doesn't touch the session
	reg.Del(ctxB, 1)
	if _, err := reg.Get(ctxA, 1); err != nil {
		t.Fatal(err)
	}
}

func TestStreamSessionsDuplicate(t *testing.T) {
	var reg StreamSessions[testSession]
	ctx, _ := ownerCtx(t, 1)
	if err := reg.Add(ctx, 1, &testSession{}); err != nil {
		t.Fatal(err)
	}
	if err := reg.Add(ctx, 1, &testSession{}); !errors.Is(err, ErrStreamSessionExists) {
		t.Fatalf("got %v, expected %v", err, ErrStreamSessionExists)
	}
}

func TestStreamSessionsLimit(t *testing.T) {
	// the limit is per owner, across all the registries
	var reg1 StreamSessions[testSession]
	var reg2 StreamSessions[int]
	ctx, _ := ownerCtx(t, 2)
	other, _ := ownerCtx(t, 2)

	if err := reg1.Add(ctx, 1, &testSession{}); err != nil {
		t.Fatal(err)
	}
	if err := reg2.Add(ctx, 1, new(int)); err != nil {
		t.Fatal(err)
	}
	if err := reg1.Add(ctx, 2, &testSession{}); !errors.Is(err, ErrStreamTooManySessions) {
		t.Fatalf("got %v, expected %v", err, ErrStreamTooManySessions)
	}
	if err := reg1.Add(other, 2, &testSession{}); err != nil {
		t.Fatalf("other owner: %v", err)
	}

	reg2.Del(ctx, 1)
	reg2.Del(ctx, 1) // releases only once
	if err := reg1.Add(ctx, 2, &testSession{}); err != nil {
		t.Fatal(err)
	}
	if err := reg1.Add(ctx, 3, &testSession{}); !errors.Is(err, ErrStreamTooManySessions) {
		t.Fatalf("got %v, expected %v", err, ErrStreamTooManySessions)
	}
}

func TestStreamSessionsDroppedWithOwner(t *testing.T) {
	var reg StreamSessions[testSession]
	ctx, cancel := ownerCtx(t, 0)
	for id := range uint64(3) {
		if err := reg.Add(ctx, id, &testSession{}); err != nil {
			t.Fatal(err)
		}
	}
	cancel()

	deadline := time.Now().Add(time.Second)
	for {
		reg.mu.Lock()
		n := len(reg.m)
		reg.mu.Unlock()
		if n == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("sessions not dropped after the owner's context was done")
		}
		time.Sleep(time.Millisecond)
	}
	if err := reg.Add(ctx, 10, &testSession{}); err == nil {
		t.Fatal("Add succeeded on a closed owner")
	}
}

func TestStreamSessionsWithoutOwner(t *testing.T) {
	// a runtime that doesn't provide an owner: one shared, unlimited registry
	var reg StreamSessions[testSession]
	s := &testSession{}
	if err := reg.Add(context.Background(), 1, s); err != nil {
		t.Fatal(err)
	}
	if got, err := reg.Get(context.TODO(), 1); err != nil || got != s {
		t.Fatalf("got %v, %v", got, err)
	}
	reg.Del(context.Background(), 1)
	if _, err := reg.Get(context.Background(), 1); !errors.Is(err, ErrStreamUnknownSession) {
		t.Fatalf("got %v", err)
	}
}

func TestStreamSessionEndsWithOwner(t *testing.T) {
	ctx, cancel := ownerCtx(t, 0)
	callCtx, cancelCall := context.WithCancel(ctx)
	sess := NewStreamSession(callCtx)

	// the session outlives the call that opened it
	cancelCall()
	if err := sess.Context().Err(); err != nil {
		t.Fatalf("session ended with the call: %v", err)
	}

	// and ends with the connection
	cancel()
	select {
	case <-sess.Context().Done():
	case <-time.After(time.Second):
		t.Fatal("session not ended after the owner's context was done")
	}
}

func TestStreamSessionWithoutOwner(t *testing.T) {
	type key struct{}
	callCtx, cancelCall := context.WithCancel(context.WithValue(context.Background(), key{}, "v"))
	sess := NewStreamSession(callCtx)
	cancelCall()
	if err := sess.Context().Err(); err != nil {
		t.Fatalf("session ended with the call: %v", err)
	}
	if sess.Context().Value(key{}) != "v" {
		t.Fatal("session lost the call's values")
	}
	sess.End(ErrStreamSessionEnded)
	if !errors.Is(context.Cause(sess.Context()), ErrStreamSessionEnded) {
		t.Fatal("End didn't end the session")
	}
}
