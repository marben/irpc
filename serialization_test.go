package irpc

import (
	"bytes"
	"testing"
)

func TestPacketHeaderSerializeDeserialize(t *testing.T) {
	phIn := packetHeader{
		typ: rpcRequestPacketType,
	}
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf)
	if err := phIn.Serialize(enc); err != nil {
		t.Fatalf("serialize(): %v", err)
	}
	if err := enc.flush(); err != nil {
		t.Fatalf("enc.flush(): %v", err)
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
		typ: rpcRequestPacketType,
	}
	var length int
	buf := bytes.NewBuffer(make([]byte, 100))
	enc := NewEncoder(buf)
	b.ResetTimer()
	for range b.N {
		if err := ph.Serialize(enc); err != nil {
			b.Fatalf("serialization failed: %v", err)
		}
		length = enc.w.Buffered()
		buf.Reset()
		enc.w.Reset(buf)
	}
	b.ReportMetric(float64(length), "Byte_len")
}

func BenchmarkPacketHeaderDeSerialization(b *testing.B) {
	phInit := packetHeader{
		typ: rpcRequestPacketType,
	}
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf)
	if err := phInit.Serialize(enc); err != nil {
		b.Fatal("failed to serialize test packet")
	}
	enc.flush()
	data := buf.Bytes()

	dec := NewDecoder(bytes.NewReader(data))
	b.ResetTimer()
	for range b.N {
		var phOut packetHeader
		if err := phOut.Deserialize(dec); err != nil {
			b.Fatalf("deserialization failed: %v", err)
		}

		if phOut != phInit {
			b.Fatalf("%v != %v", phOut, phInit)
		}
		dec.r.Reset(bytes.NewReader(data))
	}
}
