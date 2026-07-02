package irpc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/marben/irpc/irpcgen"
)

type dummyReadCloser struct {
	io.Reader
}

func (d dummyReadCloser) Close() error { return nil }

type dummyWriteCloser struct {
	io.Writer
}

func (d dummyWriteCloser) Close() error { return nil }

type dummyService struct{}

func (s *dummyService) Id() irpcgen.ServiceId {
	return irpcgen.ServiceId(123)
}

func (s *dummyService) GetFuncCall(funcId irpcgen.FuncId) (irpcgen.ArgDeserializer, error) {
	if funcId != 1 {
		return nil, fmt.Errorf("func not found")
	}
	return func(d *irpcgen.Decoder) (irpcgen.FuncExecutor, error) {
		var arg valPacket
		if err := arg.Deserialize(d); err != nil {
			return nil, err
		}
		return func(ctx context.Context) irpcgen.Serializable {
			return valPacket{val: arg.val * 2}
		}, nil
	}, nil
}

type valPacket struct {
	val uint32
}

func (p valPacket) Serialize(e *irpcgen.Encoder) error {
	return irpcgen.EncUint32(e, p.val)
}

func (p *valPacket) Deserialize(d *irpcgen.Decoder) error {
	return irpcgen.DecUint32(d, &p.val)
}

func TestSingleHandler_HandleOnce_Success(t *testing.T) {
	svc := &dummyService{}
	var buf bytes.Buffer
	enc := irpcgen.NewEncoder(&buf)

	// Serialize request header
	err := packetHeader{typ: rpcRequestPacketType}.Serialize(enc)
	if err != nil {
		t.Fatalf("Serialize header: %v", err)
	}

	// requestPacket
	reqPack := requestPacket{
		ReqNum:    1,
		ServiceId: svc.Id(),
		FuncId:    1,
	}
	if err := reqPack.Serialize(enc); err != nil {
		t.Fatalf("Serialize request packet: %v", err)
	}

	// arg (valPacket)
	arg := valPacket{val: 21}
	if err := arg.Serialize(enc); err != nil {
		t.Fatalf("Serialize arg: %v", err)
	}
	enc.Flush()

	var outBuf bytes.Buffer
	r := dummyReadCloser{Reader: &buf}
	w := dummyWriteCloser{Writer: &outBuf}
	handler := NewSingleHandler(r, w, svc)

	err = handler.HandleOnce()
	if err != nil {
		t.Fatalf("HandleOnce failed: %v", err)
	}

	// Verify response
	dec := irpcgen.NewDecoder(&outBuf)
	var h packetHeader
	if err := h.Deserialize(dec); err != nil {
		t.Fatalf("Deserialize response header: %v", err)
	}
	if h.typ != rpcResponsePacketType {
		t.Errorf("Unexpected header type: %v", h.typ)
	}

	var resp responsePacket
	if err := resp.Deserialize(dec); err != nil {
		t.Fatalf("Deserialize response packet: %v", err)
	}
	if resp.ReqNum != 1 {
		t.Errorf("Unexpected ReqNum: %v", resp.ReqNum)
	}

	var result valPacket
	if err := result.Deserialize(dec); err != nil {
		t.Fatalf("Deserialize result: %v", err)
	}
	if result.val != 42 {
		t.Errorf("Expected 42, got %v", result.val)
	}
}

func TestSingleHandler_HandleOnce_ClosingNow(t *testing.T) {
	var buf bytes.Buffer
	enc := irpcgen.NewEncoder(&buf)

	err := packetHeader{typ: closingNowPacketType}.Serialize(enc)
	if err != nil {
		t.Fatalf("Serialize header: %v", err)
	}
	enc.Flush()

	var outBuf bytes.Buffer
	r := dummyReadCloser{Reader: &buf}
	w := dummyWriteCloser{Writer: &outBuf}
	handler := NewSingleHandler(r, w)

	err = handler.HandleOnce()
	if !errors.Is(err, ErrClosingNow) {
		t.Errorf("Expected ErrClosingNow, got: %v", err)
	}
}

func TestSingleHandler_HandleOnce_CtxEnd(t *testing.T) {
	var buf bytes.Buffer
	enc := irpcgen.NewEncoder(&buf)

	err := packetHeader{typ: ctxEndPacketType}.Serialize(enc)
	if err != nil {
		t.Fatalf("Serialize header: %v", err)
	}

	ctxEnd := ctxEndPacket{
		ReqNum: 1,
		ErrStr: "canceled",
	}
	if err := ctxEnd.Serialize(enc); err != nil {
		t.Fatalf("Serialize ctxEnd: %v", err)
	}
	enc.Flush()

	var outBuf bytes.Buffer
	r := dummyReadCloser{Reader: &buf}
	w := dummyWriteCloser{Writer: &outBuf}
	handler := NewSingleHandler(r, w)

	err = handler.HandleOnce()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}
