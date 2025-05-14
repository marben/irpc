package irpc

import (
	"context"
)

type Semaphore struct {
	c chan struct{}
}

func NewContextSemaphore(parallels int) Semaphore {
	return Semaphore{c: make(chan struct{}, parallels)}
}

func (cm Semaphore) Lock(ctx context.Context) (unlock func(), err error) {
	select {
	case <-ctx.Done():
		// log.Println("Semaphore: context expired")
		return nil, ctx.Err()
	case cm.c <- struct{}{}:
		return cm.unlock, nil
	}
}

func (cm Semaphore) unlock() {
	// todo: mabe make idempotent? (that would probably require allocation)
	<-cm.c
}
