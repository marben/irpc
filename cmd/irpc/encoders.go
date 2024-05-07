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

var (
	boolEncoder = primitiveTypeEncoder{ // todo: could use bitpacking instead of one bool per byte
		bufSize: 1,
		encStr: `if %s {
			b[0] = 1
		} else {
			b[0] = 0
		}`,
		decStr: `if b[0] == 0 {
			%[1]s = false
		} else {
			%[1]s = true
		}`,
		typeName: "bool",
		imps:     nil,
	}
	intEncoder = primitiveTypeEncoder{
		bufSize:  8,
		encStr:   "binary.LittleEndian.PutUint64(b, uint64(%s))",
		decStr:   "%s = int(binary.LittleEndian.Uint64(b))",
		typeName: "int",
		imps:     []string{binaryImport},
	}
	uintEncoder = primitiveTypeEncoder{
		bufSize:  8,
		encStr:   "binary.LittleEndian.PutUint64(b, uint64(%s))",
		decStr:   "%s = uint(binary.LittleEndian.Uint64(b))",
		typeName: "uint",
		imps:     []string{binaryImport},
	}
	int8Encoder = primitiveTypeEncoder{
		bufSize:  1,
		encStr:   "b[0] = byte(%s)",
		decStr:   "%s = int8(b[0])",
		typeName: "int8",
		imps:     nil,
	}
	uint8Encoder = primitiveTypeEncoder{
		bufSize:  1,
		encStr:   "b[0] = byte(%s)",
		decStr:   "%s = uint8(b[0])",
		typeName: "uint8",
		imps:     nil,
	}
	int16Encoder = primitiveTypeEncoder{
		bufSize:  2,
		encStr:   "binary.LittleEndian.PutUint16(b, uint16(%s))",
		decStr:   "%s = int16(binary.LittleEndian.Uint16(b))",
		typeName: "int16",
		imps:     []string{binaryImport},
	}
	uint16Encoder = primitiveTypeEncoder{
		bufSize:  2,
		encStr:   "binary.LittleEndian.PutUint16(b, %s)",
		decStr:   "%s = binary.LittleEndian.Uint16(b)",
		typeName: "uint16",
		imps:     []string{binaryImport},
	}
	int32Encoder = primitiveTypeEncoder{
		bufSize:  4,
		encStr:   "binary.LittleEndian.PutUint32(b, uint32(%s))",
		decStr:   "%s = int32(binary.LittleEndian.Uint32(b))",
		typeName: "int32",
		imps:     []string{binaryImport},
	}
	uint32Encoder = primitiveTypeEncoder{
		bufSize:  4,
		encStr:   "binary.LittleEndian.PutUint32(b, %s)",
		decStr:   "%s = binary.LittleEndian.Uint32(b)",
		typeName: "uint32",
		imps:     []string{binaryImport},
	}
	int64Encoder = primitiveTypeEncoder{
		bufSize:  8,
		encStr:   "binary.LittleEndian.PutUint64(b, uint64(%s))",
		decStr:   "%s = int64(binary.LittleEndian.Uint64(b))",
		typeName: "int64",
		imps:     []string{binaryImport},
	}
	uint64Encoder = primitiveTypeEncoder{
		bufSize:  8,
		encStr:   "binary.LittleEndian.PutUint64(b, %s)",
		decStr:   "%s = binary.LittleEndian.Uint64(b)",
		typeName: "uint64",
		imps:     []string{binaryImport},
	}
	float32Encoder = primitiveTypeEncoder{
		bufSize:  4,
		encStr:   "binary.LittleEndian.PutUint32(b, math.Float32bits(%s))",
		decStr:   "%s = math.Float32frombits(binary.LittleEndian.Uint32(b))",
		typeName: "float32",
		imps:     []string{binaryImport, mathImport},
	}
	float64Encoder = primitiveTypeEncoder{
		bufSize:  8,
		encStr:   "binary.LittleEndian.PutUint64(b, math.Float64bits(%s))",
		decStr:   "%s = math.Float64frombits(binary.LittleEndian.Uint64(b))",
		typeName: "float64",
		imps:     []string{binaryImport, mathImport},
	}
)

type primitiveTypeEncoder struct {
	bufSize        int
	encStr, decStr string
	typeName       string
	imps           []string
}

func (e primitiveTypeEncoder) encode(varId string) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "b := make([]byte, %d)\n", e.bufSize)
	fmt.Fprintf(sb, e.encStr, varId)
	fmt.Fprintf(sb, `
	if _, err := w.Write(b[:%d]); err != nil {
		return fmt.Errorf("%s %s write: %%w", err)
	}
	`, e.bufSize, varId, e.typeName)

	return sb.String()
}

func (e primitiveTypeEncoder) decode(varId string) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "b := make([]byte, %d)", e.bufSize)
	fmt.Fprintf(sb, `
	if _, err := io.ReadFull(r, b[:%d]); err != nil {
		return fmt.Errorf("%s %s decode: %%w", err)
	}
	`, e.bufSize, varId, e.typeName)
	fmt.Fprintf(sb, e.decStr, varId)
	return sb.String()
}

func (e primitiveTypeEncoder) imports() []string {
	return append(e.imps, fmtImport)
}

func (e primitiveTypeEncoder) codeblock() string {
	return ""
}

// todo: perhaps rewrite using slice encoder?
type stringEncoder struct{}

func (e stringEncoder) encode(varId string) string {
	s := fmt.Sprintf("var l int = len(%s)\n", varId)
	s += intEncoder.encode("l")
	s += fmt.Sprintf(`
	_, err := w.Write([]byte(%s))
	if err != nil {
		return fmt.Errorf("failed to write string to writer: %%w", err)
	}
	`, varId)
	return s
}

func (e stringEncoder) decode(varId string) string {
	s := "var l int\n"
	s += intEncoder.decode("l")
	s += fmt.Sprintf(`
	sbuf := make([]byte, l)
	_, err := io.ReadFull(r, sbuf)
	if err != nil {
		return fmt.Errorf("failed to read string data from reader: %%w", err)
	}
	%s = string(sbuf)
	`, varId)
	return "{\n" + s + "}\n"
}

func (e stringEncoder) imports() []string {
	return append(intEncoder.imports(), fmtImport)
}

func (e stringEncoder) codeblock() string {
	return ""
}

type sliceEncoder struct {
	elemEnc  encoder
	lenEnc   encoder
	elemType types.Type
}

func newSliceEncoder(t *types.Slice) (sliceEncoder, error) {
	elemEnc, err := varEncoder(t.Elem())
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
	fmt.Fprintf(sb, "var l int = len(%s)\n", varId)
	sb.WriteString(e.lenEnc.encode("l"))
	fmt.Fprintf(sb, `
	for i := 0; i < l; i++ {
		%s
	}
	`, e.elemEnc.encode(varId+"[i]"))

	return sb.String()
}

func (e sliceEncoder) decode(varId string) string {
	sb := &strings.Builder{}
	sb.WriteString("var l int\n")
	sb.WriteString(e.lenEnc.decode("l"))
	fmt.Fprintf(sb, `
	%s = make([]%s, l)
	for i := 0; i < l; i++ {
		%s
	}`, varId, e.elemType.String(), e.elemEnc.decode(varId+"[i]"))

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

func newNamedStructEncoder(s *types.Struct) (namedStructEncoder, error) {
	structFields := []structField{}
	for i := 0; i < s.NumFields(); i++ {
		f := s.Field(i)
		enc, err := varEncoder(f.Type())
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
		sb.WriteString("{\n") // todo: add type to comments, like in paramstruct generator
		sb.WriteString(f.enc.encode(varId + "." + f.name))
		sb.WriteString("}\n")
	}
	return sb.String()
}
func (e namedStructEncoder) decode(varId string) string {
	sb := &strings.Builder{}
	for _, f := range e.fields {
		sb.WriteString("{\n") // todo: add type to comments, like in paramstruct generator
		sb.WriteString(f.enc.decode(varId + "." + f.name))
		sb.WriteString("}\n")
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

func newInterfaceEncoder(name string, it *types.Interface) (interfaceEncoder, error) {
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
			enc, err := varEncoder(r.Type())
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
		name:         name,
		implTypeName: "_" + name + "_irpcInterfaceImpl",
		fncs:         fncs,
	}, nil
}

func (e interfaceEncoder) encode(varId string) string {
	sb := &strings.Builder{}
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

	return sb.String()
}
func (e interfaceEncoder) decode(varId string) string {
	sb := &strings.Builder{}
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
