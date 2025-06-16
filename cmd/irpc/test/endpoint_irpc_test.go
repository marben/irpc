package irpctestpkg

import (
	"errors"
	"testing"
	"time"

	"github.com/marben/irpc"
	"github.com/marben/irpc/cmd/irpc/test/testtools"
)

func TestEndpointClose(t *testing.T) {
	p1, p2 := testtools.NewDoubleEndedPipe()

	ep1 := irpc.NewEndpoint(p1)

	ep2 := irpc.NewEndpoint(p2)

	time.Sleep(10 * time.Millisecond)
	if err := ep1.Close(); err != nil {
		t.Fatalf("ep1.Close(): %v", err)
	}
	<-ep2.Ctx.Done() // need to wait for readMsgs goroutine to process the 'closingNowPacket'
	if err := ep2.Close(); !errors.Is(err, irpc.ErrEndpointClosed) {
		t.Fatalf("ep2.Close(): %v", err)
	}

	if err := ep1.Close(); !errors.Is(err, irpc.ErrEndpointClosed) {
		t.Fatalf("ep1.Close(): %+v", err)
	}

	if err := ep2.Close(); !errors.Is(err, irpc.ErrEndpointClosed) {
		t.Fatalf("ep2.Close(): %+v", err)
	}
}
