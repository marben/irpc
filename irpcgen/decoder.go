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

func (d *Decoder) bool(dst *bool) error {
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

func (d *Decoder) uVarInt(dst *uint) error {
	val64, err := binary.ReadUvarint(d.r)
	if err != nil {
		return err
	}
	*dst = uint(val64)
	return nil
}

func (d *Decoder) byte(dst *byte) error {
	v, err := d.r.ReadByte()
	if err != nil {
		return err
	}
	*dst = v
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

func (d *Decoder) varInt64(dst *int64) error {
	val, err := binary.ReadVarint(d.r)
	if err != nil {
		return err
	}
	*dst = val
	return nil
}

func (d *Decoder) uVarInt64(dst *uint64) error {
	val, err := binary.ReadUvarint(d.r)
	if err != nil {
		return err
	}
	*dst = val
	return nil
}

func (d *Decoder) len(l *int) error {
	var l64 uint64
	if err := d.uVarInt64(&l64); err != nil {
		return fmt.Errorf("slice len: %w", err)
	}
	*l = int(l64)
	return nil
}

func (d *Decoder) float32le(dst *float32) error {
	if _, err := io.ReadFull(d.r, d.buf[:4]); err != nil {
		return err
	}
	*dst = math.Float32frombits(binary.LittleEndian.Uint32(d.buf[:4]))

	return nil
}

func (d *Decoder) float64le(dst *float64) error {
	if _, err := io.ReadFull(d.r, d.buf[:8]); err != nil {
		return err
	}
	*dst = math.Float64frombits(binary.LittleEndian.Uint64(d.buf[:8]))

	return nil
}

func (d *Decoder) byteSlice(dst *[]byte) error {
	var l int
	if err := d.len(&l); err != nil {
		return fmt.Errorf("slice len: %w", err)
	}
	s := make([]byte, l)
	if _, err := io.ReadFull(d.r, s); err != nil {
		return err
	}
	*dst = s

	return nil
}

func (d *Decoder) string(dst *string) error {
	var bs []byte
	if err := d.byteSlice(&bs); err != nil {
		return err
	}
	*dst = string(bs)
	return nil
}

func (d *Decoder) binaryUnmarshaler(dst encoding.BinaryUnmarshaler) error {
	var data []byte
	if err := d.byteSlice(&data); err != nil {
		return err
	}
	return dst.UnmarshalBinary(data)
}

func (d *Decoder) boolSlice(dst *[]bool) error {
	var l int
	if err := d.len(&l); err != nil {
		return fmt.Errorf("slice len: %w", err)
	}

	s := make([]bool, 0, l)

	for len(s) < l {
		b, err := d.r.ReadByte()
		if err != nil {
			return err
		}

		// extract 8 bits MSB-first.
		for i := 0; i < 8 && len(s) < l; i++ {
			mask := byte(1 << (7 - i))
			s = append(s, b&mask != 0)
		}
	}
	*dst = s
	return nil
}
