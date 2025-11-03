package main

import (
	"fmt"
	"go/types"
)

type Type interface {
	encoder
	Name(q *qualifier) string
}

type namedInfo struct {
	namedName  string
	importSpec importSpec
}

func (ni namedInfo) qualifiedName(q *qualifier) string {
	return q.qualifyNamedInfo(ni)
}

// ast can be nil if not available
func (tr *typeResolver) newBasicType(bt *types.Basic, ni *namedInfo) (Type, error) {
	var irpcFuncName string
	switch bt.Kind() {
	case types.Bool:
		irpcFuncName = "Bool"
	case types.Int:
		irpcFuncName = "VarInt"
	case types.Uint:
		irpcFuncName = "UvarInt"
	case types.Int8:
		irpcFuncName = "Int8"
	case types.Uint8: // serves 'types.Byte' as well
		irpcFuncName = "Uint8"
	case types.Int16:
		irpcFuncName = "VarInt16"
	case types.Uint16:
		irpcFuncName = "UvarInt16"
	case types.Int32: // serves 'types.Rune' as well
		irpcFuncName = "VarInt32"
	case types.Uint32:
		irpcFuncName = "UvarInt32"
	case types.Int64:
		irpcFuncName = "VarInt64"
	case types.Uint64:
		irpcFuncName = "UvarInt64"
	case types.Float32:
		irpcFuncName = "Float32le"
	case types.Float64:
		irpcFuncName = "Float64le"
	case types.String:
		irpcFuncName = "String"
	default:
		return nil, fmt.Errorf("unsupported basic type %q", bt.Name())
	}

	return tr.newDirectCallType(irpcFuncName, irpcFuncName, bt.Name(), ni)
}

func (tr typeResolver) newBinaryMarshalerType(ni *namedInfo) (Type, error) {
	if ni == nil { // todo: perhaps possible? no time to test now
		return nil, fmt.Errorf("obtained binary marshaller with no named info. that is not supported atm")
	}
	return tr.newDirectCallType("BinaryMarshaler", "BinaryUnmarshaler", ni.namedName, ni)
}

// crappy name, but it's a type wrapping directCallEncoder
type directCallType struct {
	ni       *namedInfo
	typeName string
	enc      encoder
}

func (tr *typeResolver) newDirectCallType(encFunc, decFunc string, typeName string, ni *namedInfo) (Type, error) {
	return directCallType{
		ni:       ni,
		enc:      newDirectCallEncoder(encFunc, decFunc, typeName, ni), //newSymmetricDirectCallEncoder(encDecFunc, typeName, ni),
		typeName: typeName,
	}, nil
}

func (t directCallType) Name(q *qualifier) string {
	if t.ni == nil {
		return t.typeName
	}

	return q.qualifyNamedInfo(*t.ni)
}

func (t directCallType) codeblock(q *qualifier) string {
	return t.enc.codeblock(q)
}

func (t directCallType) decode(varId string, existingVars varNames, q *qualifier) string {
	return t.enc.decode(varId, existingVars, q)
}

func (t directCallType) encode(varId string, existingVars varNames, q *qualifier) string {
	return t.enc.encode(varId, existingVars, q)
}

// todo: use field for signatures too?
type field struct {
	name string
	t    Type
}
