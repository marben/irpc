package irpcgen

import (
	"encoding"
	"fmt"
	"io"
	"math"
)

func EncBool[T ~bool](enc *Encoder, v T) error {
	return enc.bool(bool(v))
}

func DecBool[T ~bool](dec *Decoder, v *T) error {
	b, err := dec.bool()
	if err != nil {
		return err
	}
	*v = T(b)
	return nil
}

func EncUint8[T ~uint8](enc *Encoder, v T) error {
	return enc.byte(byte(v))
}

func DecUint8[T ~uint8](dec *Decoder, v *T) error {
	b, err := dec.byte()
	if err != nil {
		return err
	}
	*v = T(b)
	return nil
}

func EncInt[T ~int](enc *Encoder, v T) error {
	return enc.varInt64(int64(v))
}

func DecInt[T ~int](dec *Decoder, v *T) error {
	i64, err := dec.varInt64()
	if err != nil {
		return err
	}
	*v = T(i64)
	return nil
}

func EncUint[T ~uint](enc *Encoder, v T) error {
	return enc.uVarInt(uint(v))
}

func DecUint[T ~uint](dec *Decoder, v *T) error {
	ui64, err := dec.uVarInt64()
	if err != nil {
		return err
	}
	*v = T(ui64)
	return nil
}

func EncInt8[T ~int8](enc *Encoder, v T) error {
	return enc.byte(byte(v))
}

func DecInt8[T ~int8](dec *Decoder, v *T) error {
	b, err := dec.byte()
	if err != nil {
		return err
	}
	*v = T(b)
	return nil
}

func EncInt16[T ~int16](enc *Encoder, v T) error {
	return enc.varInt64(int64(v))
}

func DecInt16[T ~int16](dec *Decoder, v *T) error {
	val64, err := dec.varInt64InRange(math.MinInt16, math.MaxInt16)
	if err != nil {
		return err
	}

	*v = T(val64)
	return nil
}

func EncUint16[T ~uint16](enc *Encoder, v T) error {
	return enc.uVarInt64(uint64(v))
}

func DecUint16[T ~uint16](dec *Decoder, v *T) error {
	val64, err := dec.uvarInt64InRange(math.MaxUint16)
	if err != nil {
		return err
	}
	*v = T(val64)
	return nil
}

func EncInt32[T ~int32](enc *Encoder, v T) error {
	return enc.varInt64(int64(v))
}

func DecInt32[T ~int32](dec *Decoder, v *T) error {
	val64, err := dec.varInt64InRange(math.MinInt32, math.MaxInt32)
	if err != nil {
		return err
	}
	*v = T(val64)
	return nil
}

func EncUint32[T ~uint32](enc *Encoder, v T) error {
	return enc.uVarInt64(uint64(v))
}

func DecUint32[T ~uint32](dec *Decoder, v *T) error {
	val64, err := dec.uvarInt64InRange(math.MaxUint32)
	if err != nil {
		return err
	}
	*v = T(val64)
	return nil
}

func EncInt64[T ~int64](enc *Encoder, v T) error {
	return enc.varInt64(int64(v))
}

func DecInt64[T ~int64](dec *Decoder, v *T) error {
	i64, err := dec.varInt64()
	if err != nil {
		return err
	}
	*v = T(i64)
	return nil
}

func EncUint64[T ~uint64](enc *Encoder, v T) error {
	return enc.uVarInt64(uint64(v))
}

func DecUint64[T ~uint64](dec *Decoder, v *T) error {
	ui64, err := dec.uVarInt64()
	if err != nil {
		return err
	}
	*v = T(ui64)
	return nil
}

func EncFloat32[T ~float32](enc *Encoder, v T) error {
	return enc.float32le(float32(v))
}

func DecFloat32[T ~float32](dec *Decoder, v *T) error {
	f, err := dec.float32le()
	if err != nil {
		return err
	}
	*v = T(f)
	return nil
}

func EncFloat64[T ~float64](enc *Encoder, v T) error {
	return enc.float64le(float64(v))
}

func DecFloat64[T ~float64](dec *Decoder, v *T) error {
	f, err := dec.float64le()
	if err != nil {
		return err
	}
	*v = T(f)
	return nil
}

func EncString[T ~string](enc *Encoder, v T) error {
	return enc.string(string(v))
}

func DecString[T ~string](dec *Decoder, v *T) error {
	str, err := dec.string()
	if err != nil {
		return err
	}
	*v = T(str)
	return nil
}

func EncSlice[S ~[]E, E any](enc *Encoder, sl S, elemType string, elemEncFnc func(enc *Encoder, v E) error) error {
	if sl == nil {
		if err := enc.isNil(true); err != nil {
			return fmt.Errorf("serialize isNil: %w", err)
		}
		return nil
	}
	if err := enc.isNil(false); err != nil {
		return fmt.Errorf("serialize isNil: %w", err)
	}
	if err := enc.len(len(sl)); err != nil {
		return fmt.Errorf("serialize slice len: %w", err)
	}
	for _, e := range sl {
		if err := elemEncFnc(enc, e); err != nil {
			return fmt.Errorf("serialize element of type %q: %w", elemType, err)
		}
	}

	return nil
}

func DecSlice[S ~[]E, E any](dec *Decoder, sl *S, elemType string, elemDecFnc func(*Decoder, *E) error) error {
	isNil, err := dec.isNil()
	if err != nil {
		return fmt.Errorf("deserialize isNil: %w", err)
	}
	if isNil {
		sl = nil
		return nil
	}
	l, err := dec.len()
	if err != nil {
		return fmt.Errorf("deserialize slice len: %w", err)
	}
	lsl := make(S, l)
	for i := range l {
		if err := elemDecFnc(dec, &lsl[i]); err != nil {
			return fmt.Errorf("deserialize slice element of type %q: %w", elemType, err)
		}
	}
	*sl = lsl
	return nil
}

func EncMap[M ~map[K]V, K comparable, V any](enc *Encoder, m M, kType string, kEncFunc func(*Encoder, K) error, vType string, vEncFunc func(*Encoder, V) error) error {
	if m == nil {
		if err := enc.isNil(true); err != nil {
			return fmt.Errorf("serialize isNil: %w", err)
		}
		return nil
	}
	if err := enc.isNil(false); err != nil {
		return fmt.Errorf("serialize isNil: %w", err)
	}
	if err := enc.len(len(m)); err != nil {
		return fmt.Errorf("serialize map len: %w", err)
	}
	for k, v := range m {
		if err := kEncFunc(enc, k); err != nil {
			return fmt.Errorf("serialize key of type %q: %w", kType, err)
		}
		if err := vEncFunc(enc, v); err != nil {
			return fmt.Errorf("serialize value of type %q: %w", vType, err)
		}
	}
	return nil
}

func DecMap[M ~map[K]V, K comparable, V any](dec *Decoder, m *M, kName string, kDecFunc func(*Decoder, *K) error, vName string, vDecFunc func(*Decoder, *V) error) error {
	isNil, err := dec.isNil()
	if err != nil {
		return fmt.Errorf("deserialize isNil: %w", err)
	}
	if isNil {
		m = nil
		return nil
	}
	l, err := dec.len()
	if err != nil {
		return fmt.Errorf("deserialize map len: %w", err)
	}
	lm := make(M, l)
	for range l {
		var k K
		if err := kDecFunc(dec, &k); err != nil {
			return fmt.Errorf("deserialize map key: %w", err)
		}
		var v V
		if err := vDecFunc(dec, &v); err != nil {
			return fmt.Errorf("deserialize map val: %w", err)
		}
		lm[k] = v
	}
	*m = lm
	return nil
}

func EncByteSlice[T ~[]byte](enc *Encoder, v T) error {
	isNil := v == nil
	if err := enc.isNil(isNil); err != nil {
		return fmt.Errorf("serialize isNil: %w", err)
	}
	if isNil {
		return nil
	}

	if err := enc.len(len(v)); err != nil {
		return fmt.Errorf("slice len: %w", err)
	}
	if _, err := enc.w.Write(v); err != nil {
		return err
	}

	return nil
}

func DecByteSlice[T ~[]byte](dec *Decoder, v *T) error {
	isNil, err := dec.isNil()
	if err != nil {
		return fmt.Errorf("deserialize isNil: %w", err)
	}
	if isNil {
		v = nil
		return nil
	}

	l, err := dec.len()
	if err != nil {
		return fmt.Errorf("deserialize len: %w", err)
	}

	s := make([]byte, l)
	if _, err := io.ReadFull(dec.r, s); err != nil {
		return err
	}
	*v = s

	return nil
}

func EncBoolSlice[S ~[]bool](enc *Encoder, vs S) error {
	isNil := vs == nil
	if err := enc.isNil(isNil); err != nil {
		return fmt.Errorf("serialize isNil: %w", err)
	}
	if isNil {
		return nil
	}

	if err := enc.len(len(vs)); err != nil {
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
			if err := enc.w.WriteByte(b); err != nil {
				return err
			}
			b, bitCount = 0, 0
		}
	}

	if bitCount > 0 {
		// shift the last partial byte to align bits to the MSB
		b <<= uint(8 - bitCount)
		if err := enc.w.WriteByte(b); err != nil {
			return err
		}
	}

	return nil

}

func DecBoolSlice[T ~[]bool](dec *Decoder, v *T) error {
	isNil, err := dec.isNil()
	if err != nil {
		return fmt.Errorf("deserialize isNil: %w", err)
	}
	if isNil {
		v = nil
		return nil
	}

	l, err := dec.len()
	if err != nil {
		return fmt.Errorf("slice len: %w", err)
	}

	s := make([]bool, 0, l)

	for len(s) < l {
		b, err := dec.r.ReadByte()
		if err != nil {
			return err
		}

		// extract 8 bits MSB-first.
		for i := 0; i < 8 && len(s) < l; i++ {
			mask := byte(1 << (7 - i))
			s = append(s, b&mask != 0)
		}
	}
	*v = s
	return nil
}

func EncBinaryMarshaller(enc *Encoder, v encoding.BinaryMarshaler) error {
	data, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	return enc.byteSliceNonNil(data)
}

func DecBinaryUnmarshaller(dec *Decoder, dst encoding.BinaryUnmarshaler) error {
	data, err := dec.byteSliceNonNil()
	if err != nil {
		return err
	}
	return dst.UnmarshalBinary(data)
}

func EncIsNil(enc *Encoder, isNil bool) error {
	return enc.isNil(isNil)
}

func DecIsNil(dec *Decoder, isNil *bool) error {
	in, err := dec.isNil()
	if err != nil {
		return err
	}
	*isNil = in
	return nil
}
