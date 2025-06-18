package irpc

import (
	"context"
	"fmt"
	"sync"

	"github.com/marben/irpc/irpcgen"
)

// ourPendingRequest represents a pending request that is being executed on the opposing endpoint
type ourPendingRequest struct {
	reqNum    reqNumT
	resp      irpcgen.Deserializable
	deserErrC chan error
}
type ourPendingRequestsLog struct {
	reqNumsC        chan reqNumT
	m               sync.Mutex
	pendingRequests map[reqNumT]ourPendingRequest
}

func newOurPendingRequestsLog(parallelClientCalls int) *ourPendingRequestsLog {
	reqNumsC := make(chan reqNumT, parallelClientCalls)
	for i := range parallelClientCalls {
		reqNumsC <- reqNumT(i)
	}
	return &ourPendingRequestsLog{
		reqNumsC:        reqNumsC,
		pendingRequests: make(map[reqNumT]ourPendingRequest),
	}
}

func (l *ourPendingRequestsLog) newRequestNumber(ctx context.Context) (reqNumT, error) {
	select {
	case n := <-l.reqNumsC:
		return n, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func (l *ourPendingRequestsLog) addPendingRequest(ctx context.Context, resp irpcgen.Deserializable) (ourPendingRequest, error) {
	reqNum, err := l.newRequestNumber(ctx)
	if err != nil {
		return ourPendingRequest{}, fmt.Errorf("newRequestNumber: %w", err)
	}

	pr := ourPendingRequest{
		reqNum:    reqNum,
		resp:      resp,
		deserErrC: make(chan error, 1),
	}

	l.m.Lock()
	defer l.m.Unlock()
	l.pendingRequests[reqNum] = pr

	return pr, nil
}

func (l *ourPendingRequestsLog) popPendingRequest(reqNum reqNumT) (ourPendingRequest, error) {
	l.m.Lock()
	defer l.m.Unlock()

	pr, found := l.pendingRequests[reqNum]
	if !found {
		return ourPendingRequest{}, fmt.Errorf("pending request %d not found", reqNum)
	}
	delete(l.pendingRequests, reqNum)
	l.reqNumsC <- reqNum
	return pr, nil
}
