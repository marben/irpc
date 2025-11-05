package main

import (
	"fmt"
	"go/types"
	"strings"
)

type Type interface {
	encode(varId string, existingVars varNames, q *qualifier) string // inline variable encode
	decode(varId string, existingVars varNames, q *qualifier) string // inline variable decode
	codeblock(q *qualifier) string                                   // requested encoder's code block at top level
	name(q *qualifier) string
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

	return newDirectCallType(irpcFuncName, irpcFuncName, bt.Name(), ni), nil
}

func (tr typeResolver) newBinaryMarshalerType(ni *namedInfo) (Type, error) {
	if ni == nil { // todo: perhaps possible? no time to test now
		return nil, fmt.Errorf("obtained binary marshaller with no named info. that is not supported atm")
	}
	return newDirectCallType("BinaryMarshaler", "BinaryUnmarshaler", ni.namedName, ni), nil
}

// crappy name, but it's a type wrapping directCallEncoder
type directCallType struct {
	ni                       *namedInfo
	typeName                 string
	encFuncName, decFuncName string
	// enc                      encoder
}

func newDirectCallType(encFunc, decFunc string, typeName string, ni *namedInfo) Type {
	return directCallType{
		ni: ni,
		// enc:         newDirectCallEncoder(encFunc, decFunc, typeName, ni),
		typeName:    typeName,
		encFuncName: encFunc,
		decFuncName: decFunc,
	}
}

func (t directCallType) name(q *qualifier) string {
	if t.ni == nil {
		return t.typeName
	}

	return q.qualifyNamedInfo(*t.ni)
}

// analogous to direcCallType.Name(). will do for now, but eventually need to improve it
// it's "forComments", because it doesn't alter the qualifier since that would possibly introduce imports of unused(in workable code) packages
func (t directCallType) nameForComments(q *qualifier) string {
	if t.ni == nil {
		return t.typeName
	}

	// return e.ni.importSpec.pkgName + "." + e.ni.namedName

	return q.qualifyNamedInfoWithoutAddingImports(*t.ni)
}

func (t directCallType) codeblock(q *qualifier) string {
	return ""
}

func (t directCallType) decode(varId string, existingVars varNames, q *qualifier) string {
	var varParam string
	if t.needsCasting() {
		varParam = fmt.Sprintf("(*%s)(&%s)", t.typeName, varId)
	} else {
		varParam = "&" + varId
	}
	sb := &strings.Builder{}
	fmt.Fprintf(sb, `if err := d.%s(%s); err != nil{
		return fmt.Errorf("deserialize %s of type \"%s\": %%w",err)
	}
	`, t.decFuncName, varParam, varId, t.nameForComments(q))
	return sb.String()
}

func (t directCallType) encode(varId string, existingVars varNames, q *qualifier) string {
	var varParam string
	if t.needsCasting() {
		varParam = fmt.Sprintf("%s(%s)", t.typeName, varId)
	} else {
		varParam = varId
	}

	return fmt.Sprintf(`if err := e.%s(%s); err != nil {
		return fmt.Errorf("serialize %s of type \"%s\": %%w", err)
	}
	`, t.encFuncName, varParam, varId, t.nameForComments(q))
}

func (t directCallType) needsCasting() bool {
	if t.ni != nil && t.ni.namedName != t.typeName {
		return true
	}
	return false
}

// todo: use field for signatures too?
type field struct {
	name string
	t    Type
}
