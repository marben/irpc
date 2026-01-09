package irpctestpkg

import (
	"bytes"
	"testing"

	"github.com/marben/irpc/cmd/irpc/test/testtools"
)

func TestPointer(t *testing.T) {
	localEp, remoteEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create endpoints: %v", err)
	}
	defer localEp.Close()
	defer remoteEp.Close()

	impl := newPointerTestImpl()
	remoteEp.RegisterService(newPointerTestIrpcService(impl))

	c, err := newPointerTestIrpcClient(localEp)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	img := c.getImage()
	if img.Bounds() != impl.img.Bounds() || !bytes.Equal(img.Pix, impl.img.Pix) {
		t.Fatalf("images are not equal")
	}

	var intval int = 1
	pInt1 := &intval
	if c.isNil(pInt1) {
		t.Fatalf("%v == nil", pInt1)
	}

	var pInt2 *int
	if !c.isNil(pInt2) {
		t.Fatalf("%v != nil", pInt2)
	}

	if !c.isNil(nil) {
		t.Fatalf("nil != nil")
	}
}
