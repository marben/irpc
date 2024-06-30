package irpc

import (
	"io"
)

func (ph packetHeader) Serialize(w io.Writer) error {
	if err := writeUint8(w, uint8(ph.typ)); err != nil {
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

func (rp requestPacket) Serialize(w io.Writer) error {
	if err := writeUint16(w, uint16(rp.ReqNum)); err != nil {
		return err
	}
	if err := writeUint16(w, uint16(rp.ServiceId)); err != nil {
		return err
	}
	if err := writeUint16(w, uint16(rp.FuncId)); err != nil {
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

func (rp responsePacket) Serialize(w io.Writer) error {
	if err := writeUint16(w, uint16(rp.ReqNum)); err != nil {
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

func writeUint8(w io.Writer, data uint8) error {
	var buf [1]byte
	buf[0] = data
	_, err := w.Write(buf[:])
	return err
}

func writeUint16(w io.Writer, data uint16) error {
	var buf [2]byte
	endian.PutUint16(buf[:], data)
	_, err := w.Write(buf[:])
	return err
}

func writeUint64(w io.Writer, data uint64) error {
	var buf [8]byte
	endian.PutUint64(buf[:], data)
	_, err := w.Write(buf[:])
	return err
}

func writeInt64(w io.Writer, data int64) error {
	return writeUint64(w, uint64(data))
}

func writeByteSlice(w io.Writer, data []byte) error {
	l := len(data)
	if err := writeInt64(w, int64(l)); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}

func writeString(w io.Writer, s string) error {
	return writeByteSlice(w, []byte(s))
}
