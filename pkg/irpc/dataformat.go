package irpc

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

// TODO: replace json with something better (gob for a starter?)

var endian = binary.LittleEndian

type packetType uint8

const (
	rpcRequest packetType = iota
	rpcResponse
)

type packetHeader struct {
	typ     packetType
	dataLen uint64
}

func (ph packetHeader) write(w io.Writer) error {
	if err := writeUint8(w, uint8(ph.typ)); err != nil {
		return err
	}
	if err := writeUint64(w, ph.dataLen); err != nil {
		return err
	}
	return nil
}

func (ph *packetHeader) read(r io.Reader) error {
	typUi8, err := readUint8(r)
	if err != nil {
		return err
	}
	ph.typ = packetType(typUi8)

	ph.dataLen, err = readUint64(r)

	return err
}

type requestPacket struct {
	ReqNum     uint16
	ServiceId  string
	FuncNameId string
	Data       []byte
}

func (rp *requestPacket) deserialize(data []byte) error {
	if err := json.Unmarshal(data, rp); err != nil {
		return fmt.Errorf("failed to deserialize packet. err: %w", err)
	}
	return nil
}

func (rp requestPacket) serialize() ([]byte, error) {
	return json.Marshal(rp)
}

type responsePacket struct {
	ReqNum uint16 // number of the initiating request
	Data   []byte
	Err    string // not sure about this yet
}

func (rp responsePacket) serialize() ([]byte, error) {
	return json.Marshal(rp)
}

func (rp *responsePacket) deserialize(data []byte) error {
	return json.Unmarshal(data, rp)
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
