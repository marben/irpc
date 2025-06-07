package irpc

import (
	"context"
	"fmt"
	"github.com/marben/irpc/irpcgen"
	"sync"
)

// executor executes procedure calls requested from our counterpart
type executor struct {
	wrkrQueue chan struct{}

	ctx context.Context // all workers derive their context from this context

	// active running workers
	// lock m before accessing
	serviceWorkers map[ReqNumT]serviceWorker
	m              sync.Mutex // todo: clear up usage
	errC           chan error
}

func newExecutor(ctx context.Context) *executor {
	return &executor{
		wrkrQueue:      make(chan struct{}, ParallelWorkers),
		serviceWorkers: make(map[ReqNumT]serviceWorker),
		errC:           make(chan error, ParallelWorkers), // maybe 1? maybe parallel workers -1?
		ctx:            ctx,
	}
}

// cancels context of all workers
func (e *executor) cancelAllWorkers(err error) {
	e.m.Lock()
	defer e.m.Unlock()

	for _, w := range e.serviceWorkers {
		w.cancel(err)
	}
}

// todo: rename to 'execute' or something like that
func (e *executor) startServiceWorker(reqNum ReqNumT, rpcExecutor irpcgen.FuncExecutor, serialize *serializer) error {
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

		if err := serialize.sendResponse(reqNum, resp); err != nil {
			e.errC <- fmt.Errorf("failed to serialize response %d to connection: %w", reqNum, err)
		}
	}()

	return nil
}

func (e *executor) cancelRequest(rnum ReqNumT, cancelErr error) {
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
