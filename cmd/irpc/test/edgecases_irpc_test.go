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
	unnamedIntParamC := make(chan int, 1)
	mixedParamIdsC := make(chan int, 1)
	remoteEp.RegisterService(newEdgeCasesIRpcService(edgeCasesImpl{
		noReturnFunc:        func(i int) { noReturnC <- i },
		noParamsFunc:        func() int { return 9 },
		nothingAtAllFunc:    func() { nothingAtAllC <- 12 },
		unnamedIntParamFunc: func(i int) { unnamedIntParamC <- i },
		mixedParamIdsFunc:   func(i int) { mixedParamIdsC <- i },
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

	c.unnamedIntParam(99)
	unnamedIntParamRes := <-unnamedIntParamC
	if unnamedIntParamRes != 99 {
		t.Fatalf("unnamedIntParamRes: %d", unnamedIntParamRes)
	}

	c.mixedParamIds(4, 2, struct{ a int }{9})
	mixedParamIdsRes := <-mixedParamIdsC
	if mixedParamIdsRes != 15 {
		t.Fatalf("mixedParamIdsRes: %d", mixedParamIdsRes)
	}
}
