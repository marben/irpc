package irpctestpkg

import (
	"testing"

	"github.com/marben/irpc/cmd/irpc/test/testtools"
)

func TestError(t *testing.T) {
	localEp, remoteEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create endpoints: %v", err)
	}

	remoteEp.RegisterService(newInterfaceTestIRpcService(interfaceTestImpl{}))

	c, err := newInterfaceTestIRpcClient(localEp)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	res1 := c.rtnErrorWithMessage("le message")
	if res1 == nil {
		t.Fatalf("returned error is nil")
	}
	if res1.Error() != "le message" {
		t.Fatalf("unexpected message inside error: %s", res1.Error())
	}

	res2 := c.rtnNilError()
	if res2 != nil {
		t.Fatalf("unexpected non-nil error: %v", res2)
	}

	res3_string, res3_error := c.rtnStringAndError("messaaage")
	if res3_string != "messaaage" {
		t.Fatalf("unexpected string: %s", res3_string)
	}
	if res3_error == nil {
		t.Fatalf("unexpected nil, but expected error")
	}
	if res3_error.Error() != "messaaage" {
		t.Fatalf("unexpected error: %v", res3_error)
	}

	// return two errors
	res4_1, res4_2 := c.rtnTwoErrors()
	if res4_1 != nil {
		t.Fatalf("unexpected first error: %v", res4_1)
	}
	if res4_2 == nil || res4_2.Error() != "err2" {
		t.Fatalf("unexpected second error: %v", res4_2)
	}

	res5_val, res5_err := c.passCustomInterfaceAndReturnItModified(customInterfaceImpl{
		i: 7,
		s: "hello",
	})
	if res5_err != nil {
		t.Fatalf("unexpected error: %+v", res5_err)
	}
	if res5_val == nil || res5_val.IntFunc() != 7+1 || res5_val.StringFunc() != "hello"+"_modified" {
		t.Fatalf("unexpected value: %+v", res5_val)
	}

	// pass empty interface
	res6_val, res6_err := c.passCustomInterfaceAndReturnItModified(nil)
	if res6_val != nil {
		t.Fatalf("expected nil return, got: %+v", res6_val)
	}
	if res6_err == nil || res6_err.Error() != "nil pointer" {
		t.Fatalf("expected 'nil pointer' error, but got: %+v", res6_err)
	}
}
