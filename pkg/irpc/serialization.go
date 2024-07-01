package irpc

func (ph packetHeader) Serialize(e *Encoder) error {
	if err := e.Uint8(uint8(ph.typ)); err != nil {
		return err
	}

	return nil
}

func (ph *packetHeader) Deserialize(d *Decoder) error {
	if err := d.Uint8((*uint8)(&ph.typ)); err != nil {
		return err
	}
	return nil
}

func (rp requestPacket) Serialize(e *Encoder) error {
	if err := e.Uint16(uint16(rp.ReqNum)); err != nil {
		return err
	}
	if err := e.Uint16(uint16(rp.ServiceId)); err != nil {
		return err
	}
	if err := e.Uint16(uint16(rp.FuncId)); err != nil {
		return err
	}

	return nil
}

func (rp *requestPacket) Deserialize(d *Decoder) error {
	if err := d.Uint16((*uint16)(&rp.ReqNum)); err != nil {
		return err
	}
	if err := d.Uint16((*uint16)(&rp.ServiceId)); err != nil {
		return err
	}

	if err := d.Uint16((*uint16)(&rp.FuncId)); err != nil {
		return err
	}
	return nil
}

func (rp responsePacket) Serialize(e *Encoder) error {
	if err := e.Uint16(uint16(rp.ReqNum)); err != nil {
		return err
	}
	return nil
}

func (rp *responsePacket) Deserialize(d *Decoder) error {
	if err := d.Uint16((*uint16)(&rp.ReqNum)); err != nil {
		return err
	}

	return nil
}
