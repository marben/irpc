package irpc

import (
	"context"
	"fmt"
	"sync"

	"github.com/marben/irpc/irpcgen"
)

// executor executes procedure calls requested from our counterpart
type executor struct {
	wrkrQueue chan struct{}

	ctx context.Context // all workers derive their context from this context

	// active running workers
	// lock m before accessing
	serviceWorkers map[reqNumT]serviceWorker
	m              sync.Mutex // todo: clear up usage
	errC           chan error
}

func newExecutor(ctx context.Context, parallelWorkers int) *executor {
	return &executor{
		wrkrQueue:      make(chan struct{}, parallelWorkers),
		serviceWorkers: make(map[reqNumT]serviceWorker),
		errC:           make(chan error, parallelWorkers), // maybe 1? maybe parallel workers -1?
		ctx:            ctx,
	}
}

// todo: rename to 'execute' or something like that
func (e *executor) startServiceWorker(reqNum reqNumT, rpcExecutor irpcgen.FuncExecutor, sendResponseF func(reqNum reqNumT, respData irpcgen.Serializable) error) error {
	// waits until worker slot is available (blocks here on too many long rpcs)
	select {
	case e.wrkrQueue <- struct{}{}:
	case <-e.ctx.Done():
		return e.ctx.Err()
	}

	// workerCtx is passed to the service's actual implementation
	// cancelling it doesn't mean end of executor
	workerCtx, cancelWorker := context.WithCancelCause(e.ctx)

	rw := serviceWorker{
		cancel: cancelWorker,
	}

	// a goroutine is created for each remote call

	e.m.Lock()
	e.serviceWorkers[reqNum] = rw
	e.m.Unlock()

	go func() {
		// release the worker queue
		defer func() { <-e.wrkrQueue }()

		defer func() {
			e.m.Lock()
			delete(e.serviceWorkers, reqNum)
			e.m.Unlock()
		}()

		resp := rpcExecutor(workerCtx)

		// if executor's context was canceled, we don't even bother with sending response
		if e.ctx.Err() != nil {
			return
		}

		if err := sendResponseF(reqNum, resp); err != nil {
			e.errC <- fmt.Errorf("failed to serialize response %d to connection: %w", reqNum, err)
		}
	}()

	return nil
}

func (e *executor) cancelRequest(rnum reqNumT, cancelErr error) {
	e.m.Lock()
	defer e.m.Unlock()

	sw, found := e.serviceWorkers[rnum]
	if !found {
		// the work may have finished, while this request was on the way.
		// we don't care
		return
	}

	sw.cancel(cancelErr)
}

// a request from opposing endpoint, that we are executing
type serviceWorker struct {
	cancel context.CancelCauseFunc // todo: we don't need this struct for just a function pointer
}
