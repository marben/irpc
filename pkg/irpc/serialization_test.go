package irpc

import (
	"bytes"
	"testing"
)

func TestPacketHeaderSerializeDeserialize(t *testing.T) {
	phIn := packetHeader{
		typ: rpcRequestPacket,
	}
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf)
	if err := phIn.Serialize(enc); err != nil {
		t.Fatalf("serialize(): %v", err)
	}

	var phOut packetHeader
	dec := NewDecoder(buf)
	if err := phOut.Deserialize(dec); err != nil {
		t.Fatalf("deserialize(): %v", err)
	}

	if phIn != phOut {
		t.Fatalf("in: %q != out: %q", phIn, phOut)
	}
}

func BenchmarkPacketHeaderSerialization(b *testing.B) {
	ph := packetHeader{
		typ: rpcRequestPacket,
	}
	var length int
	buf := bytes.NewBuffer(make([]byte, 100))
	b.ResetTimer()
	for range b.N {
		buf.Reset()
		enc := NewEncoder(buf)
		if err := ph.Serialize(enc); err != nil {
			b.Fatalf("serialization failed: %v", err)
		}
		length = buf.Len()
	}
	b.ReportMetric(float64(length), "Byte_len")
}

func BenchmarkPacketHeaderDeSerialization(b *testing.B) {
	phInit := packetHeader{
		typ: rpcRequestPacket,
	}
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf)
	if err := phInit.Serialize(enc); err != nil {
		b.Fatal("failed to serialize test packet")
	}
	data := buf.Bytes()
	b.ResetTimer()
	for range b.N {
		var phOut packetHeader
		buf2 := bytes.NewBuffer(data)
		dec := NewDecoder(buf2)
		if err := phOut.Deserialize(dec); err != nil {
			b.Fatalf("deserialization failed: %v", err)
		}
		if phOut != phInit {
			b.Fatalf("%v != %v", phOut, phInit)
		}
	}
}
