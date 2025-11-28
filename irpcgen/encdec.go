package irpcgen

import (
	"encoding"
	"fmt"
	"math"
)

func EncBool[T ~bool](enc *Encoder, v T) error {
	return enc.bool(bool(v))
}

func DecBool[T ~bool](dec *Decoder, v *T) error {
	var x bool
	if err := dec.bool(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncUint8[T ~uint8](enc *Encoder, v T) error {
	return enc.byte(byte(v))
}

func DecUint8[T ~uint8](dec *Decoder, v *T) error {
	var b byte
	if err := dec.byte(&b); err != nil {
		return err
	}
	*v = T(b)
	return nil
}

func EncInt[T ~int](enc *Encoder, v T) error {
	return enc.varInt64(int64(v))
}

func DecInt[T ~int](dec *Decoder, v *T) error {
	var x int64
	if err := dec.varInt64(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncUint[T ~uint](enc *Encoder, v T) error {
	return enc.uVarInt(uint(v))
}

func DecUint[T ~uint](dec *Decoder, v *T) error {
	var x uint
	if err := dec.uVarInt(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncInt8[T ~int8](enc *Encoder, v T) error {
	return enc.byte(byte(v))
}

func DecInt8[T ~int8](dec *Decoder, v *T) error {
	var x byte
	if err := dec.byte(&x); err != nil {
		return err
	}
	*v = T(x)
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
	var x int64
	if err := dec.varInt64(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncUint64[T ~uint64](enc *Encoder, v T) error {
	return enc.uVarInt64(uint64(v))
}

func DecUint64[T ~uint64](dec *Decoder, v *T) error {
	var x uint64
	if err := dec.uVarInt64(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncFloat32[T ~float32](enc *Encoder, v T) error {
	return enc.float32le(float32(v))
}

func DecFloat32[T ~float32](dec *Decoder, v *T) error {
	var x float32
	if err := dec.float32le(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncFloat64[T ~float64](enc *Encoder, v T) error {
	return enc.float64le(float64(v))
}

func DecFloat64[T ~float64](dec *Decoder, v *T) error {
	var x float64
	if err := dec.float64le(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncString[T ~string](enc *Encoder, v T) error {
	return enc.string(string(v))
}

func DecString[T ~string](dec *Decoder, v *T) error {
	var x string
	if err := dec.string(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

// todo: put slice parameter first!
func EncSlice[S ~[]E, E any](enc *Encoder, elemType string, elemEncFnc func(enc *Encoder, v E) error, sl S) error {
	// todo: handle nil slice
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

func DecSlice[S ~[]E, E any](dec *Decoder, elemType string, elemDecFnc func(*Decoder, *E) error, sl *S) error {
	var l int
	if err := dec.len(&l); err != nil {
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
		if err := enc.bool(true); err != nil {
			return fmt.Errorf("serialize isNil: %w", err)
		}
		return nil
	}
	if err := enc.bool(false); err != nil {
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
	var isNil bool
	if err := dec.bool(&isNil); err != nil {
		return fmt.Errorf("deserialize isNil: %w", err)
	}
	if isNil {
		m = nil
		return nil
	}
	var l int
	if err := dec.len(&l); err != nil {
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
	// todo: handle nil slice
	return enc.byteSlice(v)
}

func DecByteSlice[T ~[]byte](dec *Decoder, v *T) error {
	var x []byte
	if err := dec.byteSlice(&x); err != nil {
		return err
	}
	*v = T(x)
	return nil
}

func EncBoolSlice(enc *Encoder, v []bool) error {
	// todo: handle nil slice
	return enc.boolSlice(v)
}

func DecBoolSlice[T ~[]bool](dec *Decoder, v *T) error {
	var bs []bool
	if err := dec.boolSlice(&bs); err != nil {
		return err
	}
	*v = bs
	return nil
}

func EncLen(enc *Encoder, l int) error {
	return enc.len(l)
}

func DecLen(dec *Decoder, l *int) error {
	return dec.len(l)
}

func EncBinaryMarshaller(enc *Encoder, v encoding.BinaryMarshaler) error {
	return enc.binaryMarshaler(v)
}

func DecBinaryUnmarshaller(dec *Decoder, dst encoding.BinaryUnmarshaler) error {
	return dec.binaryUnmarshaler(dst)
}
