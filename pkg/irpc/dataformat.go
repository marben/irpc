package irpc

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
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

func (ph packetHeader) serialize() ([]byte, error) {
	// TODO: binary.Write uses reflection for structs.
	// TODO: serialize headers manually, to avoid (presumably slow) reflection
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, ph); err != nil {
		return nil, fmt.Errorf("failed to serialize packet header into binary form: %w", err)
	}

	return buf.Bytes(), nil
}

func (ph *packetHeader) deserialize(data []byte) error {
	buf := bytes.NewBuffer(data)
	return binary.Read(buf, binary.LittleEndian, &ph)
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
