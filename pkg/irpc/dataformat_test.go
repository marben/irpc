package irpc

import (
	"bytes"
	"testing"
)

func TestPacketHeaderSerializeDeserialize(t *testing.T) {
	phIn := packetHeader{
		typ: rpcRequest,
	}
	buf := bytes.NewBuffer(nil)
	if err := phIn.Serialize(buf); err != nil {
		t.Fatalf("serialize(): %v", err)
	}

	var phOut packetHeader
	if err := phOut.Deserialize(buf); err != nil {
		t.Fatalf("deserialize(): %v", err)
	}

	if phIn != phOut {
		t.Fatalf("in: %q != out: %q", phIn, phOut)
	}
}

func BenchmarkPacketHeaderSerialization(b *testing.B) {
	ph := packetHeader{
		typ: rpcRequest,
	}
	var length int
	buf := bytes.NewBuffer(make([]byte, 100))
	b.ResetTimer()
	for range b.N {
		buf.Reset()
		if err := ph.Serialize(buf); err != nil {
			b.Fatalf("serialization failed: %v", err)
		}
		length = buf.Len()
	}
	b.ReportMetric(float64(length), "Byte_len")
}

func BenchmarkPacketHeaderDeSerialization(b *testing.B) {
	phInit := packetHeader{
		typ: rpcRequest,
	}
	buf := bytes.NewBuffer(nil)
	if err := phInit.Serialize(buf); err != nil {
		b.Fatal("failed to serialize test packet")
	}
	data := buf.Bytes()
	b.ResetTimer()
	for range b.N {
		var phOut packetHeader
		buf2 := bytes.NewBuffer(data)
		if err := phOut.Deserialize(buf2); err != nil {
			b.Fatalf("deserialization failed: %v", err)
		}
		if phOut != phInit {
			b.Fatalf("%v != %v", phOut, phInit)
		}
	}
}
