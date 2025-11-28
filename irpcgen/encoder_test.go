package irpcgen

import (
	"bytes"
	"slices"
	"testing"
)

func BenchmarkVarUint64Enc(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 4))
	enc := NewEncoder(buf)
	encodedInt := uint64(1798453)
	for range b.N {
		if err := enc.uVarInt64(encodedInt); err != nil {
			b.Fatalf("enc: %v", err)
		}
		buf.Reset()
	}
}

func TestBoolSlice(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf)
	s := []bool{true, false, true, false, false}
	if err := enc.boolSlice(s); err != nil {
		t.Fatalf("enc.BoolSlice(): %+v", err)
	}
	enc.Flush()
	t.Logf("slice: %v => %b", s, buf.Bytes())
	dec := NewDecoder(buf)
	var s_out []bool
	if err := dec.boolSlice(&s_out); err != nil {
		t.Fatalf("dec.BoolSlice(): %v", err)
	}

	if !slices.Equal(s, s_out) {
		t.Fatalf("%v != %v", s, s_out)
	}
}

func FuzzBoolSlice(f *testing.F) {
	testcases := [][]byte{{3}, {5, 7, 8, 155, 255, 12}}
	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, s []byte) {
		buf := bytes.NewBuffer(nil)
		enc := NewEncoder(buf)
		dec := NewDecoder(buf)

		// we need []bool
		bs := bytesToBools(s)
		// t.Logf("bools: %v", bs)

		if err := enc.boolSlice(bs); err != nil {
			t.Fatalf("enc.BoolSlice(): %v", err)
		}
		if err := enc.Flush(); err != nil {
			t.Fatalf("enc.Flush(): %v", err)
		}

		// check the serialized data is same as the original input
		// we need to prefix the original data with correctly encoded length
		lenBuf := bytes.NewBuffer(nil)
		lenEnc := NewEncoder(lenBuf)
		lenEnc.len(len(bs))
		lenEnc.Flush()
		lenPrefixedOrig := append(lenBuf.Bytes(), s...)
		if !slices.Equal(buf.Bytes(), lenPrefixedOrig) {
			t.Fatalf("written data doesn't equal input: %v != %v", buf.Bytes(), lenPrefixedOrig)
		}

		// ok, decode
		var resBS []bool
		if err := dec.boolSlice(&resBS); err != nil {
			t.Fatalf("dec.BoolSlice(): %v", err)
		}
		if !slices.Equal(resBS, bs) {
			t.Fatalf("encoded != decode: %v != %v", bs, resBS)
		}
	})
}

func bytesToBools(data []byte) []bool {
	out := make([]bool, 0, len(data)*8)
	for _, b := range data {
		for i := 7; i >= 0; i-- { // MSB first
			out = append(out, (b&(1<<i)) != 0)
		}
	}
	return out
}
