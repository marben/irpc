package irpctestpkg

import (
	"testing"
	"time"

	"github.com/marben/irpc/pkg/irpc"
	"github.com/marben/irpc/test/testtools"
)

func TestEndpointClose(t *testing.T) {
	p1, p2 := testtools.NewDoubleEndedPipe()

	ep1 := irpc.NewEndpoint()
	errC1 := make(chan error)
	go func() {
		errC1 <- ep1.Serve(p1)
	}()
	// c := newEndpointApiRpcClient(clientEp)

	ep2 := irpc.NewEndpoint()
	errC2 := make(chan error)
	go func() {
		errC2 <- ep2.Serve(p2)
	}()

	time.Sleep(10 * time.Millisecond)
	if err := ep1.Close(); err != nil {
		t.Fatalf("ep1.Close(): %v", err)
	}
	if err := ep2.Close(); err != nil {
		t.Fatalf("ep2.Close(): %v", err)
	}

	t.Logf("ep1.Serve(): %v", <-errC1)
	t.Logf("ep2.Serve(): %v", <-errC2)
}
