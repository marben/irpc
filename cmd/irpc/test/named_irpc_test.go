package irpctestpkg

import (
	"testing"

	"github.com/marben/irpc/cmd/irpc/test/testtools"
)

func TestNamed(t *testing.T) {
	localEp, remoteEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create enpoints: %v", err)
	}
	defer localEp.Close()
	defer remoteEp.Close()

	remoteEp.RegisterService(newNamedTestIRpcService(namedTestImpl{}))

	c, err := newNamedTestIRpcClient(localEp)
	if err != nil {
		t.Fatalf("failed to create client: %+v", err)
	}

	if !c.isWeekend(Sunday) {
		t.Fatalf("Sunday is weekend!")
	}
	if c.isWeekend2(Wednesday2) {
		t.Fatalf("Wednesday2 is not a weekend 2!")
	}

	if c.containsSaturday([]weekDay{Monday, Wednesday, Friday}) {
		t.Fatal("there is no saturday!")
	}

	if !c.containsSaturday([]weekDay{Monday, Saturday, Wednesday, Friday}) {
		t.Fatal("there is saturday!")
	}
}
