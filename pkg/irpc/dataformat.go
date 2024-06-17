package irpc

import (
	"encoding/binary"
	"encoding/json"
	"io"
)

var endian = binary.LittleEndian

type packetType uint8

const (
	rpcRequest packetType = iota
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

func (rp *clientRegisterReq) Deserialize(r io.Reader) error {
	dec := json.NewDecoder(r)
	if err := dec.Decode(rp); err != nil {
		return err
	}

	return nil
}

func (rp clientRegisterReq) Serialize(w io.Writer) error {
	enc := json.NewEncoder(w)
	if err := enc.Encode(rp); err != nil {
		return err
	}
	return nil
}

type clientRegisterResp struct {
	ServiceId RegisteredServiceId
	Err       string // todo: turn into proper error, once we don't do json serialization
}

func (rp *clientRegisterResp) Deserialize(r io.Reader) error {
	dec := json.NewDecoder(r)
	if err := dec.Decode(rp); err != nil {
		return err
	}

	return nil
}

func (rp clientRegisterResp) Serialize(w io.Writer) error {
	enc := json.NewEncoder(w)
	if err := enc.Encode(rp); err != nil {
		return err
	}
	return nil
}

type requestPacket struct {
	ReqNum    uint16
	ServiceId RegisteredServiceId
	FuncId    FuncId
	Data      []byte
}

func (rp *requestPacket) Deserialize(r io.Reader) error {
	dec := json.NewDecoder(r)
	if err := dec.Decode(rp); err != nil {
		return err
	}

	return nil
}

func (rp requestPacket) Serialize(w io.Writer) error {
	enc := json.NewEncoder(w)
	if err := enc.Encode(rp); err != nil {
		return err
	}
	return nil
}

type responsePacket struct {
	ReqNum uint16 // number of the initiating request
	Data   []byte
	Err    string // not sure about this yet
}

func (rp responsePacket) Serialize(w io.Writer) error {
	enc := json.NewEncoder(w)
	if err := enc.Encode(rp); err != nil {
		return err
	}
	return nil
}

func (rp *responsePacket) Deserialize(r io.Reader) error {
	dec := json.NewDecoder(r)
	if err := dec.Decode(rp); err != nil {
		return err
	}

	return nil
}

func writeUint64(w io.Writer, data uint64) error {
	var buf [8]byte
	endian.PutUint64(buf[:], data)
	_, err := w.Write(buf[:])
	return err
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

func readUint64(r io.Reader) (uint64, error) {
	buf := make([]byte, 8)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	return endian.Uint64(buf), nil
}
