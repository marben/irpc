package irpc

import (
	"encoding/binary"
	"io"
)

var endian = binary.LittleEndian

type packetType uint8

const (
	rpcRequest packetType = iota + 1
	rpcResponse
)

type packetHeader struct {
	typ packetType
}

func (ph packetHeader) Serialize(w io.Writer) error {
	if err := writeUint8(w, uint8(ph.typ)); err != nil {
		return err
	}

	return nil
}

func (ph *packetHeader) Deserialize(r io.Reader) error {
	typUi8, err := readUint8(r)
	if err != nil {
		return err
	}
	ph.typ = packetType(typUi8)

	return err
}

type clientRegisterReq struct {
	ServiceHash []byte
}

func (rp clientRegisterReq) Serialize(w io.Writer) error {
	if err := writeByteSlice(w, rp.ServiceHash); err != nil {
		return err
	}

	return nil
}

func (rp *clientRegisterReq) Deserialize(r io.Reader) error {
	var err error
	rp.ServiceHash, err = readByteSlice(r)
	if err != nil {
		return err
	}

	return nil
}

type clientRegisterResp struct {
	ServiceId RegisteredServiceId
	Err       string // todo: turn into proper error, once we don't do json serialization
}

func (rp clientRegisterResp) Serialize(w io.Writer) error {
	if err := writeUint16(w, uint16(rp.ServiceId)); err != nil {
		return err
	}
	if err := writeString(w, rp.Err); err != nil {
		return err
	}
	return nil
}

func (rp *clientRegisterResp) Deserialize(r io.Reader) error {
	var err error
	sid, err := readUint16(r)
	if err != nil {
		return err
	}
	rp.ServiceId = RegisteredServiceId(sid)

	rp.Err, err = readString(r)
	if err != nil {
		return err
	}

	return nil
}

type requestPacket struct {
	ReqNum    uint16
	ServiceId RegisteredServiceId
	FuncId    FuncId
}

func (rp requestPacket) Serialize(w io.Writer) error {
	if err := writeUint16(w, rp.ReqNum); err != nil {
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

func (rp *requestPacket) Deserialize(r io.Reader) error {
	var err error
	rp.ReqNum, err = readUint16(r)
	if err != nil {
		return err
	}
	sId, err := readUint16(r)
	if err != nil {
		return err
	}
	rp.ServiceId = RegisteredServiceId(sId)

	fId, err := readUint16(r)
	if err != nil {
		return err
	}
	rp.FuncId = FuncId(fId)
	return nil
}

type responsePacket struct {
	ReqNum uint16 // number of the initiating request
	Data   []byte
	Err    string // not sure about this yet
}

func (rp responsePacket) Serialize(w io.Writer) error {
	if err := writeUint16(w, rp.ReqNum); err != nil {
		return err
	}
	if err := writeByteSlice(w, rp.Data); err != nil {
		return err
	}
	if err := writeString(w, rp.Err); err != nil {
		return err
	}

	return nil
}

func (rp *responsePacket) Deserialize(r io.Reader) error {
	var err error
	rp.ReqNum, err = readUint16(r)
	if err != nil {
		return err
	}
	rp.Data, err = readByteSlice(r)
	if err != nil {
		return err
	}
	rp.Err, err = readString(r)
	if err != nil {
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

func readUint8(r io.Reader) (uint8, error) {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	return uint8(buf[0]), nil
}

func writeUint16(w io.Writer, data uint16) error {
	var buf [2]byte
	endian.PutUint16(buf[:], data)
	_, err := w.Write(buf[:])
	return err
}

func readUint16(r io.Reader) (uint16, error) {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	return endian.Uint16(buf), nil
}

func writeUint64(w io.Writer, data uint64) error {
	var buf [8]byte
	endian.PutUint64(buf[:], data)
	_, err := w.Write(buf[:])
	return err
}

func readUint64(r io.Reader) (uint64, error) {
	buf := make([]byte, 8)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	return endian.Uint64(buf), nil
}

func writeInt64(w io.Writer, data int64) error {
	return writeUint64(w, uint64(data))
}

func readInt64(r io.Reader) (int64, error) {
	ui64, err := readUint64(r)
	if err != nil {
		return 0, err
	}

	return int64(ui64), nil
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

func readByteSlice(r io.Reader) ([]byte, error) {
	l, err := readInt64(r)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, l)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	return buf, nil
}

func writeString(w io.Writer, s string) error {
	return writeByteSlice(w, []byte(s))
}

func readString(r io.Reader) (string, error) {
	b, err := readByteSlice(r)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
