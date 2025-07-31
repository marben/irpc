package irpctestpkg

import (
	"testing"

	"github.com/marben/irpc/cmd/irpc/test/out"
	"github.com/marben/irpc/cmd/irpc/test/testtools"
)

func TestOutsidePkgAlias(t *testing.T) {
	localEp, remoteEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create enpoints: %v", err)
	}
	defer localEp.Close()
	defer remoteEp.Close()

	remoteEp.RegisterService(newOutsidepkgaliasIRpcService(&outsidepkgaliasImpl{}))

	c, err := newOutsidepkgaliasIRpcClient(localEp)
	if err != nil {
		t.Fatalf("newOutsidepkgaliasIRpcClient(): %+v", err)
	}

	if res := c.add(out.Uint8(1), out.Uint8(2)); res != 3 {
		t.Fatalf("c.add(1, 2) == %d", res)
	}

	if res := c.add2(out.Uint8(4), out.Uint8(3)); res != 7 {
		t.Fatalf("c.add2(4, 4) == %d", res)
	}

	if res := c.add3(6, out.Uint8(8)); res != out.Uint8(14) {
		t.Fatalf("c.add3(6, 8) == %d", res)
	}
}
