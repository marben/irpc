package irpc

import (
	"context"
	"fmt"
	"sync"

	"github.com/marben/irpc/irpcgen"
)

// executor runs rpcExecutor functions in go routine
// limits the maximum number of parallel go routines
// allows for cancellation of running rpcExecutors through context
type executor struct {
	wrkrQueue chan struct{}

	ctx context.Context // all workers derive their context from this context

	// active running workers
	serviceWorkers map[reqNumT]serviceWorker
	m              sync.Mutex

	errC chan error
}

func newExecutor(ctx context.Context, parallelWorkers int) *executor {
	return &executor{
		wrkrQueue:      make(chan struct{}, parallelWorkers),
		serviceWorkers: make(map[reqNumT]serviceWorker),
		errC:           make(chan error, parallelWorkers), // maybe 1? maybe parallel workers -1?
		ctx:            ctx,
	}
}

func (e *executor) addWorker(reqNum reqNumT, wrkr serviceWorker) {
	e.m.Lock()
	defer e.m.Unlock()

	e.serviceWorkers[reqNum] = wrkr
}

func (e *executor) delWorker(reqNum reqNumT) {
	e.m.Lock()
	defer e.m.Unlock()

	delete(e.serviceWorkers, reqNum)
}

// runServiceWorker waits for worker slot and then runs rpcExecutor in a new goroutine
// returns once work was succesfully started
func (e *executor) runServiceWorker(reqNum reqNumT, rpcExecutor irpcgen.FuncExecutor, sendResponseF func(reqNum reqNumT, respData irpcgen.Serializable) error) error {
	// waits until worker slot is available (blocks here on too many long rpcs)
	select {
	case e.wrkrQueue <- struct{}{}:
	case <-e.ctx.Done():
		return e.ctx.Err()
	}

	// workerCtx is passed to the service's actual implementation
	// cancelling it doesn't mean end of executor
	workerCtx, cancelWorker := context.WithCancelCause(e.ctx)

	wrkr := serviceWorker{
		cancel: cancelWorker,
	}

	// a goroutine is created for each remote call

	e.addWorker(reqNum, wrkr)
	go func() {
		// release the worker queue
		defer func() { <-e.wrkrQueue }()

		defer e.delWorker(reqNum)

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
