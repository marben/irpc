package irpcgen

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// Encoder serializes given data type to byte stream.
// Encoder is meant to be used by generated code, not directly by the user.
type Encoder struct {
	w   *bufio.Writer
	buf []byte
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:   bufio.NewWriter(w),
		buf: make([]byte, binary.MaxVarintLen64),
	}
}

func (e *Encoder) Flush() error {
	return e.w.Flush()
}

func (e *Encoder) Bool(v bool) error {
	var b byte
	if v {
		b = 1
	}
	if err := e.w.WriteByte(b); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) VarInt(v int) error {
	return e.VarInt64(int64(v))
}

func (e *Encoder) UvarInt(v uint) error {
	return e.UvarInt64(uint64(v))
}

func (e *Encoder) Int8(v int8) error {
	return e.w.WriteByte(byte(v))
}

func (e *Encoder) Uint8(v uint8) error {
	return e.w.WriteByte(v)
}

func (e *Encoder) VarInt16(v int16) error {
	return e.VarInt64(int64(v))
}

func (e *Encoder) UvarInt16(v uint16) error {
	return e.UvarInt64(uint64(v))
}

func (e *Encoder) VarInt32(v int32) error {
	return e.VarInt64(int64(v))
}

func (e *Encoder) UvarInt32(v uint32) error {
	return e.UvarInt64(uint64(v))
}

func (e *Encoder) VarInt64(v int64) error {
	n := binary.PutVarint(e.buf, v)
	if _, err := e.w.Write(e.buf[:n]); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) UvarInt64(v uint64) error {
	n := binary.PutUvarint(e.buf, v)
	if _, err := e.w.Write(e.buf[:n]); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) Float32le(v float32) error {
	binary.LittleEndian.PutUint32(e.buf, math.Float32bits(v))
	if _, err := e.w.Write(e.buf[:4]); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) Float64le(v float64) error {
	binary.LittleEndian.PutUint64(e.buf, math.Float64bits(v))
	if _, err := e.w.Write(e.buf[:8]); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) ByteSlice(v []byte) error {
	if err := e.VarInt64(int64(len(v))); err != nil {
		return fmt.Errorf("slice len: %w", err)
	}
	if _, err := e.w.Write(v); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) String(v string) error {
	return e.ByteSlice([]byte(v))
}
