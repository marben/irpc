package main

import (
	"bytes"
	"fmt"
	"go/types"
	"strings"
)

type encoder interface {
	encode(varId string, existingVars []string) string // inline variable encode
	decode(varId string, existingVars []string) string // inline variable decode
	imports() []string                                 // necessary imports
	codeblock() string                                 // requested encoder's code block at top level
}

func varEncoder(apiName string, t types.Type, q types.Qualifier) (encoder, error) {
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
			return stringEncoder, nil
		default:
			return nil, fmt.Errorf("unsupported basic type '%s'", t.Name())
		}
	case *types.Slice:
		// []byte
		elemEnc, err := varEncoder(apiName, t.Elem(), q)
		if err != nil {
			return nil, fmt.Errorf("couldn't find encoder for type %v", t.Elem())
		}
		if elemEnc == uint8Encoder {
			return byteSliceEncoder, nil
		}

		// anything else
		return newSliceEncoder(apiName, t, q)
	case *types.Named:
		switch t.String() {
		case "context.Context": // we treat context special
			return contextEncoder{}, nil
		default:
			name := t.Obj().Name()
			switch ut := t.Underlying().(type) {
			case *types.Struct:
				return newNamedStructEncoder(apiName, ut, q)
			case *types.Interface:
				return newInterfaceEncoder(name, apiName, ut, q)
			default:
				return nil, fmt.Errorf("unsupported named type: %s", ut)
			}
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
		decFuncName: "VarInt",
		typeName:    "int",
	}
	uintEncoder = primitiveTypeEncoder{
		decFuncName: "UvarInt",
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
		decFuncName: "VarInt16",
		typeName:    "int16",
	}
	uint16Encoder = primitiveTypeEncoder{
		decFuncName: "UvarInt16",
		typeName:    "uint16",
	}
	int32Encoder = primitiveTypeEncoder{
		decFuncName: "VarInt32",
		typeName:    "int32",
	}
	uint32Encoder = primitiveTypeEncoder{
		decFuncName: "UvarInt32",
		typeName:    "uint32",
	}
	int64Encoder = primitiveTypeEncoder{
		decFuncName: "VarInt64",
		typeName:    "int64",
	}
	uint64Encoder = primitiveTypeEncoder{
		decFuncName: "UvarInt64",
		typeName:    "uint64",
	}
	float32Encoder = primitiveTypeEncoder{
		decFuncName: "Float32le",
		typeName:    "float32",
	}
	float64Encoder = primitiveTypeEncoder{
		decFuncName: "Float64le",
		typeName:    "float64",
	}
	stringEncoder = primitiveTypeEncoder{
		decFuncName: "String",
		typeName:    "string",
	}
	byteSliceEncoder = primitiveTypeEncoder{
		decFuncName: "ByteSlice",
		typeName:    "[]byte",
	}
)

type primitiveTypeEncoder struct {
	decFuncName string
	typeName    string
}

func (e primitiveTypeEncoder) encode(varId string, existingVars []string) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, `if err := e.%s(%s); err != nil {
		return fmt.Errorf("serialize %s of type '%s': %%w", err)
	}
	`, e.decFuncName, varId, varId, e.typeName)

	return sb.String()
}

func (e primitiveTypeEncoder) decode(varId string, existingVars []string) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, `if err := d.%s(&%s); err != nil{
		return fmt.Errorf("deserialize %s of type '%s': %%w",err)
	}
	`, e.decFuncName, varId, varId, e.typeName)
	return sb.String()
}

func (e primitiveTypeEncoder) imports() []string {
	return []string{irpcGenImport, fmtImport}
}

func (e primitiveTypeEncoder) codeblock() string {
	return ""
}

// slice encoder doesn't use generic slice function for performance reasons
// unrolling the code in place reduces time and allocs
type sliceEncoder struct {
	elemEnc  encoder
	lenEnc   encoder
	elemType types.Type
	q        types.Qualifier
}

func newSliceEncoder(apiName string, t *types.Slice, q types.Qualifier) (sliceEncoder, error) {
	elemEnc, err := varEncoder(apiName, t.Elem(), q)
	if err != nil {
		return sliceEncoder{}, fmt.Errorf("unsupported slice underlying type %s: %w", t.Elem(), err)
	}

	return sliceEncoder{
		elemEnc:  elemEnc,
		lenEnc:   intEncoder,
		elemType: t.Elem(),
		q:        q,
	}, nil
}

func (e sliceEncoder) encode(varId string, existingVars []string) string {
	sb := &strings.Builder{}

	// length
	fmt.Fprintf(sb, "{ // %s\n", varId)
	fmt.Fprintf(sb, "var l int = len(%s)\n", varId)
	sb.WriteString(e.lenEnc.encode("l", existingVars))
	existingVars = append(existingVars, "l")

	// for loop
	itName := generateSliceIteratorName(existingVars)
	fmt.Fprintf(sb, `
	for %[1]s := 0; %[1]s < l; %[1]s++ {
		%s
	}
	`, itName, e.elemEnc.encode(varId+"["+itName+"]", append(existingVars, itName)))
	sb.WriteString("}\n")

	return sb.String()
}

func (e sliceEncoder) decode(varId string, existingVars []string) string {
	sb := &strings.Builder{}

	// length
	fmt.Fprintf(sb, "{ // %s\n", varId)
	sb.WriteString("var l int\n")
	sb.WriteString(e.lenEnc.decode("l", existingVars))
	existingVars = append(existingVars, "l")

	// for loop
	itName := generateSliceIteratorName(existingVars)
	fmt.Fprintf(sb, `
	%s = make([]%s, l)
	for %[3]s := 0; %[3]s < l; %[3]s++ {
		%s
	}
	`, varId, types.TypeString(e.elemType, e.q), itName, e.elemEnc.decode(varId+"["+itName+"]", append(existingVars, itName)))
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

func newNamedStructEncoder(apiName string, s *types.Struct, q types.Qualifier) (namedStructEncoder, error) {
	structFields := []structField{}
	for i := 0; i < s.NumFields(); i++ {
		f := s.Field(i)
		enc, err := varEncoder(apiName, f.Type(), q)
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

func (e namedStructEncoder) encode(varId string, existingVars []string) string {
	sb := strings.Builder{}
	for _, f := range e.fields {
		sb.WriteString(f.enc.encode(varId+"."+f.name, existingVars))
	}
	return sb.String()
}

func (e namedStructEncoder) decode(varId string, existingVars []string) string {
	sb := &strings.Builder{}
	for _, f := range e.fields {
		sb.WriteString(f.enc.decode(varId+"."+f.name, existingVars))
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

func newInterfaceEncoder(name string, apiName string, it *types.Interface, q types.Qualifier) (interfaceEncoder, error) {
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
			enc, err := varEncoder(apiName, r.Type(), q)
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

func (e interfaceEncoder) encode(varId string, existingVars []string) string {
	sb := &strings.Builder{}
	sb.WriteString("{\n") // separate block
	fmt.Fprintf(sb, `var isNil bool
	if %s == nil {
		isNil = true
	}
	%s
	`, varId, boolEncoder.encode("isNil", existingVars))
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
			sb.WriteString(v.enc.encode(v.implParamName, existingVars))
		}
		sb.WriteString("}\n")
	}
	sb.WriteString("}\n") // if !isNil
	sb.WriteString("}\n") // separate block

	return sb.String()
}

func (e interfaceEncoder) decode(varId string, existingVars []string) string {
	sb := &strings.Builder{}
	sb.WriteString("{\n") // separate block
	fmt.Fprintf(sb, `var isNil bool
	%s
	if isNil {
		%s = nil
	} else {
	`, boolEncoder.decode("isNil", existingVars), varId)

	fmt.Fprintf(sb, "var impl %s\n", e.implTypeName)
	for _, f := range e.fncs {
		fmt.Fprintf(sb, "{ // %s()\n", f.funcName)
		for _, v := range f.results {
			sb.WriteString(v.enc.decode("impl."+v.implParamName, existingVars))
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

var _ encoder = contextEncoder{}

// contextEncoder is a placeholder for context params
// we don't actually pass any data. instead implement our own cancelling function
type contextEncoder struct {
}

// codeblock implements encoder.
func (c contextEncoder) codeblock() string {
	return ""
}

// decode implements encoder.
func (c contextEncoder) decode(varId string, existingVars []string) string {
	return "// no code for context decoding\n"
}

// encode implements encoder.
func (c contextEncoder) encode(varId string, existingVars []string) string {
	return "// no code for context encoding\n"
}

// imports implements encoder.
func (c contextEncoder) imports() []string {
	return []string{contextImport}
}
