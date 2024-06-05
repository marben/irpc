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

const packetHeaderSize = 1 + 8 // packetType + DataLen

func (ph packetHeader) write(w io.Writer) error {
	b := make([]byte, packetHeaderSize)
	b[0] = byte(ph.typ)
	endian.PutUint64(b[1:], ph.dataLen)
	_, err := w.Write(b)
	return err
}

func (ph *packetHeader) read(r io.Reader) error {
	b := make([]byte, packetHeaderSize)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return err
	}

	ph.typ = packetType(b[0])
	ph.dataLen = endian.Uint64(b[1:])

	return nil
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
