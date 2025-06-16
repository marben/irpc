package irpctestpkg

import (
	"testing"

	"github.com/marben/irpc/cmd/irpc/test/testtools"
)

func TestStructParam(t *testing.T) {
	localEp, remoteEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create endpoints: %v", err)
	}

	skew := 8
	service := newStructAPIIRpcService(structImpl{skew: skew})
	remoteEp.RegisterService(service)

	c, err := newStructAPIIRpcClient(localEp)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	res := c.VectSum(vect3{1, 2, 3})
	if res != 1+2+3+skew {
		t.Fatalf("unexpected vector sum result: %d", res)
	}

	// Vect3x3
	res2 := c.Vect3x3Sum(vect3x3{
		v1: vect3{1, 2, 3},
		v2: vect3{3, 5, 6},
		v3: vect3{7, 8, 9},
	})

	exp2 := vect3{1 + 3 + 7 + skew, 2 + 5 + 8 + skew, 3 + 6 + 9 + skew}
	if res2 != exp2 {
		t.Fatalf("unexpected res2: %v", res2)
	}

	// sliceStruct
	res3 := c.SumSliceStruct(sliceStruct{
		s1: []int{1, 2, 3, 4, 5, 6},
		s2: []int{7, 8},
	})
	exp3 := 1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + skew
	if res3 != exp3 {
		t.Fatalf("unexpected res3: %v", res2)
	}
}
