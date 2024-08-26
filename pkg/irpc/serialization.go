package irpc

// todo: this code should be generated

func (pt packetType) Serialize(e *Encoder) error    { return e.Uint8(uint8(pt)) }
func (pt *packetType) Deserialize(d *Decoder) error { return d.Uint8((*uint8)(pt)) }

func (rn ReqNumT) Serialize(e *Encoder) error    { return e.Uint16(uint16(rn)) }
func (rn *ReqNumT) Deserialize(d *Decoder) error { return d.Uint16((*uint16)(rn)) }

func (sid RegisteredServiceId) Serialize(e *Encoder) error    { return e.Uint16(uint16(sid)) }
func (sid *RegisteredServiceId) Deserialize(d *Decoder) error { return d.Uint16((*uint16)(sid)) }

func (fid FuncId) Serialize(e *Encoder) error    { return e.Uint16(uint16(fid)) }
func (fid *FuncId) Deserialize(d *Decoder) error { return d.Uint16((*uint16)(fid)) }

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

func (rp requestPacket) Serialize(e *Encoder) error {
	if err := rp.ReqNum.Serialize(e); err != nil {
		return err
	}
	if err := rp.ServiceId.Serialize(e); err != nil {
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
	if err := rp.ServiceId.Deserialize(d); err != nil {
		return err
	}
	if err := rp.FuncId.Deserialize(d); err != nil {
		return err
	}
	return nil
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
