package irpc

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

type Encoder struct {
	w      *bufio.Writer
	buf    []byte
	endian binary.ByteOrder
}

type Decoder struct {
	r      *bufio.Reader
	buf    []byte
	endian binary.ByteOrder
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:      bufio.NewWriter(w),
		buf:    make([]byte, 8),
		endian: binary.LittleEndian,
	}
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:      bufio.NewReader(r),
		buf:    make([]byte, 8),
		endian: binary.LittleEndian,
	}
}

func (e *Encoder) flush() error {
	return e.w.Flush()
}

func (e *Encoder) Bool(v bool) error {
	if v {
		e.buf[0] = 1
	} else {
		e.buf[0] = 0
	}
	if _, err := e.w.Write(e.buf[:1]); err != nil {
		return err
	}

	return nil
}

func (d *Decoder) Bool(dst *bool) error {
	if _, err := io.ReadFull(d.r, d.buf[:1]); err != nil {
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

func (e *Encoder) Int(v int) error {
	return e.Uint64(uint64(v))
}

func (d *Decoder) Int(dst *int) error {
	if _, err := io.ReadFull(d.r, d.buf[:8]); err != nil {
		return err
	}

	*dst = int(d.endian.Uint64(d.buf[:8]))
	return nil
}

func (e *Encoder) Uint(v uint) error {
	return e.Uint64(uint64(v))
}

func (d *Decoder) Uint(dst *uint) error {
	if _, err := io.ReadFull(d.r, d.buf[:8]); err != nil {
		return err
	}
	*dst = uint(d.endian.Uint64(d.buf[:8]))
	return nil
}

func (e *Encoder) Int8(v int8) error {
	return e.Uint8(uint8(v))
}

func (d *Decoder) Int8(dst *int8) error {
	if _, err := io.ReadFull(d.r, d.buf[0:1]); err != nil {
		return err
	}
	*dst = int8(d.buf[0])

	return nil
}

func (e *Encoder) Uint8(v uint8) error {
	e.buf[0] = byte(v)
	if _, err := e.w.Write(e.buf[:1]); err != nil {
		return err
	}
	return nil
}

func (d *Decoder) Uint8(dst *uint8) error {
	if _, err := io.ReadFull(d.r, d.buf[0:1]); err != nil {
		return err
	}
	*dst = uint8(d.buf[0])

	return nil
}

func (e *Encoder) Int16(v int16) error {
	return e.Uint16(uint16(v))
}

func (d *Decoder) Int16(dst *int16) error {
	if _, err := io.ReadFull(d.r, d.buf[:2]); err != nil {
		return err
	}

	*dst = int16(d.endian.Uint16(d.buf[:2]))
	return nil
}

func (e *Encoder) Uint16(v uint16) error {
	e.endian.PutUint16(e.buf, v)
	if _, err := e.w.Write(e.buf[:2]); err != nil {
		return err
	}
	return nil
}

func (d *Decoder) Uint16(dst *uint16) error {
	if _, err := io.ReadFull(d.r, d.buf[:2]); err != nil {
		return err
	}

	*dst = d.endian.Uint16(d.buf[:2])
	return nil
}

func (e *Encoder) Int32(v int32) error {
	return e.Uint32(uint32(v))
}

func (d *Decoder) Int32(dst *int32) error {
	if _, err := io.ReadFull(d.r, d.buf[:4]); err != nil {
		return err
	}

	*dst = int32(d.endian.Uint32(d.buf[:4]))
	return nil
}

func (e *Encoder) Uint32(v uint32) error {
	e.endian.PutUint32(e.buf, v)
	if _, err := e.w.Write(e.buf[:4]); err != nil {
		return err
	}
	return nil
}

func (d *Decoder) Uint32(dst *uint32) error {
	if _, err := io.ReadFull(d.r, d.buf[:4]); err != nil {
		return err
	}

	*dst = d.endian.Uint32(d.buf[:4])
	return nil
}

func (e *Encoder) Int64(v int64) error {
	return e.Uint64(uint64(v))
}

func (d *Decoder) Int64(dst *int64) error {
	if _, err := io.ReadFull(d.r, d.buf[:8]); err != nil {
		return err
	}

	*dst = int64(d.endian.Uint64(d.buf[:8]))
	return nil
}

func (e *Encoder) Uint64(v uint64) error {
	e.endian.PutUint64(e.buf, v)
	if _, err := e.w.Write(e.buf[:8]); err != nil {
		return err
	}
	return nil
}

func (d *Decoder) Uint64(dst *uint64) error {
	if _, err := io.ReadFull(d.r, d.buf[:8]); err != nil {
		return err
	}

	*dst = uint64(d.endian.Uint64(d.buf[:8]))
	return nil
}

func (e *Encoder) UvarInt64(v uint64) error {
	binary.PutUvarint(e.buf, v)
	if _, err := e.w.Write(e.buf[:8]); err != nil {
		return err
	}
	return nil
}

// func (d *Decoder) UvarInt64(dst *uint64) error {
// 	// if _, err := io.ReadFull(d.R, d.buf[:8]); err != nil {
// 	// 	return err
// 	// }
// 	binary.ReadUvarint(d.R)

// 	*dst = uint64(d.endian.Uint64(d.buf[:8]))
// 	return nil
// }

func (e *Encoder) Float32(v float32) error {
	return e.Uint32(math.Float32bits(v))
}

func (d *Decoder) Float32(dst *float32) error {
	if _, err := io.ReadFull(d.r, d.buf[:4]); err != nil {
		return err
	}
	*dst = math.Float32frombits(d.endian.Uint32(d.buf[:4]))

	return nil
}

func (e *Encoder) Float64(v float64) error {
	return e.Uint64(math.Float64bits(v))
}

func (d *Decoder) Float64(dst *float64) error {
	if _, err := io.ReadFull(d.r, d.buf[:8]); err != nil {
		return err
	}
	*dst = math.Float64frombits(d.endian.Uint64(d.buf[:8]))

	return nil
}

func (e *Encoder) ByteSlice(v []byte) error {
	if err := e.Int64(int64(len(v))); err != nil {
		return fmt.Errorf("slice len: %w", err)
	}
	if _, err := e.w.Write(v); err != nil {
		return err
	}
	return nil
}

// we don't do generic slices as it generally adds time and allocations
func (d *Decoder) ByteSlice(dst *[]byte) error {
	var l int64
	if err := d.Int64(&l); err != nil {
		return fmt.Errorf("slice len: %w", err)
	}
	*dst = make([]byte, l)
	if _, err := io.ReadFull(d.r, *dst); err != nil {
		return err
	}

	return nil
}

func (e *Encoder) String(v string) error {
	return e.ByteSlice([]byte(v))
}

func (d *Decoder) String(dst *string) error {
	var bs []byte
	if err := d.ByteSlice(&bs); err != nil {
		return err
	}
	*dst = string(bs)
	return nil
}
