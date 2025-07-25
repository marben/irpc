package irpcgen

import (
	"bufio"
	"encoding"
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// Decoder is a binary decoder that reads various data types from an io.Reader.
//
// It is meant to be used by generated code to decode messages in the IRPC protocol.
type Decoder struct {
	r   *bufio.Reader
	buf []byte
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:   bufio.NewReader(r),
		buf: make([]byte, binary.MaxVarintLen64),
	}
}

func (d *Decoder) Bool(dst *bool) error {
	val, err := d.r.ReadByte()
	if err != nil {
		return err
	}
	switch val {
	case 0:
		*dst = false
	case 1:
		*dst = true
	default:
		return fmt.Errorf("unexpected bool value: %d", val)
	}

	return nil
}

func (d *Decoder) VarInt(dst *int) error {
	val64, err := binary.ReadVarint(d.r)
	if err != nil {
		return err
	}

	*dst = int(val64)
	return nil
}

func (d *Decoder) UvarInt(dst *uint) error {
	val64, err := binary.ReadUvarint(d.r)
	if err != nil {
		return err
	}
	*dst = uint(val64)
	return nil
}

func (d *Decoder) Int8(dst *int8) error {
	b, err := d.r.ReadByte()
	if err != nil {
		return err
	}
	*dst = int8(b)
	return nil
}

func (d *Decoder) Uint8(dst *uint8) error {
	b, err := d.r.ReadByte()
	if err != nil {
		return err
	}
	*dst = b
	return nil
}

func (d *Decoder) varInt64InRange(min, max int64) (int64, error) {
	val, err := binary.ReadVarint(d.r)
	if err != nil {
		return 0, err
	}
	if val < min || val > max {
		return 0, fmt.Errorf("varint val %d is outside <%d, %d>", val, min, max)
	}
	return val, nil
}

func (d *Decoder) VarInt16(dst *int16) error {
	val64, err := d.varInt64InRange(math.MinInt16, math.MaxInt16)
	if err != nil {
		return err
	}
	*dst = int16(val64)
	return nil
}

func (d *Decoder) uvarInt64InRange(max uint64) (uint64, error) {
	val, err := binary.ReadUvarint(d.r)
	if err != nil {
		return 0, err
	}
	if val > max {
		return 0, fmt.Errorf("uvarint val %d is bigger than %d", val, max)
	}
	return val, nil
}

func (d *Decoder) UvarInt16(dst *uint16) error {
	val64, err := d.uvarInt64InRange(math.MaxUint16)
	if err != nil {
		return err
	}
	*dst = uint16(val64)
	return nil
}

func (d *Decoder) VarInt32(dst *int32) error {
	val64, err := d.varInt64InRange(math.MinInt32, math.MaxInt32)
	if err != nil {
		return err
	}
	*dst = int32(val64)
	return nil
}

func (d *Decoder) UvarInt32(dst *uint32) error {
	val64, err := d.uvarInt64InRange(math.MaxUint32)
	if err != nil {
		return err
	}
	*dst = uint32(val64)
	return nil
}

func (d *Decoder) VarInt64(dst *int64) error {
	val, err := binary.ReadVarint(d.r)
	if err != nil {
		return err
	}
	*dst = val
	return nil
}

func (d *Decoder) UvarInt64(dst *uint64) error {
	val, err := binary.ReadUvarint(d.r)
	if err != nil {
		return err
	}
	*dst = val
	return nil
}

func (d *Decoder) Float32le(dst *float32) error {
	if _, err := io.ReadFull(d.r, d.buf[:4]); err != nil {
		return err
	}
	*dst = math.Float32frombits(binary.LittleEndian.Uint32(d.buf[:4]))

	return nil
}

func (d *Decoder) Float64le(dst *float64) error {
	if _, err := io.ReadFull(d.r, d.buf[:8]); err != nil {
		return err
	}
	*dst = math.Float64frombits(binary.LittleEndian.Uint64(d.buf[:8]))

	return nil
}

func (d *Decoder) ByteSlice(dst *[]byte) error {
	var l int64
	if err := d.VarInt64(&l); err != nil {
		return fmt.Errorf("slice len: %w", err)
	}
	s := make([]byte, l)
	if _, err := io.ReadFull(d.r, s); err != nil {
		return err
	}
	*dst = s

	return nil
}

func (d *Decoder) String(dst *string) error {
	var bs []byte
	if err := d.ByteSlice(&bs); err != nil {
		return err
	}
	*dst = string(bs)
	return nil
}

func (d *Decoder) BinaryUnmarshaler(dst encoding.BinaryUnmarshaler) error {
	var data []byte
	if err := d.ByteSlice(&data); err != nil {
		return err
	}
	return dst.UnmarshalBinary(data)
}
