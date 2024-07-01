package main

import (
	"bytes"
	"fmt"
	"go/types"
	"strings"
)

type encoder interface {
	encode(varId string) string // inline variable encode
	decode(varId string) string // inline variable decode
	imports() []string          // necessary imports
	codeblock() string          // requested encoder's code block at top level
}

func varEncoder(apiName string, t types.Type) (encoder, error) {
	switch t := t.(type) {
	case *types.Basic:
		switch t.Kind() {
		case types.Bool:
			return boolEncoder, nil
		case types.Int:
			return intEncoder, nil
		case types.Uint:
			return uintEncoder, nil
		case types.Int8:
			return int8Encoder, nil
		case types.Uint8: // serves 'types.Byte' as well
			return uint8Encoder, nil
		case types.Int16:
			return int16Encoder, nil
		case types.Uint16:
			return uint16Encoder, nil
		case types.Int32: // serves 'types.Rune' as well
			return int32Encoder, nil
		case types.Uint32:
			return uint32Encoder, nil
		case types.Int64:
			return int64Encoder, nil
		case types.Uint64:
			return uint64Encoder, nil
		case types.Float32:
			return float32Encoder, nil
		case types.Float64:
			return float64Encoder, nil
		case types.String:
			return newStringEncoder(), nil
		default:
			return nil, fmt.Errorf("unsupported basic type '%s'", t.Name())
		}
	case *types.Slice:
		return newSliceEncoder(apiName, t)
	case *types.Named:
		name := t.Obj().Name()
		switch ut := t.Underlying().(type) {
		case *types.Struct:
			return newNamedStructEncoder(apiName, ut)
		case *types.Interface:
			return newInterfaceEncoder(name, apiName, ut)
		default:
			return nil, fmt.Errorf("unsupported named type: %s", ut)
		}
	default:
		return nil, fmt.Errorf("unsupported type '%T' of %s ", t, t)
	}

}

var (
	boolEncoder = primitiveTypeEncoder{ // todo: could use bitpacking instead of one bool per byte
		decFuncName: "Bool",
		typeName:    "bool",
	}
	intEncoder = primitiveTypeEncoder{
		decFuncName: "Int",
		typeName:    "int",
	}
	uintEncoder = primitiveTypeEncoder{
		decFuncName: "Uint",
		typeName:    "uint",
	}
	int8Encoder = primitiveTypeEncoder{
		decFuncName: "Int8",
		typeName:    "int8",
	}
	uint8Encoder = primitiveTypeEncoder{
		decFuncName: "Uint8",
		typeName:    "uint8",
	}
	int16Encoder = primitiveTypeEncoder{
		decFuncName: "Int16",
		typeName:    "int16",
	}
	uint16Encoder = primitiveTypeEncoder{
		decFuncName: "Uint16",
		typeName:    "uint16",
	}
	int32Encoder = primitiveTypeEncoder{
		decFuncName: "Int32",
		typeName:    "int32",
	}
	uint32Encoder = primitiveTypeEncoder{
		decFuncName: "Uint32",
		typeName:    "uint32",
	}
	int64Encoder = primitiveTypeEncoder{
		decFuncName: "Int64",
		typeName:    "int64",
	}
	uint64Encoder = primitiveTypeEncoder{
		decFuncName: "Uint64",
		typeName:    "uint64",
	}
	float32Encoder = primitiveTypeEncoder{
		decFuncName: "Float32",
		typeName:    "float32",
	}
	float64Encoder = primitiveTypeEncoder{
		decFuncName: "Float64",
		typeName:    "float64",
	}
)

type primitiveTypeEncoder struct {
	decFuncName string
	typeName    string
}

func (e primitiveTypeEncoder) encode(varId string) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, `if err := e.%s(%s); err != nil {
		return fmt.Errorf("serialize %s of type '%s': %%w", err)
	}
	`, e.decFuncName, varId, varId, e.typeName)

	return sb.String()
}

func (e primitiveTypeEncoder) decode(varId string) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, `if err := d.%s(&%s); err != nil{
		return fmt.Errorf("deserialize %s of type '%s': %%w",err)
	}
	`, e.decFuncName, varId, varId, e.typeName)
	return sb.String()
}

func (e primitiveTypeEncoder) imports() []string {
	return []string{irpcImport, fmtImport}
}

func (e primitiveTypeEncoder) codeblock() string {
	return ""
}

// todo: perhaps rewrite using slice encoder?
type stringEncoder struct {
	lenEncoder encoder
}

func newStringEncoder() stringEncoder {
	return stringEncoder{intEncoder}
}

func (e stringEncoder) encode(varId string) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, `if err := e.String(%s); err != nil {
		return err
	}
	`, varId)

	return sb.String()
}

func (e stringEncoder) decode(varId string) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, `if err := d.String(&%s); err != nil{
		return fmt.Errorf("deserialize %s of type 'string': %%w",err)
	}
	`, varId, varId)
	return sb.String()
}

func (e stringEncoder) imports() []string {
	return append(e.lenEncoder.imports(), fmtImport)
}

func (e stringEncoder) codeblock() string {
	return ""
}

type sliceEncoder struct {
	elemEnc  encoder
	lenEnc   encoder
	elemType types.Type
}

func newSliceEncoder(apiName string, t *types.Slice) (sliceEncoder, error) {
	elemEnc, err := varEncoder(apiName, t.Elem())
	if err != nil {
		return sliceEncoder{}, fmt.Errorf("unsupported slice underlying type %s: %w", t.Elem(), err)
	}

	return sliceEncoder{
		elemEnc:  elemEnc,
		lenEnc:   intEncoder,
		elemType: t.Elem(),
	}, nil
}

func (e sliceEncoder) encode(varId string) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "{ // %s\n", varId)
	fmt.Fprintf(sb, "var l int = len(%s)\n", varId)
	sb.WriteString(e.lenEnc.encode("l"))
	fmt.Fprintf(sb, `
	for i := 0; i < l; i++ {
		%s
	}
	`, e.elemEnc.encode(varId+"[i]"))
	sb.WriteString("}\n")

	return sb.String()
}

func (e sliceEncoder) decode(varId string) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "{ // %s\n", varId)
	sb.WriteString("var l int\n")
	sb.WriteString(e.lenEnc.decode("l"))
	fmt.Fprintf(sb, `
	%s = make([]%s, l)
	for i := 0; i < l; i++ {
		%s
	}
	`, varId, e.elemType.String(), e.elemEnc.decode(varId+"[i]"))
	sb.WriteString("}\n")

	return sb.String()
}

func (e sliceEncoder) imports() []string {
	return append(e.elemEnc.imports(), e.lenEnc.imports()...)
}

func (e sliceEncoder) codeblock() string {
	return ""
}

type structField struct {
	name string
	enc  encoder
}

type namedStructEncoder struct {
	fields []structField
}

func newNamedStructEncoder(apiName string, s *types.Struct) (namedStructEncoder, error) {
	structFields := []structField{}
	for i := 0; i < s.NumFields(); i++ {
		f := s.Field(i)
		enc, err := varEncoder(apiName, f.Type())
		if err != nil {
			return namedStructEncoder{}, fmt.Errorf("cannot encode structs field '%s' of type '%s': %w", f.Name(), f.Type(), err)
		}
		structFields = append(structFields, structField{
			name: f.Name(),
			enc:  enc,
		})
	}
	return namedStructEncoder{
		fields: structFields,
	}, nil
}

func (e namedStructEncoder) encode(varId string) string {
	sb := strings.Builder{}
	for _, f := range e.fields {
		sb.WriteString(f.enc.encode(varId + "." + f.name))
	}
	return sb.String()
}

func (e namedStructEncoder) decode(varId string) string {
	sb := &strings.Builder{}
	for _, f := range e.fields {
		sb.WriteString(f.enc.decode(varId + "." + f.name))
	}
	return sb.String()
}

func (e namedStructEncoder) imports() []string {
	imps := newOrderedSet[string]()
	for _, f := range e.fields {
		imps.add(f.enc.imports()...)
	}
	return imps.ordered
}

func (e namedStructEncoder) codeblock() string {
	return ""
}

type ifaceFunc struct {
	funcName string
	results  []ifaceRtnVar
}

// comma separated list of variable names and types. ex: "a int, b float64"
func (ifnc ifaceFunc) rtnParams() string {
	// todo: somehow share with paramStructGenerator's funcCallParams ?
	b := &strings.Builder{}
	for i, v := range ifnc.results {
		fmt.Fprintf(b, "%s %s", v.varName, v.rtnTypeName)
		if i != len(ifnc.results)-1 {
			b.WriteString(",")
		}
	}
	return b.String()
}

func (ifnc ifaceFunc) retParamsPrefixed(prefix string) string {
	// todo: share with paramstructgenerator
	buf := bytes.NewBuffer(nil)
	for i, p := range ifnc.results {
		fmt.Fprintf(buf, "%s%s", prefix, p.implParamName)
		if i != len(ifnc.results)-1 {
			fmt.Fprintf(buf, ",")
		}
	}
	return buf.String()
}

type ifaceRtnVar struct {
	rtnTypeName   string
	rtnType       types.Type
	enc           encoder
	varName       string
	implParamName string // name as used within interface's implementation struct
}

type interfaceEncoder struct {
	name         string
	fncs         []ifaceFunc
	implTypeName string
}

func newInterfaceEncoder(name string, apiName string, it *types.Interface) (interfaceEncoder, error) {
	fncs := []ifaceFunc{}
	for i := 0; i < it.NumMethods(); i++ {
		m := it.Method(i)
		sig := m.Type().(*types.Signature)

		if sig.Params().Len() != 0 {
			return interfaceEncoder{}, fmt.Errorf("unexpectedly params are not nil. they are: %+v", sig.Params())
		}

		results := []ifaceRtnVar{}
		for i := 0; i < sig.Results().Len(); i++ {
			r := sig.Results().At(i)
			enc, err := varEncoder(apiName, r.Type())
			if err != nil {
				return interfaceEncoder{}, fmt.Errorf("newInterfaceEncoder: no encoder for %s of type %s: %w", r.Name(), r.Type(), err)
			}
			results = append(results, ifaceRtnVar{
				varName:       r.Name(),
				enc:           enc,
				rtnType:       r.Type(),
				rtnTypeName:   r.Type().String(),
				implParamName: fmt.Sprintf("_%s_%d_%s", m.Name(), i, r.Name()), //"_" + m.Name() + r.Name(),
			})
		}

		fncs = append(fncs, ifaceFunc{
			funcName: m.Name(),
			results:  results,
		})
	}

	return interfaceEncoder{
		name: name,
		// we need to make the type unique within package, because we want each file to be self contained
		// it would be possible to use filename instead of apiName, but that would confuse file renaming. this will do for now
		implTypeName: "_" + name + "_" + apiName + "_irpcInterfaceImpl",
		fncs:         fncs,
	}, nil
}

func (e interfaceEncoder) encode(varId string) string {
	sb := &strings.Builder{}
	sb.WriteString("{\n") // separate block
	fmt.Fprintf(sb, `var isNil bool
	if %s == nil {
		isNil = true
	}
	%s
	`, varId, boolEncoder.encode("isNil"))
	sb.WriteString("if !isNil{\n")
	for _, f := range e.fncs {
		fmt.Fprintf(sb, "{ // %s()\n", f.funcName)
		for i, v := range f.results {
			sb.WriteString(v.implParamName)
			if i != len(f.results)-1 {
				sb.WriteString(",")
			}
		}
		fmt.Fprintf(sb, ":= %s.%s()\n", varId, f.funcName)
		for _, v := range f.results {
			sb.WriteString(v.enc.encode(v.implParamName))
		}
		sb.WriteString("}\n")
	}
	sb.WriteString("}\n") // if !isNil
	sb.WriteString("}\n") // separate block

	return sb.String()
}

func (e interfaceEncoder) decode(varId string) string {
	sb := &strings.Builder{}
	sb.WriteString("{\n") // separate block
	fmt.Fprintf(sb, `var isNil bool
	%s
	if isNil {
		%s = nil
	} else {
	`, boolEncoder.decode("isNil"), varId)

	fmt.Fprintf(sb, "var impl %s\n", e.implTypeName)
	for _, f := range e.fncs {
		fmt.Fprintf(sb, "{ // %s()\n", f.funcName)
		for _, v := range f.results {
			sb.WriteString(v.enc.decode("impl." + v.implParamName))
		}
		sb.WriteString("}\n")
	}
	fmt.Fprintf(sb, "%s = impl\n", varId)
	sb.WriteString("}\n") // else {
	sb.WriteString("}\n") // separate block

	return sb.String()
}

func (e interfaceEncoder) imports() []string {
	imps := newOrderedSet[string]()
	for _, f := range e.fncs {
		for _, v := range f.results {
			imps.add(v.enc.imports()...)
		}
	}
	return imps.ordered
}

func (e interfaceEncoder) codeblock() string {
	sb := &strings.Builder{}

	// type declaration
	fmt.Fprintf(sb, "type %s struct {\n", e.implTypeName)
	for _, f := range e.fncs {
		for _, v := range f.results {
			fmt.Fprintf(sb, "%s %s\n", v.implParamName, v.rtnTypeName)
		}
	}
	sb.WriteString("}\n")

	// fncs
	for _, f := range e.fncs {
		fmt.Fprintf(sb, "func (i %s)%s()(%s){\n", e.implTypeName, f.funcName, f.rtnParams())
		fmt.Fprintf(sb, "return %s\n", f.retParamsPrefixed("i."))
		sb.WriteString("}\n")
	}

	return sb.String()
}
