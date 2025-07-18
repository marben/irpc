package irpctestpkg

import (
	"testing"

	"github.com/marben/irpc/cmd/irpc/test/out"
	"github.com/marben/irpc/cmd/irpc/test/testtools"
)

func TestOutsidePkg(t *testing.T) {
	localEp, remoteEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create enpoints: %v", err)
	}
	defer localEp.Close()
	defer remoteEp.Close()

	remoteEp.RegisterService(newOutsideTestIRpcService(outsideTestImpl{}))

	c, err := newOutsideTestIRpcClient(localEp)
	if err != nil {
		t.Fatalf("newOutsideTestIRpcClient(): %+v", err)
	}

	v3 := 3
	if res := c.addUint8(1, out.Uint8(v3)); res != 4 {
		t.Fatalf("res == %d", res)
	}
}
