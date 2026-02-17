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

// Flush writes any buffered data to the underlying io.Writer.
func (e *Encoder) Flush() error {
	return e.w.Flush()
}

func (e *Encoder) bool(v bool) error {
	var b byte
	if v {
		b = 1
	}
	if err := e.w.WriteByte(b); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) isNil(isNil bool) error {
	// we encode isNil as 0 and isNotNil(aka data follows) as 1
	// to be more consistent with general thinking about wire protocols
	return e.bool(!isNil)
}

func (e *Encoder) uVarInt(v uint) error {
	return e.uVarInt64(uint64(v))
}

func (e *Encoder) byte(b byte) error {
	return e.w.WriteByte(b)
}

func (e *Encoder) varInt64(v int64) error {
	n := binary.PutVarint(e.buf, v)
	if _, err := e.w.Write(e.buf[:n]); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) uVarInt64(v uint64) error {
	n := binary.PutUvarint(e.buf, v)
	if _, err := e.w.Write(e.buf[:n]); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) len(l int) error {
	return e.uVarInt64(uint64(l))
}

func (e *Encoder) float32le(v float32) error {
	binary.LittleEndian.PutUint32(e.buf, math.Float32bits(v))
	if _, err := e.w.Write(e.buf[:4]); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) float64le(v float64) error {
	binary.LittleEndian.PutUint64(e.buf, math.Float64bits(v))
	if _, err := e.w.Write(e.buf[:8]); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) byteSliceNonNil(v []byte) error {
	if err := e.len(len(v)); err != nil {
		return fmt.Errorf("slice len: %w", err)
	}
	if _, err := e.w.Write(v); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) string(v string) error {
	return e.byteSliceNonNil([]byte(v))
}
