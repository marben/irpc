package irpc

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

// TODO: replace json with something better (gob for a starter?)

type packetType uint8

const (
	rpcRequest packetType = iota
	rpcResponse
)

type packetHeader struct {
	Type    packetType
	DataLen uint64
}

func (ph packetHeader) write(w io.Writer) error {
	// TODO: binary.Write uses reflection for structs.
	// TODO: serialize headers manually, to avoid (presumably slow) reflection
	return binary.Write(w, binary.LittleEndian, ph)
}

func (ph *packetHeader) read(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, ph)
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
