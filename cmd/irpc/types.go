package main

import (
	"fmt"
	"go/types"
)

type Type interface {
	genEncFunc(encoderVarName string, q *qualifier) string // todo: encVarName not necessary?
	genDecFunc(decoderVarName string, q *qualifier) string
	codeblocks(q *qualifier) []string // requested type's encode/decode code block at top level
	name(q *qualifier) string
}

// ast can be nil if not available
func (tr *typeResolver) newBasicType(bt *types.Basic, ni *namedInfo) (Type, error) {
	var encFunc, decFunc string
	switch bt.Kind() {
	case types.Bool:
		encFunc = "irpcgen.EncBool"
		decFunc = "irpcgen.DecBool"
	case types.Int:
		encFunc = "irpcgen.EncInt"
		decFunc = "irpcgen.DecInt"
	case types.Uint:
		encFunc = "irpcgen.EncUint"
		decFunc = "irpcgen.DecUint"
	case types.Int8:
		encFunc = "irpcgen.EncInt8"
		decFunc = "irpcgen.DecInt8"
	case types.Uint8: // serves 'types.Byte' as well
		encFunc = "irpcgen.EncUint8"
		decFunc = "irpcgen.DecUint8"
	case types.Int16:
		encFunc = "irpcgen.EncInt16"
		decFunc = "irpcgen.DecInt16"
	case types.Uint16:
		encFunc = "irpcgen.EncUint16"
		decFunc = "irpcgen.DecUint16"
	case types.Int32: // serves 'types.Rune' as well
		encFunc = "irpcgen.EncInt32"
		decFunc = "irpcgen.DecInt32"
	case types.Uint32:
		encFunc = "irpcgen.EncUint32"
		decFunc = "irpcgen.DecUint32"
	case types.Int64:
		encFunc = "irpcgen.EncInt64"
		decFunc = "irpcgen.DecInt64"
	case types.Uint64:
		encFunc = "irpcgen.EncUint64"
		decFunc = "irpcgen.DecUint64"
	case types.Float32:
		encFunc = "irpcgen.EncFloat32"
		decFunc = "irpcgen.DecFloat32"
	case types.Float64:
		encFunc = "irpcgen.EncFloat64"
		decFunc = "irpcgen.DecFloat64"
	case types.String:
		encFunc = "irpcgen.EncString"
		decFunc = "irpcgen.DecString"
	default:
		return nil, fmt.Errorf("unsupported basic type %q", bt.Name())
	}

	return newDirectCallType(encFunc, decFunc, bt.Name(), ni), nil
}

func (tr typeResolver) newBinaryMarshalerType(ni *namedInfo) (Type, error) {
	if ni == nil { // todo: perhaps possible? no time to test now
		return nil, fmt.Errorf("obtained binary marshaller with no named info. that is not supported atm")
	}
	return newDirectCallType("irpcgen.EncBinaryMarshaller", "irpcgen.DecBinaryUnmarshaller", ni.namedName, ni), nil
}

var _ Type = directCallType{}

type directCallType struct {
	ni               *namedInfo
	typeName         string
	encFunc, decFunc string
}

func newDirectCallType(encFunc, decFunc string, typeName string, ni *namedInfo) directCallType {
	return directCallType{
		ni:       ni,
		typeName: typeName,
		encFunc:  encFunc,
		decFunc:  decFunc,
	}
}

// genEncFunc implements Type.
func (t directCallType) genEncFunc(_ string, q *qualifier) string {
	return t.encFunc
}

// genDecFunc implements Type.
func (t directCallType) genDecFunc(_ string, q *qualifier) string {
	return t.decFunc
}

func (t directCallType) name(q *qualifier) string {
	if t.ni == nil {
		return t.typeName
	}

	return q.qualifyNamedInfo(*t.ni)
}

func (t directCallType) codeblocks(q *qualifier) []string {
	return nil
}

func newField(name string, t Type) field {
	return field{name, t}
}

// todo: use field for signatures too?
type field struct {
	name string
	t    Type
}
