package irpc

import (
	"bytes"
	"testing"
)

func BenchmarkUint64Enc(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 4))
	enc := NewEncoder(buf)
	encodedInt := uint64(1798453)
	for range b.N {
		if err := enc.Int64(int64(encodedInt)); err != nil {
			b.Fatalf("enc: %v", err)
		}
		buf.Reset()
	}
}

func BenchmarkVarUint64Enc(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 4))
	enc := NewEncoder(buf)
	encodedInt := uint64(1798453)
	for range b.N {
		if err := enc.UvarInt64(encodedInt); err != nil {
			b.Fatalf("enc: %v", err)
		}
		buf.Reset()
	}
}
