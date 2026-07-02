package irpc

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/marben/irpc/irpcgen"
)

// ErrClosingNow is returned when a closing packet is received from the peer.
var ErrClosingNow = errors.New("irpc: closing now packet received")

type SingleHandler struct {
	dec         *irpcgen.Decoder
	enc         *irpcgen.Encoder
	servicesMap map[irpcgen.ServiceId]irpcgen.Service
}

func NewSingleHandler(r io.ReadCloser, w io.WriteCloser, services ...irpcgen.Service) *SingleHandler {
	servicesMap := make(map[irpcgen.ServiceId]irpcgen.Service)
	for _, s := range services {
		servicesMap[s.Id()] = s
	}
	return &SingleHandler{
		dec:         irpcgen.NewDecoder(r),
		enc:         irpcgen.NewEncoder(w),
		servicesMap: servicesMap,
	}
}

// HandleRequest reads one RPC request from r, executes it synchronously, and sends the response to w.
func (sh *SingleHandler) HandleOnce() error {
	var h packetHeader
	if err := h.Deserialize(sh.dec); err != nil {
		return err
	}

	switch h.typ {
	case rpcRequestPacketType:
		var req requestPacket
		if err := req.Deserialize(sh.dec); err != nil {
			return err
		}

		svc := sh.servicesMap[req.ServiceId]
		if svc == nil {
			return fmt.Errorf("irpc: service not found: %s", req.ServiceId)
		}

		argDeser, err := svc.GetFuncCall(req.FuncId)
		if err != nil {
			return err
		}

		funcExec, err := argDeser(sh.dec)
		if err != nil {
			return err
		}

		// Execute function synchronously in the current goroutine.
		// TinyGo might have limited goroutines/concurrency, so we run synchronously.
		resp := funcExec(context.Background())

		// Send Response
		respHeader := packetHeader{typ: rpcResponsePacketType}
		if err := respHeader.Serialize(sh.enc); err != nil {
			return err
		}

		respPack := responsePacket{ReqNum: req.ReqNum}
		if err := respPack.Serialize(sh.enc); err != nil {
			return err
		}

		if err := resp.Serialize(sh.enc); err != nil {
			return err
		}

		if err := sh.enc.Flush(); err != nil {
			return err
		}

		return nil

	case closingNowPacketType:
		return ErrClosingNow

	case ctxEndPacketType:
		// Just deserialize and ignore since we are synchronous and don't support cancellation
		var cancelRequest ctxEndPacket
		if err := cancelRequest.Deserialize(sh.dec); err != nil {
			return err
		}
		return nil

	default:
		return fmt.Errorf("irpc: unexpected packet type: %v", h.typ)
	}
}
