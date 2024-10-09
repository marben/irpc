package irpc

// todo: serialization/deserialization code should be generated, not hand written

type packetType uint64

func (pt packetType) Serialize(e *Encoder) error    { return e.UvarInt64(uint64(pt)) }
func (pt *packetType) Deserialize(d *Decoder) error { return d.UvarInt64((*uint64)(pt)) }

type packetHeader struct {
	typ packetType
}

func (ph packetHeader) Serialize(e *Encoder) error {
	if err := ph.typ.Serialize(e); err != nil {
		return err
	}
	return nil
}

func (ph *packetHeader) Deserialize(d *Decoder) error {
	if err := ph.typ.Deserialize(d); err != nil {
		return err
	}
	return nil
}

type FuncId uint64

func (fid FuncId) Serialize(e *Encoder) error    { return e.UvarInt64(uint64(fid)) }
func (fid *FuncId) Deserialize(d *Decoder) error { return d.UvarInt64((*uint64)(fid)) }

type ReqNumT uint64

func (rn ReqNumT) Serialize(e *Encoder) error    { return e.UvarInt64(uint64(rn)) }
func (rn *ReqNumT) Deserialize(d *Decoder) error { return d.UvarInt64((*uint64)(rn)) }

type requestPacket struct {
	ReqNum    ReqNumT
	ServiceId []byte
	FuncId    FuncId
}

func (rp requestPacket) Serialize(e *Encoder) error {
	if err := rp.ReqNum.Serialize(e); err != nil {
		return err
	}
	if err := e.ByteSlice(rp.ServiceId); err != nil {
		return err
	}
	if err := rp.FuncId.Serialize(e); err != nil {
		return err
	}
	return nil
}

func (rp *requestPacket) Deserialize(d *Decoder) error {
	if err := rp.ReqNum.Deserialize(d); err != nil {
		return err
	}
	if err := d.ByteSlice(&rp.ServiceId); err != nil {
		return err
	}
	if err := rp.FuncId.Deserialize(d); err != nil {
		return err
	}
	return nil
}

type responsePacket struct {
	ReqNum ReqNumT // request number that initiated this response
}

func (rp responsePacket) Serialize(e *Encoder) error {
	if err := rp.ReqNum.Serialize(e); err != nil {
		return err
	}
	return nil
}

func (rp *responsePacket) Deserialize(d *Decoder) error {
	if err := rp.ReqNum.Deserialize(d); err != nil {
		return err
	}
	return nil
}

type ctxEndPacket struct {
	ReqNum ReqNumT
	ErrStr string
}

func (p ctxEndPacket) Serialize(e *Encoder) error {
	if err := p.ReqNum.Serialize(e); err != nil {
		return err
	}
	if err := e.String(p.ErrStr); err != nil {
		return err
	}
	return nil
}

func (p *ctxEndPacket) Deserialize(d *Decoder) error {
	if err := p.ReqNum.Deserialize(d); err != nil {
		return err
	}
	if err := d.String(&p.ErrStr); err != nil {
		return err
	}
	return nil
}
