package irpcgen

import (
	"bufio"
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

func (d *Decoder) bool() (bool, error) {
	val, err := d.r.ReadByte()
	if err != nil {
		return false, err
	}
	switch val {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("unexpected bool value: %d", val)
	}
}

func (d *Decoder) isNil() (bool, error) {
	isNotNil, err := d.bool()
	if err != nil {
		return false, err
	}

	return !isNotNil, nil
}

func (d *Decoder) byte() (byte, error) {
	return d.r.ReadByte()
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

func (d *Decoder) varInt64() (int64, error) {
	return binary.ReadVarint(d.r)
}

func (d *Decoder) uVarInt64() (uint64, error) {
	return binary.ReadUvarint(d.r)
}

func (d *Decoder) len() (int, error) {
	l64, err := d.uVarInt64()
	if err != nil {
		return 0, fmt.Errorf("slice len: %w", err)
	}
	return int(l64), nil
}

func (d *Decoder) float32le() (float32, error) {
	if _, err := io.ReadFull(d.r, d.buf[:4]); err != nil {
		return 0, err
	}
	f := math.Float32frombits(binary.LittleEndian.Uint32(d.buf[:4]))

	return f, nil
}

func (d *Decoder) float64le() (float64, error) {
	if _, err := io.ReadFull(d.r, d.buf[:8]); err != nil {
		return 0, err
	}
	f := math.Float64frombits(binary.LittleEndian.Uint64(d.buf[:8]))

	return f, nil
}

func (d *Decoder) byteSliceNonNil() ([]byte, error) {
	l, err := d.len()
	if err != nil {
		return nil, fmt.Errorf("slice len: %w", err)
	}
	s := make([]byte, l)
	if _, err := io.ReadFull(d.r, s); err != nil {
		return nil, err
	}

	return s, nil
}

func (d *Decoder) string() (string, error) {
	bs, err := d.byteSliceNonNil()
	if err != nil {
		return "", err
	}
	return string(bs), nil
}
