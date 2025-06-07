package irpctestpkg

import (
	"errors"
	"github.com/marben/irpc"
	"github.com/marben/irpc/cmd/irpc/test/testtools"
	"io"
	"math"
	"sync"
	"testing"
	"time"
)

func TestBasic(t *testing.T) {
	p1, p2 := testtools.NewDoubleEndedPipe()

	clientEp := irpc.NewEndpoint(p1)
	serviceEp := irpc.NewEndpoint(p2)

	skew := 2
	service := newBasicAPIIRpcService(basicApiImpl{skew: skew})
	serviceEp.RegisterServices(service)

	c, err := newBasicAPIIRpcClient(clientEp)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// BOOL
	negTrue := c.negBool(true)
	if negTrue != false {
		t.Fatalf("neg of 'true' failed")
	}
	negFalse := c.negBool(false)
	if negFalse != true {
		t.Fatalf("neg of 'false' failed")
	}

	// BYTE
	var ba, bb byte = math.MaxInt8, 1
	b := c.addByte(ba, bb)
	exb := ba + bb + byte(skew)
	if b != exb {
		t.Fatalf("unexpected byte: %d != %d", b, exb)
	}

	// INT
	int1 := c.addInt(1, 2)
	if int1 != 1+2+skew {
		t.Fatalf("unexpected result: %d", int1)
	}

	int2 := c.addInt(-5, -4)
	if int2 != -5-4+skew {
		t.Fatalf("unexpected result: %d", int2)
	}

	//  SWAP INT
	si1, si2 := c.swapInt(1, 2)
	if si1 != 2+skew {
		t.Fatalf("wrong swapped int si1: %d", si1)
	}
	if si2 != 1+skew {
		t.Fatalf("wrong swapped int si2: %d", si2)
	}

	// UINT
	var uia, uib uint = 1, 5
	ui := c.subUint(uia, uib)
	expectUi := uia - uib + uint(skew)

	if ui != expectUi {
		t.Fatalf("unexpected uint result: %d. expected %d", ui, expectUi)
	}

	// INT8
	var i8a, i8b int8 = -5, -2
	i8 := c.addInt8(i8a, i8b)
	exi8 := i8a + i8b + int8(skew)

	if i8 != exi8 {
		t.Fatalf("unexpected int8 result: %d. expected: %d", i8, exi8)
	}

	// UINT8
	var ui8a, ui8b uint8 = 255, 2
	ui8 := c.addUint8(ui8a, ui8b)
	expectUi8 := ui8a + ui8b + uint8(skew)

	if ui8 != expectUi8 {
		t.Fatalf("unexpected uin8 result: %d. expected %d", ui8, expectUi8)
	}

	// FLOAT64
	f64_1 := c.addFloat64(13.2, -53.919)
	expectFloat64 := float64(13.2) + float64(-53.919) + float64(skew)

	if f64_1 != expectFloat64 {
		t.Fatalf("unexpected float64 result: %f\n[%b]. expected: %f\n[%b]", f64_1, math.Float64bits(f64_1), expectFloat64, math.Float64bits(expectFloat64))
	}

	// FLOAT32
	f32 := c.addFloat32(3.987, -99.324)
	expectedF32 := float32(3.987) + float32(-99.324) + float32(skew)

	if f32 != expectedF32 {
		t.Fatalf("unexpected float32 result: %f\n[%b]. expected: %f\n[%b]", f32, math.Float32bits(f32), expectedF32, math.Float32bits(expectedF32))
	}

	// INT16
	var i16a, i16b int16 = math.MaxInt16, 1
	i16 := c.addInt16(i16a, i16b)
	exi16 := i16a + i16b + int16(skew)

	if i16 != exi16 {
		t.Fatalf("unexpected i16: %d != %d", i16, exi16)
	}

	// UINT16
	var ui16a, ui16b uint16 = math.MaxUint16, 1
	ui16 := c.addUint16(ui16a, ui16b)
	exui16 := ui16a + ui16b + uint16(skew)

	if ui16 != exui16 {
		t.Fatalf("unexpected ui16: %d != %d", ui16, exui16)
	}

	// UINT32
	var ui32a, ui32b uint32 = math.MaxUint32, 1
	ui32 := c.addUint32(ui32a, ui32b)
	exui32 := ui32a + ui32b + uint32(skew)

	if ui32 != exui32 {
		t.Fatalf("unexpected ui32: %d != %d", ui32, exui32)
	}

	// INT32
	var i32a, i32b int32 = math.MaxInt32, 1
	i32 := c.addInt32(i32a, i32b)
	exi32 := i32a + i32b + int32(skew)

	if i32 != exi32 {
		t.Fatalf("unexpected i32: %d != %d", i32, exi32)
	}

	// INT64
	var i64a, i64b int64 = math.MaxInt64, 1
	i64 := c.addInt64(i64a, i64b)
	exi64 := i64a + i64b + int64(skew)

	if i64 != exi64 {
		t.Fatalf("unexpected i64: %d != %d", i64, exi64)
	}

	// UINT64
	var ui64a, ui64b uint64 = math.MaxUint64, 1
	ui64 := c.addUint64(ui64a, ui64b)
	exui64 := ui64a + ui64b + uint64(skew)

	if ui64 != exui64 {
		t.Fatalf("unexpected ui64: %d != %d", ui64, exui64)
	}

	// RUNE
	r := c.toUpper('ř')

	if r != 'Ř' {
		t.Fatalf("unexpected toupper result: %c", r)
	}

	// STRING
	s := c.toUpperString("abcŘža")
	if s != "ABCŘŽA" {
		t.Fatalf("unepected toUpperString result: '%s'", s)
	}

	// close endpoints
	if err := serviceEp.Close(); err != nil {
		t.Fatalf("serviceEp.Close(): %+v", err)
	}
	time.Sleep(10 * time.Millisecond) // to make sure the closing notification has arrived
	if err := clientEp.Close(); !errors.Is(err, irpc.ErrEndpointClosedByCounterpart) {
		t.Fatalf("clientEp.Close(): %+v", err)
	}
}

type CountingReadWriteCloser struct {
	rwc    io.ReadWriteCloser
	rBytes int
	rmux   sync.Mutex
	wBytes int
	wmux   sync.Mutex
}

func (crw *CountingReadWriteCloser) Reset() {
	crw.rBytes, crw.wBytes = 0, 0
}

func (crw *CountingReadWriteCloser) Read(p []byte) (n int, err error) {
	n, err = crw.rwc.Read(p)

	crw.rmux.Lock()
	defer crw.rmux.Unlock()

	crw.rBytes += n
	return n, err
}

func (crw *CountingReadWriteCloser) Write(p []byte) (n int, err error) {
	n, err = crw.rwc.Write(p)

	crw.wmux.Lock()
	defer crw.wmux.Unlock()
	crw.wBytes += n
	return n, err
}

func (crw *CountingReadWriteCloser) RBytes() int {
	crw.rmux.Lock()
	defer crw.rmux.Unlock()
	return crw.rBytes
}

func (crw *CountingReadWriteCloser) WBytes() int {
	crw.wmux.Lock()
	defer crw.wmux.Unlock()
	return crw.wBytes
}

func (crw *CountingReadWriteCloser) Close() error {
	return crw.rwc.Close()
}
