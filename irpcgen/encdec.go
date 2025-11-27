package irpcgen

import (
	"encoding"
	"fmt"
)

func EncBool[T ~bool](enc *Encoder, v T) error {
	return enc.Bool(bool(v))
}

func DecBool[T ~bool](dec *Decoder, v *T) error {
	var x bool
	if err := dec.Bool(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncUint8[T ~uint8](enc *Encoder, v T) error {
	return enc.Uint8(uint8(v))
}

func DecUint8[T ~uint8](dec *Decoder, v *T) error {
	var x uint8
	if err := dec.Uint8(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncInt[T ~int](enc *Encoder, v T) error {
	return enc.VarInt(int(v))
}

func DecInt[T ~int](dec *Decoder, v *T) error {
	var x int
	if err := dec.VarInt(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncUint[T ~uint](enc *Encoder, v T) error {
	return enc.UvarInt(uint(v))
}

func DecUint[T ~uint](dec *Decoder, v *T) error {
	var x uint
	if err := dec.UvarInt(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncInt8[T ~int8](enc *Encoder, v T) error {
	return enc.Int8(int8(v))
}

func DecInt8[T ~int8](dec *Decoder, v *T) error {
	var x int8
	if err := dec.Int8(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncInt16[T ~int16](enc *Encoder, v T) error {
	return enc.VarInt16(int16(v))
}

func DecInt16[T ~int16](dec *Decoder, v *T) error {
	var x int16
	if err := dec.VarInt16(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncUint16[T ~uint16](enc *Encoder, v T) error {
	return enc.UvarInt16(uint16(v))
}

func DecUint16[T ~uint16](dec *Decoder, v *T) error {
	var x uint16
	if err := dec.UvarInt16(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncInt32[T ~int32](enc *Encoder, v T) error {
	return enc.VarInt32(int32(v))
}

func DecInt32[T ~int32](dec *Decoder, v *T) error {
	var x int32
	if err := dec.VarInt32(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncUint32[T ~uint32](enc *Encoder, v T) error {
	return enc.UvarInt32(uint32(v))
}

func DecUint32[T ~uint32](dec *Decoder, v *T) error {
	var x uint32
	if err := dec.UvarInt32(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncInt64[T ~int64](enc *Encoder, v T) error {
	return enc.VarInt64(int64(v))
}

func DecInt64[T ~int64](dec *Decoder, v *T) error {
	var x int64
	if err := dec.VarInt64(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncUint64[T ~uint64](enc *Encoder, v T) error {
	return enc.UvarInt64(uint64(v))
}

func DecUint64[T ~uint64](dec *Decoder, v *T) error {
	var x uint64
	if err := dec.UvarInt64(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncFloat32[T ~float32](enc *Encoder, v T) error {
	return enc.Float32le(float32(v))
}

func DecFloat32[T ~float32](dec *Decoder, v *T) error {
	var x float32
	if err := dec.Float32le(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncFloat64[T ~float64](enc *Encoder, v T) error {
	return enc.Float64le(float64(v))
}

func DecFloat64[T ~float64](dec *Decoder, v *T) error {
	var x float64
	if err := dec.Float64le(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncString[T ~string](enc *Encoder, v T) error {
	return enc.String(string(v))
}

func DecString[T ~string](dec *Decoder, v *T) error {
	var x string
	if err := dec.String(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

// todo: put slice parameter first!
func EncSlice[S ~[]E, E any](enc *Encoder, elemType string, elemEncFnc func(enc *Encoder, v E) error, sl S) error {
	// todo: handle nil slice
	if err := enc.Len(len(sl)); err != nil {
		return fmt.Errorf("serialize slice len: %w", err)
	}
	for _, e := range sl {
		if err := elemEncFnc(enc, e); err != nil {
			return fmt.Errorf("serialize element of type %q: %w", elemType, err)
		}
	}

	return nil
}

func DecSlice[S ~[]E, E any](dec *Decoder, elemType string, elemDecFnc func(*Decoder, *E) error, sl *S) error {
	var l int
	if err := dec.Len(&l); err != nil {
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
	// todo: handle nil maps
	if err := enc.Len(len(m)); err != nil {
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
	var l int
	if err := dec.Len(&l); err != nil {
		return fmt.Errorf("deserialize map len: %w", err)
	}
	lm := make(M, l)
	for range l {
		var k K
		if err := kDecFunc(dec, &k); err != nil {
			return fmt.Errorf("deserialzie map key: %w", err)
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
	// todo: handle nil slice
	return enc.ByteSlice(v)
}

func DecByteSlice[T ~[]byte](dec *Decoder, v *T) error {
	var x []byte
	if err := dec.ByteSlice(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncBoolSlice(enc *Encoder, v []bool) error {
	// todo: handle nil slice
	return enc.BoolSlice(v)
}

func DecBoolSlice[T ~[]bool](dec *Decoder, v *T) error {
	var bs []bool
	if err := dec.BoolSlice(&bs); err != nil {
		return err
	}
	*v = bs
	return nil
}

func EncLen(enc *Encoder, l int) error {
	return enc.Len(l)
}

func DecLen(dec *Decoder, l *int) error {
	return dec.Len(l)
}

func EncBinaryMarshaller(enc *Encoder, v encoding.BinaryMarshaler) error {
	return enc.BinaryMarshaler(v)
}

func DecBinaryUnmarshaller(dec *Decoder, dst encoding.BinaryUnmarshaler) error {
	return dec.BinaryUnmarshaler(dst)
}
