package irpcgen

// Support for channel streams in generated code.
//
// A method with channels (<-chan T as parameters: caller -> implementation, <-chan T as results:
// implementation -> caller) is generated as a set of normal calls sharing a session:
//   - Open:  registers the session on the service (fast). Done first, so the other calls always find it.
//   - main:  carries the scalar parameters and lives as long as the session.
//     Its context is the session context: cancelling it on the caller cancels the session on the service.
//   - Push/Close: one pair for each channel parameter, moving the values in batches (backpressure included).
//   - Ready: only for methods returning channels: waits for the implementation and returns its scalar results.
//   - Pull:  one for each returned channel, a long poll returning the values in batches.
//
// The session ends when the reader stops reading: on the service side when the implementation returns
// (or, for returned channels, when the caller has received everything), on the caller side when it
// cancels the context passed to the call.

import (
	"context"
	"errors"
	"math/rand/v2"
	"sync"
	"sync/atomic"
)

// StreamMaxBatch is the maximum number of values moved by a single Push or Pull call.
const StreamMaxBatch = 64

var (
	ErrStreamUnknownSession  = errors.New("irpc stream: unknown session")
	ErrStreamSessionExists   = errors.New("irpc stream: session already exists")
	ErrStreamTooManySessions = errors.New("irpc stream: too many open sessions")
	ErrStreamSessionEnded    = errors.New("irpc stream: session ended")
	ErrStreamClosed          = errors.New("irpc stream: channel already closed")
	ErrStreamSessionInUse    = errors.New("irpc stream: session already started")
	ErrStreamBusy            = errors.New("irpc stream: overlapping call on the same stream")
)

// NewStreamSessionId returns a new session id, chosen by the caller.
// The service keeps the sessions of each connection apart (see [StreamOwner]),
// so the id only has to be unique among the sessions opened by one caller.
func NewStreamSessionId() uint64 {
	return rand.Uint64()
}

// StreamOwner identifies the connection (the endpoint) the stream sessions belong to,
// and limits the number of sessions it can keep open at the same time.
//
// The runtime creates one for each endpoint and puts it in the context of every call
// it serves (see [WithStreamOwner]). A session can only be used from the connection that opened it,
// and the sessions of a connection are dropped when the owner's context is done.
type StreamOwner struct {
	ctx context.Context
	max int // <= 0 means unlimited

	mu   sync.Mutex
	open int
}

// NewStreamOwner returns an owner whose sessions are dropped once ctx is done.
// maxSessions limits the sessions open at the same time (all the stream methods together), <= 0 means unlimited.
func NewStreamOwner(ctx context.Context, maxSessions int) *StreamOwner {
	return &StreamOwner{ctx: ctx, max: maxSessions}
}

func (o *StreamOwner) acquire() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.max > 0 && o.open >= o.max {
		return ErrStreamTooManySessions
	}
	o.open++
	return nil
}

func (o *StreamOwner) release(n int) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.open -= n
}

type streamOwnerKey struct{}

// WithStreamOwner returns a copy of ctx carrying owner.
func WithStreamOwner(ctx context.Context, owner *StreamOwner) context.Context {
	return context.WithValue(ctx, streamOwnerKey{}, owner)
}

// sharedStreamOwner is used when the runtime doesn't provide an owner:
// all the callers share the sessions and there is no limit.
var sharedStreamOwner = NewStreamOwner(context.Background(), 0)

func streamOwnerFrom(ctx context.Context) *StreamOwner {
	if o, ok := ctx.Value(streamOwnerKey{}).(*StreamOwner); ok && o != nil {
		return o
	}
	return sharedStreamOwner
}

// StreamSessions is the registry of the open sessions of one method, on the service side.
// Sessions are kept per owner (see [StreamOwner]): ctx is the context of the call, which carries the owner.
type StreamSessions[S any] struct {
	mu sync.Mutex
	m  map[*StreamOwner]*ownerSessions[S]
}

type ownerSessions[S any] struct {
	m    map[uint64]*S
	stop func() bool // stops the cleanup registered on the owner's context
}

func (s *StreamSessions[S]) Add(ctx context.Context, id uint64, sess *S) error {
	owner := streamOwnerFrom(ctx)
	if err := owner.ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	own := s.m[owner]
	if own != nil {
		if _, found := own.m[id]; found {
			return ErrStreamSessionExists
		}
	}
	if err := owner.acquire(); err != nil {
		return err
	}
	if own == nil {
		if s.m == nil {
			s.m = make(map[*StreamOwner]*ownerSessions[S])
		}
		own = &ownerSessions[S]{m: make(map[uint64]*S)}
		// the connection is gone: nobody can use its sessions anymore
		own.stop = context.AfterFunc(owner.ctx, func() { s.dropOwner(owner, own) })
		s.m[owner] = own
	}
	own.m[id] = sess
	return nil
}

func (s *StreamSessions[S]) dropOwner(owner *StreamOwner, own *ownerSessions[S]) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.m[owner] != own {
		return
	}
	delete(s.m, owner)
	owner.release(len(own.m))
}

func (s *StreamSessions[S]) Get(ctx context.Context, id uint64) (*S, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if own := s.m[streamOwnerFrom(ctx)]; own != nil {
		if sess, found := own.m[id]; found {
			return sess, nil
		}
	}
	return nil, ErrStreamUnknownSession
}

func (s *StreamSessions[S]) Del(ctx context.Context, id uint64) {
	owner := streamOwnerFrom(ctx)
	s.mu.Lock()
	defer s.mu.Unlock()
	own := s.m[owner]
	if own == nil {
		return
	}
	if _, found := own.m[id]; !found {
		return
	}
	delete(own.m, id)
	owner.release(1)
	if len(own.m) == 0 {
		own.stop()
		delete(s.m, owner)
	}
}

// StreamSession is the service side state shared by every generated session.
type StreamSession struct {
	ctx    context.Context
	cancel context.CancelCauseFunc

	ready          chan struct{} // closed once the implementation has returned
	readyTaken     chan struct{} // closed once the caller has collected the results with Ready
	readyTakenOnce sync.Once

	// the generated caller runs the main call once, and never overlaps the other calls of the same stream:
	// a peer doing it is misbehaving, and is rejected before it can start the implementation twice
	// or pile up goroutines on calls that don't take a worker slot
	started atomic.Bool
	busyMu  sync.Mutex
	busy    map[string]bool
}

// NewStreamSession creates the session of a stream, opened by the call whose context is callCtx (Open).
//
// The session outlives that call, so its context doesn't derive from callCtx: it derives from the context
// of the connection the call came from (see [StreamOwner]), and it ends when the connection closes.
// The main call, once it arrives, also ends it when cancelled (see [StreamSession.Link]).
//
// Without an owner in callCtx (a runtime that doesn't provide one) there is no connection context:
// the session keeps callCtx's values, and only the main call ends it.
func NewStreamSession(callCtx context.Context) *StreamSession {
	parent := context.WithoutCancel(callCtx)
	if owner, ok := callCtx.Value(streamOwnerKey{}).(*StreamOwner); ok && owner != nil {
		parent = owner.ctx
	}
	ctx, cancel := context.WithCancelCause(parent)
	return &StreamSession{
		ctx:        ctx,
		cancel:     cancel,
		ready:      make(chan struct{}),
		readyTaken: make(chan struct{}),
	}
}

// Context is the context passed to the implementation.
func (s *StreamSession) Context() context.Context {
	return s.ctx
}

// Link cancels the session when callCtx (the main call's context) is cancelled. stop must be deferred.
func (s *StreamSession) Link(callCtx context.Context) (stop func() bool) {
	return context.AfterFunc(callCtx, func() { s.cancel(context.Cause(callCtx)) })
}

// Start lets the main call run once: the next ones fail with [ErrStreamSessionInUse].
func (s *StreamSession) Start() error {
	if !s.started.CompareAndSwap(false, true) {
		return ErrStreamSessionInUse
	}
	return nil
}

// Enter lets one call at a time run for key (a channel, or Ready): an overlapping call fails with [ErrStreamBusy].
// exit must be deferred.
func (s *StreamSession) Enter(key string) (exit func(), err error) {
	s.busyMu.Lock()
	defer s.busyMu.Unlock()
	if s.busy[key] {
		return nil, ErrStreamBusy
	}
	if s.busy == nil {
		s.busy = make(map[string]bool)
	}
	s.busy[key] = true
	return func() {
		s.busyMu.Lock()
		defer s.busyMu.Unlock()
		delete(s.busy, key)
	}, nil
}

// End cancels the session.
func (s *StreamSession) End(cause error) {
	s.cancel(cause)
}

// MarkReady signals that the implementation has returned and its results are stored in the session.
func (s *StreamSession) MarkReady() {
	close(s.ready)
}

// WaitReady waits for MarkReady, or for the session to end.
func (s *StreamSession) WaitReady() error {
	select {
	case <-s.ready:
		return nil
	case <-s.ctx.Done():
		return context.Cause(s.ctx)
	}
}

// TakeReady is WaitReady for the Ready call: it also signals that the results have been collected.
func (s *StreamSession) TakeReady() error {
	if err := s.WaitReady(); err != nil {
		return err
	}
	s.readyTakenOnce.Do(func() { close(s.readyTaken) })
	return nil
}

// WaitReadyTaken waits until the caller has collected the results (or the session ends).
// The main call uses it when the implementation failed, so it doesn't delete the session before Ready has seen the error.
func (s *StreamSession) WaitReadyTaken() {
	select {
	case <-s.readyTaken:
	case <-s.ctx.Done():
	}
}

// WaitDrained waits until every returned channel has been received by the caller, or the session ends.
func (s *StreamSession) WaitDrained(drained ...<-chan struct{}) error {
	for _, d := range drained {
		select {
		case <-d:
		case <-s.ctx.Done():
			return context.Cause(s.ctx)
		}
	}
	return nil
}

// InStream is the service side of a channel parameter: the caller writes, the implementation reads.
type InStream[T any] struct {
	ch chan T

	mu     sync.Mutex // serializes Deliver and Close, so we never send on a closed channel
	closed bool
}

func NewInStream[T any]() *InStream[T] {
	return &InStream[T]{ch: make(chan T)}
}

// Chan is the channel passed to the implementation.
func (s *InStream[T]) Chan() <-chan T {
	return s.ch
}

// Deliver hands values to the implementation, one by one.
// It returns once all of them have been read (that is the backpressure) or when ctx is done.
func (s *InStream[T]) Deliver(ctx context.Context, values []T) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrStreamClosed
	}
	for _, v := range values {
		select {
		case s.ch <- v:
		case <-ctx.Done():
			return context.Cause(ctx)
		}
	}
	return nil
}

func (s *InStream[T]) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed {
		s.closed = true
		close(s.ch)
	}
}

// OutStream is the service side of a returned channel: the implementation writes, the caller reads.
type OutStream[T any] struct {
	ch          <-chan T
	drained     chan struct{}
	drainedOnce sync.Once
}

func NewOutStream[T any](ch <-chan T) *OutStream[T] {
	s := &OutStream[T]{ch: ch, drained: make(chan struct{})}
	if ch == nil {
		// a nil channel would block forever: we treat it as an empty stream
		s.markDrained()
	}
	return s
}

func (s *OutStream[T]) markDrained() {
	s.drainedOnce.Do(func() { close(s.drained) })
}

// Drained is closed once the channel has been closed by the implementation and every value has been pulled.
func (s *OutStream[T]) Drained() <-chan struct{} {
	return s.drained
}

// Pull waits for at least one value, then also takes the ones already available (up to StreamMaxBatch).
func (s *OutStream[T]) Pull(ctx context.Context) (values []T, closed bool, err error) {
	select {
	case <-s.drained:
		return nil, true, nil
	default:
	}

	select {
	case v, ok := <-s.ch:
		if !ok {
			s.markDrained()
			return nil, true, nil
		}
		values = append(values, v)
	case <-ctx.Done():
		return nil, false, context.Cause(ctx)
	}

	values, closed = appendReady(values, s.ch)
	if closed {
		s.markDrained()
	}
	return values, closed, nil
}

// PushFunc sends a batch of values of a channel parameter to the other side.
type PushFunc[T any] func(ctx context.Context, batch []T) error

// CloseFunc tells the other side that the caller closed a channel parameter.
type CloseFunc func(ctx context.Context) error

// PullFunc fetches the next batch of values of a returned channel.
// closed reports that the other side closed the channel: values is the last batch.
type PullFunc[T any] func(ctx context.Context) (values []T, closed bool, err error)

// PumpStream is the caller side of a channel parameter: it reads ch and pushes it in batches.
// When ch is closed it calls closeFn. It stops when ctx is done or a push fails
// (for example because the implementation returned and the session ended).
func PumpStream[T any](ctx context.Context, ch <-chan T, push PushFunc[T], closeFn CloseFunc) error {
	for {
		var batch []T

		// block until the first value (or ctx done)
		select {
		case v, ok := <-ch:
			if !ok {
				return closeFn(ctx)
			}
			batch = append(batch, v)
		case <-ctx.Done():
			return context.Cause(ctx)
		}

		// then take the values already queued, up to StreamMaxBatch, without waiting
		batch, closed := appendReady(batch, ch)
		if err := push(ctx, batch); err != nil {
			return err
		}

		if closed {
			return closeFn(ctx)
		}
	}
}

// appendReady appends to batch the values already queued in ch, up to StreamMaxBatch, without waiting for more.
// closed reports that ch was found closed.
func appendReady[T any](batch []T, ch <-chan T) (_ []T, closed bool) {
	for len(batch) < StreamMaxBatch {
		select {
		case v, ok := <-ch:
			if !ok {
				return batch, true
			}
			batch = append(batch, v)
		default:
			return batch, false
		}
	}
	return batch, false
}

// GoPumpStream runs PumpStream in a new goroutine. wait waits for it to finish.
//
// The pump's error is not reported: when the session ends normally the pushes fail as expected,
// and on a network error the main call fails as well, with the error worth returning.
func GoPumpStream[T any](ctx context.Context, ch <-chan T, push PushFunc[T], closeFn CloseFunc) (wait func()) {
	done := make(chan struct{})

	go func() {
		defer close(done)
		_ = PumpStream(ctx, ch, push, closeFn)
	}()

	return func() { <-done }
}

// ReceiveStream is the caller side of a returned channel: it pulls batches and writes them into out.
// It closes out when it returns.
func ReceiveStream[T any](ctx context.Context, out chan<- T, pull PullFunc[T]) error {
	defer close(out)
	for {
		values, closed, err := pull(ctx)
		if err != nil {
			return err
		}

		for _, v := range values {
			select {
			case out <- v:
			case <-ctx.Done():
				return context.Cause(ctx)
			}
		}

		if closed {
			return nil
		}
	}
}

// GoReceiveStream runs ReceiveStream in a new goroutine. wait waits for it and returns its error.
func GoReceiveStream[T any](ctx context.Context, out chan<- T, pull PullFunc[T]) (wait func() error) {
	errC := make(chan error, 1)

	go func() {
		errC <- ReceiveStream(ctx, out, pull)
	}()

	return func() error { return <-errC }
}
