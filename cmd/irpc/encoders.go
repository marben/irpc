package main

import (
	"bytes"
	"fmt"
	"go/importer"
	"go/types"
	"strings"
)

type encoder interface {
	encode(varId string, existingVars []string) string // inline variable encode
	decode(varId string, existingVars []string) string // inline variable decode
	codeblock() string                                 // requested encoder's code block at top level
}

type encoderResolver struct {
	binMarshaler, binUnmarshaler *types.Interface
	apiName                      string
	qualifier                    types.Qualifier
	imports                      orderedSet[importSpec]
	typesInfo                    *types.Info
}

func newEncoderResolver(apiName string, qualifier types.Qualifier, typesInfo *types.Info) (*encoderResolver, error) {
	imp := importer.Default()
	encodingPkg, err := imp.Import("encoding")
	if err != nil {
		return nil, fmt.Errorf("importer.Import(\"encoding\"): %w", err)
	}
	binMarshaler, ok := encodingPkg.Scope().Lookup("BinaryMarshaler").Type().Underlying().(*types.Interface)
	if !ok {
		return nil, fmt.Errorf("failed to find encoding.BinaryMarshaller type")
	}
	binUnmarshaler, ok := encodingPkg.Scope().Lookup("BinaryUnmarshaler").Type().Underlying().(*types.Interface)
	if !ok {
		return nil, fmt.Errorf("failed to find encoding.BinaryUnmarshaller type")
	}

	return &encoderResolver{
			binMarshaler:   binMarshaler,
			binUnmarshaler: binUnmarshaler,
			apiName:        apiName,
			qualifier:      qualifier,
			imports:        newOrderedSet[importSpec](),
			typesInfo:      typesInfo,
		},
		nil
}

func (er *encoderResolver) varEncoder(t types.Type) (encoder, error) {
	if types.Implements(t, er.binMarshaler) {
		if !types.Implements(types.NewPointer(t), er.binUnmarshaler) {
			return nil, fmt.Errorf("%T implements BinaryMarshaler, but %T doesn't implement BinaryUnmarshaler", t, types.NewPointer(t))
		}
		return er.newBinaryMarshalerEncoder(t)
	}

	switch t := t.(type) {
	case *types.Basic:
		return er.newBasicTypeEncoder(t)
	case *types.Slice:
		return er.newSliceEncoder(t, "")
	case *types.Map:
		return er.newMapEncoder(t.Key(), t.Elem())
	case *types.Struct:
		return er.newStructEncoder(t)
	case *types.Named:
		name := types.TypeString(t, er.qualifier)
		if name == "context.Context" {
			return contextEncoder{}, nil
		}
		// log.Printf("named type %q with pkg name %q nad path %q", t.Obj().Name(), t.Obj().Pkg().Name(), t.Obj().Pkg().Path())
		switch ut := t.Underlying().(type) {
		case *types.Basic:
			return er.newNamedBasicTypeEncoder(t, ut, name)
		case *types.Slice:
			return er.newSliceEncoder(ut, name)
		case *types.Map:
			return er.newMapEncoder(ut.Key(), ut.Elem())
		case *types.Struct:
			return er.newStructEncoder(ut)
		case *types.Interface:
			return er.newInterfaceEncoder(name, ut)

		default:
			return nil, fmt.Errorf("unsupported named type: '%s'", name)
		}

	default:
		return nil, fmt.Errorf("unsupported type '%v'", t)
	}
}

func (er *encoderResolver) newBasicTypeEncoder(t *types.Basic) (directCallEncoder, error) {
	er.imports.add(irpcGenImport, fmtImport)
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
		return directCallEncoder{}, fmt.Errorf("unsupported basic type '%s'", t.Name())
	}
}

func (er *encoderResolver) newBinaryMarshalerEncoder(t types.Type) (encoder, error) {
	named, isNamed := t.(*types.Named)
	if !isNamed {
		return nil, fmt.Errorf("binMarshaller %T is not named", t)
	}

	er.imports.add(newImportSpec("", named.Obj().Pkg()))

	return directCallEncoder{
		encFuncName:        "BinaryMarshaler",
		decFuncName:        "BinaryUnmarshaler",
		underlyingTypeName: "encoding.BinaryUnmarshaler", // todo: make encoding type/decoding type
	}, nil
}

func (er *encoderResolver) newNamedBasicTypeEncoder(n *types.Named, t *types.Basic, name string) (directCallEncoder, error) {
	basicEnc, err := er.newBasicTypeEncoder(t)
	if err != nil {
		return directCallEncoder{}, fmt.Errorf("get basic type encoder for named type %s: %w", name, err)
	}
	basicEnc.typeName = name
	return basicEnc, nil
}

func newSymmetricDirectCallEncoder(encDecFunc string, underlyingTypeName string) directCallEncoder {
	return directCallEncoder{
		encFuncName:        encDecFunc,
		decFuncName:        encDecFunc,
		underlyingTypeName: underlyingTypeName,
	}
}

var (
	boolEncoder      = newSymmetricDirectCallEncoder("Bool", "bool")
	intEncoder       = newSymmetricDirectCallEncoder("VarInt", "int")
	uintEncoder      = newSymmetricDirectCallEncoder("UvarInt", "uint")
	int8Encoder      = newSymmetricDirectCallEncoder("Int8", "int8")
	uint8Encoder     = newSymmetricDirectCallEncoder("Uint8", "uint8")
	int16Encoder     = newSymmetricDirectCallEncoder("VarInt16", "int16")
	uint16Encoder    = newSymmetricDirectCallEncoder("UvarInt16", "uint16")
	int32Encoder     = newSymmetricDirectCallEncoder("VarInt32", "int32")
	uint32Encoder    = newSymmetricDirectCallEncoder("UvarInt32", "uint32")
	int64Encoder     = newSymmetricDirectCallEncoder("VarInt64", "int64")
	uint64Encoder    = newSymmetricDirectCallEncoder("UvarInt64", "uint64")
	float32Encoder   = newSymmetricDirectCallEncoder("Float32le", "float32")
	float64Encoder   = newSymmetricDirectCallEncoder("Float64le", "float64")
	stringEncoder    = newSymmetricDirectCallEncoder("String", "string")
	byteSliceEncoder = newSymmetricDirectCallEncoder("ByteSlice", "[]byte")
)

type importSpec struct {
	alias string
	path  string
}

func newImportSpec(alias string, pkg *types.Package) importSpec {
	return importSpec{alias: alias, path: pkg.Path()}
}

type directCallEncoder struct {
	encFuncName        string
	decFuncName        string
	underlyingTypeName string
	typeName           string // if named, otherwise ""
}

func (e directCallEncoder) encode(varId string, existingVars []string) string {
	var varParam string
	if e.typeName == "" {
		varParam = varId
	} else {
		varParam = fmt.Sprintf("%s(%s)", e.underlyingTypeName, varId)
	}

	return fmt.Sprintf(`if err := e.%s(%s); err != nil {
		return fmt.Errorf("serialize %s of type '%s': %%w", err)
	}
	`, e.encFuncName, varParam, varId, e.underlyingTypeName)
}

func (e directCallEncoder) decode(varId string, existingVars []string) string {
	var varParam string
	if e.typeName == "" {
		varParam = "&" + varId
	} else {
		varParam = fmt.Sprintf("(*%s)(&%s)", e.underlyingTypeName, varId)
	}
	sb := &strings.Builder{}
	fmt.Fprintf(sb, `if err := d.%s(%s); err != nil{
		return fmt.Errorf("deserialize %s of type '%s': %%w",err)
	}
	`, e.decFuncName, varParam, varId, e.underlyingTypeName)
	return sb.String()
}

func (e directCallEncoder) codeblock() string {
	return ""
}

// slice encoder doesn't use generic slice function for performance reasons
// unrolling the code in place reduces time and allocs
type sliceEncoder struct {
	elemEnc     encoder
	elemTypeStr string
	lenEnc      encoder
}

// name can be "", in which case no type conversion will happen
func (er *encoderResolver) newSliceEncoder(t *types.Slice, name string) (encoder, error) {
	if t.Elem().String() == "byte" {
		bs := byteSliceEncoder
		bs.typeName = name
		return bs, nil
	}

	elemEnc, err := er.varEncoder(t.Elem())
	if err != nil {
		return sliceEncoder{}, fmt.Errorf("unsupported slice underlying type %s: %w", t.Elem(), err)
	}

	return sliceEncoder{
		elemEnc:     elemEnc,
		elemTypeStr: types.TypeString(t.Elem(), er.qualifier),
		lenEnc:      uint64Encoder,
	}, nil
}

func (e sliceEncoder) encode(varId string, existingVars []string) string {
	sb := &strings.Builder{}

	// length
	fmt.Fprintf(sb, "{ // %s []%s\n", varId, e.elemTypeStr)
	fmt.Fprintf(sb, "var l int = len(%s)\n", varId)
	sb.WriteString(e.lenEnc.encode("uint64(l)", existingVars))
	existingVars = append(existingVars, "l")

	// for loop
	existingVars = append(existingVars, "v")
	fmt.Fprintf(sb, "for _, v := range %s {\n", varId)
	sb.WriteString(e.elemEnc.encode("v", existingVars))
	sb.WriteString("}")
	sb.WriteString("}\n")

	return sb.String()
}

func (e sliceEncoder) decode(varId string, existingVars []string) string {
	sb := &strings.Builder{}

	// length
	fmt.Fprintf(sb, "{ // %s []%s\n", varId, e.elemTypeStr)
	sb.WriteString("var ul uint64\n")
	sb.WriteString(e.lenEnc.decode("ul", existingVars))
	sb.WriteString("var l int = int(ul)\n")
	existingVars = append(existingVars, "l", "ul")

	// for loop
	itName := generateIteratorName(existingVars)
	existingVars = append(existingVars, itName)
	fmt.Fprintf(sb, "%s = make([]%s, l)\n", varId, e.elemTypeStr)
	fmt.Fprintf(sb, "for %s := range l {", itName)
	sb.WriteString(e.elemEnc.decode(varId+"["+itName+"]", existingVars))
	sb.WriteString("}\n")
	sb.WriteString("}\n")

	return sb.String()
}

func (e sliceEncoder) codeblock() string {
	return ""
}

type mapEncoder struct {
	keyEnc     encoder
	keyTypeStr string

	valEnc     encoder
	valTypeStr string

	lenEnc encoder
}

func (er *encoderResolver) newMapEncoder(keyType, valType types.Type) (mapEncoder, error) {
	keyEnc, err := er.varEncoder(keyType)
	if err != nil {
		return mapEncoder{}, fmt.Errorf("unsupported map key type %s: %w", keyType, err)
	}

	valEnc, err := er.varEncoder(valType)
	if err != nil {
		return mapEncoder{}, fmt.Errorf("unsupported map value type %s: %w", valType, err)
	}
	return mapEncoder{
		keyEnc:     keyEnc,
		keyTypeStr: types.TypeString(keyType, er.qualifier),
		valEnc:     valEnc,
		valTypeStr: types.TypeString(valType, er.qualifier),
		lenEnc:     uint64Encoder,
	}, nil
}

func (e mapEncoder) encode(varId string, existingVars []string) string {
	sb := &strings.Builder{}

	// length
	fmt.Fprintf(sb, "{ // %s map[%s]%s\n", varId, e.keyTypeStr, e.valTypeStr)
	fmt.Fprintf(sb, "var l int = len(%s)\n", varId)
	sb.WriteString(e.lenEnc.encode("uint64(l)", existingVars))
	existingVars = append(existingVars, "l")

	keyIt, valIt := generateKeyValueIteratorNames(existingVars)
	existingVars = append(existingVars, keyIt, valIt)

	// for loop
	fmt.Fprintf(sb, "for %s, %s := range %s {", keyIt, valIt, varId)
	sb.WriteString(e.keyEnc.encode(keyIt, existingVars))
	sb.WriteString(e.valEnc.encode(valIt, existingVars))
	sb.WriteString("}\n") // end of for loop

	sb.WriteString("}\n") // end of block

	return sb.String()
}

func (e mapEncoder) decode(varId string, existingVars []string) string {
	sb := &strings.Builder{}

	// length
	fmt.Fprintf(sb, "{ // %s\n", varId)
	sb.WriteString("var ul uint64\n")
	sb.WriteString(e.lenEnc.decode("ul", existingVars))
	sb.WriteString("var l int = int(ul)\n")
	existingVars = append(existingVars, "ul", "l")

	fmt.Fprintf(sb, "%s = make(map[%s]%s, l)\n", varId, e.keyTypeStr, e.valTypeStr)
	sb.WriteString("for range l {\n")

	fmt.Fprintf(sb, "var k %s\n", e.keyTypeStr)
	existingVars = append(existingVars, "k")
	fmt.Fprintf(sb, "%s\n", e.keyEnc.decode("k", existingVars))

	fmt.Fprintf(sb, "var v %s\n", e.valTypeStr)
	existingVars = append(existingVars, "v")
	fmt.Fprintf(sb, "%s\n", e.valEnc.decode("v", existingVars))

	fmt.Fprintf(sb, "%s[k] = v", varId)
	sb.WriteString("}\n")
	sb.WriteString("}\n") // end of block
	return sb.String()
}

func (e mapEncoder) codeblock() string {
	return ""
}

type structField struct {
	name string
	enc  encoder
}

type structEncoder struct {
	fields []structField
}

func (er *encoderResolver) newStructEncoder(s *types.Struct) (structEncoder, error) {
	structFields := []structField{}
	for i := 0; i < s.NumFields(); i++ {
		f := s.Field(i)
		enc, err := er.varEncoder(f.Type())
		if err != nil {
			return structEncoder{}, fmt.Errorf("cannot encode structs field '%s' of type '%s': %w", f.Name(), f.Type(), err)
		}
		structFields = append(structFields, structField{
			name: f.Name(),
			enc:  enc,
		})
	}
	return structEncoder{
		fields: structFields,
	}, nil
}

func (e structEncoder) encode(varId string, existingVars []string) string {
	sb := strings.Builder{}
	for _, f := range e.fields {
		sb.WriteString(f.enc.encode(varId+"."+f.name, existingVars))
	}
	return sb.String()
}

func (e structEncoder) decode(varId string, existingVars []string) string {
	sb := &strings.Builder{}
	for _, f := range e.fields {
		sb.WriteString(f.enc.decode(varId+"."+f.name, existingVars))
	}
	return sb.String()
}

func (e structEncoder) codeblock() string {
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

func (er *encoderResolver) newInterfaceEncoder(name string, it *types.Interface) (interfaceEncoder, error) {
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
			enc, err := er.varEncoder(r.Type())
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
		implTypeName: "_" + name + "_" + er.apiName + "_irpcInterfaceImpl",
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
