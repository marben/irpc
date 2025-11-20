package irpctestpkg

import (
	"maps"
	"testing"

	"github.com/marben/irpc/cmd/irpc/test/testtools"
)

func TestMap(t *testing.T) {
	localEp, remoteEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create enpoints: %v", err)
	}
	defer localEp.Close()

	remoteEp.RegisterService(newMapTestIrpcService(&mapTestImpl{}))

	c, err := newMapTestIrpcClient(localEp)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// MAP SUM
	m := map[int]float64{1: 1.25, 8: 2.25, 5: 3.25, 66: -5.25}
	ksum, valsum := c.mapSum(m)
	if ksum != 80 {
		t.Fatalf("ksum: %d", ksum)
	}
	if valsum != 1.5 {
		t.Fatalf("valsum: %f", valsum)
	}
}

func TestSumStructs(t *testing.T) {
	localEp, remoteEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create enpoints: %v", err)
	}
	defer localEp.Close()

	remoteEp.RegisterService(newMapTestIrpcService(&mapTestImpl{}))

	c, err := newMapTestIrpcClient(localEp)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	m := map[intStruct]intStruct{
		{1, 2, 3, 4}: {5, 6, 7, 8},
		{1, 2, 3, 5}: {5, 6, 7, 8},
		{1, 2, 3, 6}: {5, 6, 7, 8}}
	t.Logf("len(m): %d", len(m))
	ksum, valsum := c.sumStructs(m)
	if ksum != 3*(1+2+3+4)+1+2 {
		t.Fatalf("ksum: %d", ksum)
	}
	if valsum != 3*(5+6+7+8) {
		t.Fatalf("valsum: %d", valsum)
	}
}

func TestSumMapSlices(t *testing.T) {
	localEp, remoteEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create enpoints: %v", err)
	}
	defer localEp.Close()

	remoteEp.RegisterService(newMapTestIrpcService(&mapTestImpl{}))

	c, err := newMapTestIrpcClient(localEp)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	m := map[intStruct][]intStruct{
		{1, 2, 3, 4}: {{5, 6, 7, 8}, {5, 6, 7, 8}},
		{1, 2, 3, 5}: {{5, 6, 7, 8}, {5, 6, 7, 8}},
		{1, 2, 3, 6}: {{5, 6, 7, 8}, {5, 6, 7, 8}},
	}
	t.Logf("len(m): %d", len(m))
	ksum, valsum := c.sumSlices(m)
	if ksum != 3*(1+2+3+4)+1+2 {
		t.Fatalf("ksum: %d", ksum)
	}
	if valsum != 6*(5+6+7+8) {
		t.Fatalf("valsum: %d", valsum)
	}
}

func TestEmptyInterfaceMapSum(t *testing.T) {
	localEp, remoteEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create enpoints: %v", err)
	}
	defer localEp.Close()

	remoteEp.RegisterService(newMapTestIrpcService(&mapTestImpl{}))

	c, err := newMapTestIrpcClient(localEp)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	m := make(map[int]interface{})

	var x, y interface{}
	t.Logf("x == nil: %v", x == nil)
	m[1] = x
	m[1] = y
	m[6] = x
	m[7] = nil

	t.Logf("sending map: %v", m)

	var localSum int
	for v := range m {
		localSum += v
	}

	t.Logf("local sum: %d", localSum)

	reflectMap := c.emptyInterfaceMapReflect(m)

	var remoteSum int
	for v := range reflectMap {
		remoteSum += v
	}

	if localSum != remoteSum {
		t.Fatalf("localSum %d != remoteSum %d", localSum, remoteSum)
	}

	if !maps.Equal(m, reflectMap) {
		t.Fatalf("maps are not equal! %v != %v", m, reflectMap)
	}
}
