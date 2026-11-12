package irpc

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/marben/irpc/irpcgen"
)

var errTooManyPendingRequests = errors.New("too many pending requests")

// ourPendingRequest represents a pending request that is being executed on the opposing endpoint
type ourPendingRequest struct {
	reqNum    reqNumT
	resp      irpcgen.Deserializable
	deserErrC chan error
	limited   bool // the request holds a slot of ourPendingRequestsLog.limitedSlots
}

// ourPendingRequestsLog keeps track of the requests we sent to our peer and are waiting for a response.
//
// Request numbers and the parallel calls limit are independent:
//   - limited requests (normal calls) are at most parallelClientCalls at the same time
//   - unlimited requests (streams, whose calls can last for the whole stream) are not counted
//
// Every pending request gets a request number not used by any other pending request.
type ourPendingRequestsLog struct {
	limitedSlots chan struct{} // semaphore for limited requests

	m               sync.Mutex
	pendingRequests map[reqNumT]ourPendingRequest
	nextReqNum      reqNumT // request numbers are assigned round-robin, so a just freed number is not immediately reused
}

func newOurPendingRequestsLog(parallelClientCalls int) *ourPendingRequestsLog {
	return &ourPendingRequestsLog{
		limitedSlots:    make(chan struct{}, parallelClientCalls),
		pendingRequests: make(map[reqNumT]ourPendingRequest),
	}
}

// addPendingRequest registers a new pending request.
// If limited is true, it blocks until there are less than parallelClientCalls limited requests pending (or ctx is done).
func (l *ourPendingRequestsLog) addPendingRequest(ctx context.Context, resp irpcgen.Deserializable, limited bool) (ourPendingRequest, error) {
	if limited {
		select {
		case l.limitedSlots <- struct{}{}:
		case <-ctx.Done():
			return ourPendingRequest{}, ctx.Err()
		}
	}

	l.m.Lock()
	defer l.m.Unlock()

	reqNum, err := l.freeReqNum()
	if err != nil {
		if limited {
			<-l.limitedSlots
		}
		return ourPendingRequest{}, err
	}

	pr := ourPendingRequest{
		reqNum:    reqNum,
		resp:      resp,
		deserErrC: make(chan error, 1),
		limited:   limited,
	}
	l.pendingRequests[reqNum] = pr

	return pr, nil
}

// freeReqNum returns the first request number not used by a pending request. l.m must be held.
func (l *ourPendingRequestsLog) freeReqNum() (reqNumT, error) {
	const reqNumsCount = 1 << 16 // all the values of reqNumT
	for range reqNumsCount {
		n := l.nextReqNum
		l.nextReqNum++
		if _, used := l.pendingRequests[n]; !used {
			return n, nil
		}
	}
	return 0, fmt.Errorf("%w: all %d request numbers are in use", errTooManyPendingRequests, reqNumsCount)
}

func (l *ourPendingRequestsLog) popPendingRequest(reqNum reqNumT) (ourPendingRequest, error) {
	l.m.Lock()
	defer l.m.Unlock()

	pr, found := l.pendingRequests[reqNum]
	if !found {
		return ourPendingRequest{}, fmt.Errorf("pending request %d not found", reqNum)
	}
	delete(l.pendingRequests, reqNum)
	if pr.limited {
		<-l.limitedSlots
	}
	return pr, nil
}
