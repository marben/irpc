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

	if !c.containsSaturday2(namedWeekDaysSliceType{Monday2, Saturday2}) {
		t.Fatal("there is saturday2!")
	}

	if sum := c.namedBytesSum([]byte{1, 2, 3}); sum != 6 {
		t.Fatalf("namedBytesSum: %d", sum)
	}

	if mapSum := c.namedMapSum(namedMap{1: 2.25, 2: 2.25}); mapSum != 7.5 {
		t.Fatalf("mapsum: %f", mapSum)
	}
}
