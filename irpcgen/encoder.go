package irpcgen

import (
	"bufio"
	"encoding"
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

func (e *Encoder) byteSlice(v []byte) error {
	// todo: handle nil
	if err := e.len(len(v)); err != nil {
		return fmt.Errorf("slice len: %w", err)
	}
	if _, err := e.w.Write(v); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) string(v string) error {
	// todo: write own function without nil
	return e.byteSlice([]byte(v))
}

func (e *Encoder) binaryMarshaler(bm encoding.BinaryMarshaler) error {
	// todo: implement only in irpcgen.EncBinMarshaler ?
	data, err := bm.MarshalBinary()
	if err != nil {
		return err
	}
	return e.byteSlice(data)
}

func (e *Encoder) boolSlice(vs []bool) error {
	// todo: handle nil
	if err := e.len(len(vs)); err != nil {
		return fmt.Errorf("slice len: %w", err)
	}

	// MSB first
	var b byte
	bitCount := 0

	for _, v := range vs {
		b <<= 1
		if v {
			b |= 1
		}
		bitCount++

		if bitCount == 8 {
			if err := e.w.WriteByte(b); err != nil {
				return err
			}
			b, bitCount = 0, 0
		}
	}

	if bitCount > 0 {
		// shift the last partial byte to align bits to the MSB
		b <<= uint(8 - bitCount)
		if err := e.w.WriteByte(b); err != nil {
			return err
		}
	}

	return nil
}
