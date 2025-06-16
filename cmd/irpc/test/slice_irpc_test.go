package irpctestpkg

import (
	"slices"
	"testing"

	"github.com/marben/irpc/cmd/irpc/test/testtools"
)

func TestSlice(t *testing.T) {
	localEp, remoteEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create enpoints: %v", err)
	}

	// service
	skew := 5
	remoteEp.RegisterService(newSliceTestIRpcService(sliceTestImpl{skew: skew}))

	// client
	c, err := newSliceTestIRpcClient(localEp)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// SLICE SUM (PASS SLICE)
	res := c.SliceSum([]int{1, 2, 3})
	if res != 1+2+3+5 {
		t.Fatalf("unexpected slice sum: %d", res)
	}

	// VECT MULT (RETURN SLICE)
	resV := c.VectMult([]int{1, 2, 3}, 3)
	if cmp := slices.Compare(resV, []int{1*3 + 5, 2*3 + 5, 3*3 + 5}); cmp != 0 {
		t.Fatalf("unexpected vector multiplication result: %v", resV)
	}

	// SLICE OF FLOAT64 SUM
	resF := c.SliceOfFloat64Sum([]float64{1.1, 2.2})
	if resF != 1.1+2.2+float64(skew) {
		t.Fatalf("unexpected float64 sum %f", resF)
	}

	// SLICE OF SLICE
	resS := c.SliceOfSlicesSum([][]int{{1, 2}, {3, 4, 5}})
	if resS != 1+2+3+4+5 {
		t.Fatalf("unexpected slice sum: %d", resS)
	}

	t.Log("slice of bytes")
	resSB := c.SliceOfBytesSum([]byte{1, 2, 3, 4})
	if resSB != 1+2+3+4 {
		t.Fatalf("SliceOfBytesSum(): %d", resSB)
	}
}
