package bidirectional

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

//go:generate go tool irpc

// LogPrefix returns the padded "[service]" prefix used by the example loggers, so the output of each service group is easy to tell apart.
func LogPrefix(service string) string {
	return fmt.Sprintf("%-22s", "["+service+"]")
}

// NewServiceLogger returns a logger whose messages are prefixed with the service name.
func NewServiceLogger(service string) *log.Logger {
	return log.New(os.Stderr, LogPrefix(service), log.LstdFlags|log.Lmsgprefix)
}

type ExampleData struct {
	Name    string
	Surname string
	Age     int
}

// BarService is what the server exposes to connected clients.
// Clients can ask for the current time, compute a SHA-256 hash of arbitrary bytes or get some example data.
type BarService interface {
	GetTime() (time.Time, error)
	HashBytes(data []byte) ([]byte, error)
	GetExampleData() ExampleData
}

// FooService is what clients expose back to the server.
// The server can autonomously call these methods on any connected client.
type FooService interface {
	Ping() (string, error)
	CheckDataValid(data ExampleData) error
}

// FizzService is exposed by the client.
type FizzService interface {
	IsOdd(number int) bool
}

// BoltService is exposed by the server.
type BoltService interface {
	IsEven(number int) bool
}

const LongWorkDelay = 12 * time.Second

// ContextCancellationService shows that cancelling the caller's context is propagated to the remote side.
type ContextCancellationService interface {
	LongWorkThatWillBeCancelled(ctx context.Context) error
}

// ContextCancellationImpl is shared by client and server.
// Direction is only used in the logs, to tell who is calling (e.g. "client->server").
type ContextCancellationImpl struct {
	Direction string
}

var ctxLog = NewServiceLogger("ContextCancellation")

func (c *ContextCancellationImpl) LongWorkThatWillBeCancelled(ctx context.Context) error {
	select {
	case <-ctx.Done():
		ctxLog.Printf("%s LongWorkThatWillBeCancelled() cancelled by the caller: %v", c.Direction, ctx.Err())
		return ctx.Err()

	case <-time.After(LongWorkDelay):
		ctxLog.Printf("%s LongWorkThatWillBeCancelled() completed", c.Direction)
		return nil
	}
}

// CallLongWorkPeriodically calls LongWorkThatWillBeCancelled on the remote side every interval,
// with a context that expires before LongWorkDelay, so the cancellation must be propagated to the remote side.
// It returns when ctx is done.
func CallLongWorkPeriodically(ctx context.Context, svc ContextCancellationService, interval time.Duration, direction string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctxLog.Printf("%s LongWorkThatWillBeCancelled() started", direction)
			callCtx, cancel := context.WithTimeout(ctx, LongWorkDelay-2*time.Second)
			err := svc.LongWorkThatWillBeCancelled(callCtx)
			cancel()
			ctxLog.Printf("%s LongWorkThatWillBeCancelled() => %v", direction, err)

		case <-ctx.Done():
			return
		}
	}
}
