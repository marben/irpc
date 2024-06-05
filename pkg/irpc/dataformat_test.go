package irpc

import (
	"bytes"
	"testing"
)

func TestPacketHeaderSerializeDeserialize(t *testing.T) {
	phIn := packetHeader{
		Type:    rpcRequest,
		DataLen: 144,
	}
	buf := bytes.NewBuffer(nil)
	if err := phIn.write(buf); err != nil {
		t.Fatalf("serialize(): %v", err)
	}

	var phOut packetHeader
	if err := phOut.read(buf); err != nil {
		t.Fatalf("deserialize(): %v", err)
	}

	if phIn != phOut {
		t.Fatalf("in: %q != out: %q", phIn, phOut)
	}
}

func BenchmarkPacketHeaderSerialization(b *testing.B) {
	ph := packetHeader{
		Type:    rpcRequest,
		DataLen: 0,
	}
	var length int
	buf := bytes.NewBuffer(make([]byte, 100))
	for range b.N {
		buf.Reset()
		if err := ph.write(buf); err != nil {
			b.Fatalf("serialization failed: %v", err)
		}
		length = buf.Len()
	}
	b.ReportMetric(float64(length), "Byte_len")
}
