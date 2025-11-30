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
	if err := EncBoolSlice(enc, s); err != nil {
		t.Fatalf("enc.BoolSlice(): %+v", err)
	}
	enc.Flush()
	t.Logf("slice: %v => %b", s, buf.Bytes())
	dec := NewDecoder(buf)
	var s_out []bool
	if err := DecBoolSlice(dec, &s_out); err != nil {
		t.Fatalf("dec.BoolSlice(): %v", err)
	}

	if !slices.Equal(s, s_out) {
		t.Fatalf("%v != %v", s, s_out)
	}
}

func FuzzBoolSlice(f *testing.F) {
	seeds := [][]byte{nil, {}, {3}, {5, 7, 8, 155, 255, 12}, {12, 14, 15, 15, 15, 16, 15, 16}}
	for _, tc := range seeds {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, s []byte) {
		buf := bytes.NewBuffer(nil)
		enc := NewEncoder(buf)
		dec := NewDecoder(buf)

		// we need []bool
		bs := bytesToBools(s)
		t.Logf("bools: %v", bs)

		if err := EncBoolSlice(enc, bs); err != nil {
			t.Fatalf("enc.BoolSlice(): %v", err)
		}
		if err := enc.Flush(); err != nil {
			t.Fatalf("enc.Flush(): %v", err)
		}

		// ok, decode
		var resBS []bool
		if err := DecBoolSlice(dec, &resBS); err != nil {
			t.Fatalf("dec.BoolSlice(): %v", err)
		}
		if !slices.Equal(resBS, bs) {
			t.Fatalf("encoded != decode: %v != %v", bs, resBS)
		}
	})
}

func bytesToBools(data []byte) []bool {
	if data == nil {
		return nil
	}

	out := make([]bool, len(data))
	for i, b := range data {
		out[i] = b&1 == 1
	}
	return out
}
