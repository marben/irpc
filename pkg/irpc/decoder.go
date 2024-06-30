package irpc

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

type Decoder struct {
	R      io.Reader // todo: make non public
	buf    []byte
	endian binary.ByteOrder
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		R:      r,
		buf:    make([]byte, 8),
		endian: binary.LittleEndian,
	}
}

func (d *Decoder) Bool(dst *bool) error {
	if _, err := io.ReadFull(d.R, d.buf[:1]); err != nil {
		return err
	}
	val := d.buf[0]
	if val == 0 {
		*dst = false
	} else if val == 1 {
		*dst = true
	} else {
		return fmt.Errorf("unexpected bool value: %d", val)
	}

	return nil
}

func (d *Decoder) Int(dst *int) error {
	if _, err := io.ReadFull(d.R, d.buf[:8]); err != nil {
		return err
	}

	*dst = int(d.endian.Uint64(d.buf[:8]))
	return nil
}

func (d *Decoder) Uint(dst *uint) error {
	if _, err := io.ReadFull(d.R, d.buf[:8]); err != nil {
		return err
	}

	*dst = uint(d.endian.Uint64(d.buf[:8]))
	return nil
}

func (d *Decoder) Int8(dst *int8) error {
	if _, err := io.ReadFull(d.R, d.buf[0:1]); err != nil {
		return err
	}
	*dst = int8(d.buf[0])

	return nil
}

func (d *Decoder) Uint8(dst *uint8) error {
	if _, err := io.ReadFull(d.R, d.buf[0:1]); err != nil {
		return err
	}
	*dst = uint8(d.buf[0])

	return nil
}

func (d *Decoder) Int16(dst *int16) error {
	if _, err := io.ReadFull(d.R, d.buf[:2]); err != nil {
		return err
	}

	*dst = int16(d.endian.Uint16(d.buf[:2]))
	return nil
}

func (d *Decoder) Uint16(dst *uint16) error {
	if _, err := io.ReadFull(d.R, d.buf[:2]); err != nil {
		return err
	}

	*dst = d.endian.Uint16(d.buf[:2])
	return nil
}

func (d *Decoder) Int32(dst *int32) error {
	if _, err := io.ReadFull(d.R, d.buf[:4]); err != nil {
		return err
	}

	*dst = int32(d.endian.Uint32(d.buf[:4]))
	return nil
}

func (d *Decoder) Uint32(dst *uint32) error {
	if _, err := io.ReadFull(d.R, d.buf[:4]); err != nil {
		return err
	}

	*dst = d.endian.Uint32(d.buf[:4])
	return nil
}

func (d *Decoder) Int64(dst *int64) error {
	if _, err := io.ReadFull(d.R, d.buf[:8]); err != nil {
		return err
	}

	*dst = int64(d.endian.Uint64(d.buf[:8]))
	return nil
}

func (d *Decoder) Uint64(dst *uint64) error {
	if _, err := io.ReadFull(d.R, d.buf[:8]); err != nil {
		return err
	}

	*dst = uint64(d.endian.Uint64(d.buf[:8]))
	return nil
}

func (d *Decoder) Float32(dst *float32) error {
	if _, err := io.ReadFull(d.R, d.buf[:4]); err != nil {
		return err
	}
	*dst = math.Float32frombits(d.endian.Uint32(d.buf[:4]))

	return nil
}

func (d *Decoder) Float64(dst *float64) error {
	if _, err := io.ReadFull(d.R, d.buf[:8]); err != nil {
		return err
	}
	*dst = math.Float64frombits(d.endian.Uint64(d.buf[:8]))

	return nil
}

// todo: make slices generic?
func (d *Decoder) ByteSlice(dst *[]byte) error {
	var l int64
	if err := d.Int64(&l); err != nil {
		return fmt.Errorf("slice len: %w", err)
	}
	*dst = make([]byte, l)
	if _, err := io.ReadFull(d.R, *dst); err != nil {
		return err
	}

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
