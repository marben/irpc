package irpctestpkg

import (
	"testing"

	"github.com/marben/irpc/cmd/irpc/test/testtools"
)

func TestEdgecases(t *testing.T) {
	localEp, remoteEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create enpoints: %v", err)
	}
	defer localEp.Close()
	defer remoteEp.Close()

	noReturnC := make(chan int, 1)
	nothingAtAllC := make(chan int, 1)
	remoteEp.RegisterService(newEdgeCasesIRpcService(edgeCasesImpl{
		noReturnFunc:     func(i int) { noReturnC <- i },
		noParamsFunc:     func() int { return 9 },
		nothingAtAllFunc: func() { nothingAtAllC <- 12 },
	}))

	c, err := newEdgeCasesIRpcClient(localEp)
	if err != nil {
		t.Fatalf("failed to create client: %+v", err)
	}

	c.noReturn(5)
	noRtnRtn := <-noReturnC
	if noRtnRtn != 5 {
		t.Fatalf("noReturn(5): %d", noRtnRtn)
	}

	noParamsRes := c.noParams()
	if noParamsRes != 9 {
		t.Fatalf("noParams(): %d", noParamsRes)
	}

	c.nothingAtAll()
	nothingAtAllRes := <-nothingAtAllC
	if nothingAtAllRes != 12 {
		t.Fatalf("nothingAtAll(): %d", nothingAtAllRes)
	}
}
