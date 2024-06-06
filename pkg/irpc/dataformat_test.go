package irpc

import (
	"bytes"
	"reflect"
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

func TestRequestPacket(t *testing.T) {
	rpIn := requestPacket{
		ReqNum:     10,
		ServiceId:  "service 007",
		FuncNameId: "some func",
		Data:       []byte("this is the data"),
	}
	buf := bytes.NewBuffer(nil)
	if err := rpIn.Serialize(buf); err != nil {
		t.Fatal("failed to serialize test packet")
	}
	var rpOut requestPacket
	if err := rpOut.Deserialize(buf); err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if !reflect.DeepEqual(rpIn, rpOut) {
		t.Fatalf("rpIn != rpOut: %v != %v", rpIn, rpOut)
	}
}
