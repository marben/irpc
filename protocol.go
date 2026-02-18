package irpc

import "github.com/marben/irpc/irpcgen"

// todo: serialization/deserialization code should be generated, not hand written

const (
	rpcRequestPacketType packetType = iota + 1
	rpcResponsePacketType
	closingNowPacketType // informs peer that we will immediately close the connection
	ctxEndPacketType     // informs service runner that the provided function context expired
)

type packetType uint64

func (pt packetType) Serialize(e *irpcgen.Encoder) error {
	return irpcgen.EncUint64(e, uint64(pt))
}
func (pt *packetType) Deserialize(d *irpcgen.Decoder) error {
	return irpcgen.DecUint64(d, (*uint64)(pt))
}

type packetHeader struct {
	typ packetType
}

func (ph packetHeader) Serialize(e *irpcgen.Encoder) error {
	if err := ph.typ.Serialize(e); err != nil {
		return err
	}
	return nil
}

func (ph *packetHeader) Deserialize(d *irpcgen.Decoder) error {
	if err := ph.typ.Deserialize(d); err != nil {
		return err
	}
	return nil
}

// reqNumT is used to identify requests and responses.
// request numbers are re-used. generally, we only need as big enough number as is the number of parallel workers (DefaultParallelClientCalls)
type reqNumT uint16

func (rn reqNumT) Serialize(e *irpcgen.Encoder) error    { return irpcgen.EncUint16(e, uint16(rn)) }
func (rn *reqNumT) Deserialize(d *irpcgen.Decoder) error { return irpcgen.DecUint16(d, (*uint16)(rn)) }

type requestPacket struct {
	ReqNum    reqNumT
	ServiceId []byte
	FuncId    irpcgen.FuncId
}

func (rp requestPacket) Serialize(e *irpcgen.Encoder) error {
	if err := rp.ReqNum.Serialize(e); err != nil {
		return err
	}
	if err := irpcgen.EncByteSlice(e, rp.ServiceId); err != nil {
		return err
	}
	if err := irpcgen.EncUint64(e, uint64(rp.FuncId)); err != nil {
		return err
	}
	return nil
}

func (rp *requestPacket) Deserialize(d *irpcgen.Decoder) error {
	if err := rp.ReqNum.Deserialize(d); err != nil {
		return err
	}
	if err := irpcgen.DecByteSlice(d, &rp.ServiceId); err != nil {
		return err
	}
	if err := irpcgen.DecUint64(d, (*uint64)(&rp.FuncId)); err != nil {
		return err
	}
	return nil
}

type responsePacket struct {
	ReqNum reqNumT // request number that initiated this response
}

func (rp responsePacket) Serialize(e *irpcgen.Encoder) error {
	if err := rp.ReqNum.Serialize(e); err != nil {
		return err
	}
	return nil
}

func (rp *responsePacket) Deserialize(d *irpcgen.Decoder) error {
	if err := rp.ReqNum.Deserialize(d); err != nil {
		return err
	}
	return nil
}

type ctxEndPacket struct {
	ReqNum reqNumT
	ErrStr string
}

func (p ctxEndPacket) Serialize(e *irpcgen.Encoder) error {
	if err := p.ReqNum.Serialize(e); err != nil {
		return err
	}
	if err := irpcgen.EncString(e, p.ErrStr); err != nil {
		return err
	}
	return nil
}

func (p *ctxEndPacket) Deserialize(d *irpcgen.Decoder) error {
	if err := p.ReqNum.Deserialize(d); err != nil {
		return err
	}
	if err := irpcgen.DecString(d, &p.ErrStr); err != nil {
		return err
	}
	return nil
}
