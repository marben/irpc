package irpctestpkg

import (
	"bytes"
	"slices"
	"testing"

	"github.com/marben/irpc/cmd/irpc/test/testtools"
	"github.com/marben/irpc/irpcgen"
)

func TestSlice(t *testing.T) {
	localEp, remoteEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create enpoints: %v", err)
	}

	// service
	skew := 5
	remoteEp.RegisterService(newSliceTestIrpcService(sliceTestImpl{skew: skew}))

	// client
	c, err := newSliceTestIrpcClient(localEp)
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

	t.Log("slice of uint8")
	resSU8 := c.sliceOfUint8([]uint8{1, 2, 3, 4})
	if !slices.Equal(resSU8, []uint8{4, 3, 2, 1}) {
		t.Fatalf("sliceOfUint8(): %v", resSU8)
	}

	t.Log("slice of structs")
	resSS := c.sliceOfStructs([]struct {
		A int
		B string
	}{{1, "one"}, {5, "five"}, {10, "ten"}})
	if resSS != 16 {
		t.Fatalf("sliceOfStructs(): %d", resSS)
	}
}

func TestUint8SliceCasting(t *testing.T) {
	in := []uint8{4, 2, 6, 255, 0}
	buf := bytes.NewBuffer(nil)
	enc := irpcgen.NewEncoder(buf)
	enc.ByteSlice(in)
	enc.Flush()

	var out []uint8
	dec := irpcgen.NewDecoder(buf)
	dec.ByteSlice(&out)
	if !slices.Equal(in, out) {
		t.Fatalf("%v != %v", in, out)
	}
}
