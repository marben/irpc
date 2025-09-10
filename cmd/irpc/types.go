package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"strings"
)

type Type interface {
	encoder
	Name() string
	ImportSpecs() importSpec // todo: unnecessary?
}

type basicType struct {
	enc         directCallEncoder
	name        string
	importSpecs importSpec
}

func (bt basicType) Name() string {
	return bt.name
}

var _ Type = basicType{}

func (bt basicType) ImportSpecs() importSpec {
	return bt.importSpecs
}

// ast can be nil if not available
func (tr *typeResolver) newBasicTypeT(t types.Type, bt *types.Basic, astExpr ast.Expr) (Type, error) {
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
		return basicType{}, fmt.Errorf("unsupported basic type %q", bt.Name())
	}

	name, importSpec := tr.typeNameAndImport(t, astExpr)

	var needsCasting bool
	if name != bt.Name() {
		needsCasting = true
	}

	enc := directCallEncoder{
		encFuncName:        irpcFuncName,
		decFuncName:        irpcFuncName,
		underlyingTypeName: bt.Name(),
		needsCasting:       needsCasting,
	}

	return basicType{
		enc:         enc,
		importSpecs: importSpec,
		name:        name,
	}, nil
}

func (bt basicType) codeblock() string {
	return bt.enc.codeblock()
}

func (bt basicType) decode(varId string, existingVars []string, q *qualifier) string {
	return bt.enc.decode(varId, existingVars, q)
}

func (bt basicType) encode(varId string, existingVars []string) string {
	return bt.enc.encode(varId, existingVars)
}

func (bt basicType) Encoder() encoder {
	return bt.enc
}

type sliceType struct {
	elem       Type
	lenEnc     encoder
	typeName   string
	importSpec importSpec
}

func (st sliceType) ImportSpecs() importSpec {
	return st.importSpec
}

var _ Type = sliceType{}

func (tr *typeResolver) newSliceTypeT(apiName string, t types.Type, st *types.Slice, astExpr ast.Expr) (sliceType, error) {
	name, importSpec := tr.typeNameAndImport(t, astExpr)

	var elemAst ast.Expr
	if astExpr := tr.unwrapTypeAst(t, astExpr); astExpr != nil {
		arrayAst, ok := astExpr.(*ast.ArrayType)
		if !ok {
			return sliceType{}, fmt.Errorf("slice's astExpression is not *ast.ArrayType, but %T of value %#[1]v", astExpr)
		}
		elemAst = arrayAst.Elt
	}

	elemT, err := tr.newType(apiName, st.Elem(), elemAst)
	if err != nil {
		return sliceType{}, fmt.Errorf("newType() for slices element %q: %w", st.Elem(), err)
	}

	return sliceType{
		elem:       elemT,
		lenEnc:     uint64Encoder,
		typeName:   name,
		importSpec: importSpec,
	}, nil
}

func (st sliceType) Name() string {
	return st.typeName
}

// encode implements encoder.
func (st sliceType) encode(varId string, existingVars []string) string {
	sb := &strings.Builder{}

	// length
	fmt.Fprintf(sb, "{ // %s %s\n", varId, st.Name())
	fmt.Fprintf(sb, "var l int = len(%s)\n", varId)
	sb.WriteString(st.lenEnc.encode("uint64(l)", existingVars))
	existingVars = append(existingVars, "l")

	// for loop
	existingVars = append(existingVars, "v")
	fmt.Fprintf(sb, "for _, v := range %s {\n", varId)
	sb.WriteString(st.elem.encode("v", existingVars))
	sb.WriteString("}")
	sb.WriteString("}\n")

	return sb.String()
}

// decode implements encoder.
func (st sliceType) decode(varId string, existingVars []string, q *qualifier) string {
	sb := &strings.Builder{}

	// length
	fmt.Fprintf(sb, "{ // %s %s\n", varId, q.qualifyType(st))
	sb.WriteString("var ul uint64\n")
	sb.WriteString(st.lenEnc.decode("ul", existingVars, q))
	sb.WriteString("var l int = int(ul)\n")
	existingVars = append(existingVars, "l", "ul")

	// for loop
	itName := generateIteratorName(existingVars)
	existingVars = append(existingVars, itName)
	fmt.Fprintf(sb, "%s = make(%s, l)\n", varId, q.qualifyType(st))
	fmt.Fprintf(sb, "for %s := range l {", itName)
	sb.WriteString(st.elem.decode(varId+"["+itName+"]", existingVars, q))
	sb.WriteString("}\n")
	sb.WriteString("}\n")

	return sb.String()
}

// codeblock implements encoder.
func (st sliceType) codeblock() string {
	return ""
}

var _ Type = mapType{}

type mapType struct {
	lenEnc            encoder
	key, val          Type
	qualifiedTypeName string
}

func (tr *typeResolver) newMapType(apiName string, astExpr ast.Expr, t types.Type) (mapType, error) {
	var mapAst *ast.MapType
	named, ok := t.(*types.Named)
	if ok {
		log.Printf("map type is named: %v, ast expr is %#v", named, astExpr)
		// named.Obj()
		typeSpec, err := tr.g.findAstTypeSpec(named)
		if err != nil {
			return mapType{}, fmt.Errorf("findTypeSpec(): %w", err)
		}
		log.Printf("typeSpec: %#v", typeSpec)
		mapAst, ok = typeSpec.Type.(*ast.MapType)
		if !ok {
			return mapType{}, fmt.Errorf("couldn't find *ast.MapType in named map")
		}
	} else {
		mapAst, ok = astExpr.(*ast.MapType)
		if !ok {
			return mapType{}, fmt.Errorf("map's astExpression is not *ast.MapType, but %T of value %#[1]v", astExpr)
		}
	}

	log.Printf("map's ast: %#v", mapAst)

	qualifiedTypeName, err := tr.g.qualifiedTypeName(astExpr)
	if err != nil {
		return mapType{}, fmt.Errorf("qualifiedTypeName(%v): %w", astExpr, err)
	}

	// log.Printf("qualified type name: %q", qualifiedTypeName)

	keyT, err := tr.newType(apiName, &types.Slice{}, mapAst.Key)
	if err != nil {
		return mapType{}, fmt.Errorf("newType() for map key %q: %w", mapAst.Key, err)
	}
	log.Printf("keyT: %q", keyT.Name())
	valT, err := tr.newType(apiName, &types.Slice{}, mapAst.Value)
	if err != nil {
		return mapType{}, fmt.Errorf("newType() for map value %q: %w", mapAst.Value, err)
	}
	log.Printf("valT: %q", valT.Name())

	return mapType{
		lenEnc:            uint64Encoder,
		key:               keyT,
		val:               valT,
		qualifiedTypeName: qualifiedTypeName,
	}, nil
}

// ImportSpecs implements Type.
func (m mapType) ImportSpecs() importSpec {
	return importSpec{}
}

// Name implements Type.
func (m mapType) Name() string {
	return m.qualifiedTypeName
}

// codeblock implements Type.
func (m mapType) codeblock() string {
	return ""
}

// decode implements Type.
func (m mapType) decode(varId string, existingVars []string, q *qualifier) string {
	sb := &strings.Builder{}

	// length
	fmt.Fprintf(sb, "{ // %s\n", varId)
	sb.WriteString("var ul uint64\n")
	sb.WriteString(m.lenEnc.decode("ul", existingVars, q))
	sb.WriteString("var l int = int(ul)\n")
	existingVars = append(existingVars, "ul", "l")

	// fmt.Fprintf(sb, "%s = make(map[%s]%s, l)\n", varId, m.key.Name(), m.val.Name())
	fmt.Fprintf(sb, "%s = make(%s, l)\n", varId, m.Name())
	sb.WriteString("for range l {\n")

	fmt.Fprintf(sb, "var k %s\n", m.key.Name())
	existingVars = append(existingVars, "k")
	fmt.Fprintf(sb, "%s\n", m.key.decode("k", existingVars, q))

	fmt.Fprintf(sb, "var v %s\n", m.val.Name())
	existingVars = append(existingVars, "v")
	fmt.Fprintf(sb, "%s\n", m.val.decode("v", existingVars, q))

	fmt.Fprintf(sb, "%s[k] = v", varId)
	sb.WriteString("}\n")
	sb.WriteString("}\n") // end of block
	return sb.String()
}

// encode implements Type.
func (m mapType) encode(varId string, existingVars []string) string {
	sb := &strings.Builder{}
	// length
	fmt.Fprintf(sb, "{ // %s %s\n", varId, m.Name())
	fmt.Fprintf(sb, "var l int = len(%s)\n", varId)
	sb.WriteString(m.lenEnc.encode("uint64(l)", existingVars))
	existingVars = append(existingVars, "l")

	keyIt, valIt := generateKeyValueIteratorNames(existingVars)
	existingVars = append(existingVars, keyIt, valIt)

	// for loop
	fmt.Fprintf(sb, "for %s, %s := range %s {", keyIt, valIt, varId)
	sb.WriteString(m.key.encode(keyIt, existingVars))
	sb.WriteString(m.val.encode(valIt, existingVars))
	sb.WriteString("}\n") // end of for loop

	sb.WriteString("}\n") // end of block
	return sb.String()
}

var _ Type = structType{}

type structField2 struct {
	name string
	t    Type
}

type structType struct {
	fields            []structField2
	qualifiedTypeName string
	importSpec        importSpec
}

func (tr *typeResolver) newStructTypeT(apiName string, t types.Type, st *types.Struct, astExpr ast.Expr) (structType, error) {
	name, importSpec := tr.typeNameAndImport(t, astExpr)

	var structAst *ast.StructType
	if astExpr := tr.unwrapTypeAst(t, astExpr); astExpr != nil {
		var ok bool
		structAst, ok = astExpr.(*ast.StructType)
		if !ok {
			return structType{}, fmt.Errorf("slice's astExpression is not *ast.StructType, but %T of value %#[1]v", astExpr)
		}
	}

	fields := []structField2{}
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		log.Printf("%d: field: %q", i, f.Name())
		var fAst *ast.Field
		if structAst != nil {
			// todo: use the ast, if available
			// fAst = structAst.Fields.List[i]
			log.Printf("we have struct's ast, but we cannot handle it yet. fAst: %v", fAst)
		}
		ft, err := tr.newType(apiName, f.Type(), nil)
		if err != nil {
			return structType{}, fmt.Errorf("create Type for field %q: %w", f, err)
		}
		sf := structField2{name: f.Name(), t: ft}
		log.Printf("adding structField2{}: %#v", sf)
		fields = append(fields, sf)
	}

	return structType{
		qualifiedTypeName: name,
		fields:            fields,
		importSpec:        importSpec,
	}, nil
}

/*
// astExpr can be nil
func (g *generator) newStructType(astExpr ast.Expr, t types.Type) (structType, error) {
	var structAst *ast.StructType
	// var structTT *types.Struct
	named, ok := t.(*types.Named)
	if ok {
		log.Printf("struct type is named: %v, ast expr is %#v", named, astExpr)
		// named.Obj()
		typeSpec, err := g.findAstTypeSpec(named)
		if err == nil {
			// return structType{}, fmt.Errorf("findTypeSpec(): %w", err)
			log.Printf("typeSpec: %#v", typeSpec)
			structAst, ok = typeSpec.Type.(*ast.StructType)
			if !ok {
				return structType{}, fmt.Errorf("couldn't find *ast.StructType in named struct")
			} else {
				log.Printf("couldn't find typespec for %T. continuing", t)
				// now the structAst is nil
			}
		}

	} else {
		structAst, ok = astExpr.(*ast.StructType)
		if !ok {
			return structType{}, fmt.Errorf("slice's astExpression is not *ast.StructType, but %T of value %#[1]v", astExpr)
		}
	}

	structTT, ok := t.Underlying().(*types.Struct)
	if !ok {
		return structType{}, fmt.Errorf("failed to infer *types.Struct for our struct")
	}
	// log.Printf("structAst: %#v", structAst)

	qualifiedTypeName, err := g.qualifiedTypeName(astExpr)
	if err != nil {
		return structType{}, fmt.Errorf("qualifiedTypeName(%v): %w", astExpr, err)
	}

	// log.Printf("struct's qualified type name is: %q", qualifiedTypeName)
	fields := []structField2{}
	if structAst != nil {
		for i, field := range structAst.Fields.List {
			log.Printf("%d: ast field: %#v", i, field)
			typ, err := g.newType(&types.Slice{}, field.Type)
			if err != nil {
				return structType{}, fmt.Errorf("field's newType(%v): %w", field.Type, err)
			}
			for _, name := range field.Names {
				// log.Printf("field name %d: %q", i, name.String())
				fields = append(fields, structField2{name: name.String(), t: typ})
			}
			// log.Printf("field type: %#v", field.Type)
		}
	} else {
		log.Printf("we don't have ast. iterating over type")
		for i := 0; i < structTT.NumFields(); i++ {
			f := structTT.Field(i)
			log.Printf("processing field named %q  - %v", f.Name(), f)
		}
		panic("unfinished")
	}

	return structType{
		qualifiedTypeName: qualifiedTypeName,
		fields:            fields,
	}, nil
}
*/

// ImportSpecs implements Type.
func (s structType) ImportSpecs() importSpec {
	return s.importSpec
}

// Name implements Type.
func (s structType) Name() string {
	return s.qualifiedTypeName
}

// codeblock implements Type.
func (s structType) codeblock() string {
	return ""
}

// decode implements Type.
func (s structType) decode(varId string, existingVars []string, q *qualifier) string {
	sb := &strings.Builder{}
	for _, f := range s.fields {
		sb.WriteString(f.t.decode(varId+"."+f.name, existingVars, q))
	}
	return sb.String()
}

// encode implements Type.
func (s structType) encode(varId string, existingVars []string) string {
	sb := strings.Builder{}
	for _, f := range s.fields {
		sb.WriteString(f.t.encode(varId+"."+f.name, existingVars))
	}
	return sb.String()
}

func (g *generator) findAstTypeSpec(named *types.Named) (*ast.TypeSpec, error) {
	obj := named.Obj()
	if obj.Pkg() == nil {
		return nil, fmt.Errorf("no obj.Pkg() for named type %q", named)
	}
	pkg, err := g.findPackageForPackagePath(obj.Pkg().Path())
	if err != nil {
		return nil, fmt.Errorf("no pkg for file path: %q", obj.Pkg().Path())
	}
	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}

			// log.Printf("current genDecl: %v", gen.Specs)
			for _, spec := range gen.Specs {
				ts := spec.(*ast.TypeSpec)
				// log.Printf("looking at typespec: %v", ts.Name)
				if pkg.TypesInfo.Defs[ts.Name] == obj {
					return ts, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("findTypeSpec(): found no spec for obj: %#v", obj)
}

var _ Type = interfaceType{}

type interfaceType struct {
	name         string
	fncs         []ifaceFunc
	implTypeName string
	importSpec   importSpec
}

func (tr *typeResolver) newInterfaceTypeT(apiName string, t types.Type, it *types.Interface, astExpr ast.Expr) (interfaceType, error) {
	name, importSpec := tr.typeNameAndImport(t, astExpr)

	var ifaceAst *ast.InterfaceType

	if astExpr := tr.unwrapTypeAst(t, astExpr); astExpr != nil {
		var ok bool
		ifaceAst, ok = astExpr.(*ast.InterfaceType)
		if !ok {
			return interfaceType{}, fmt.Errorf("slice's astExpression is not *ast.ArrayType, but %T of value %#[1]v", astExpr)
		}
	}

	log.Printf("iface ast: %v", ifaceAst)

	fncs := []ifaceFunc{}
	for i := 0; i < it.NumMethods(); i++ {
		method := it.Method(i)
		sig := method.Type().(*types.Signature)

		// we only handle interfaces that return values.
		// we call them and store them
		if sig.Params().Len() != 0 {
			return interfaceType{}, fmt.Errorf("unexpectedly params are not nil. they are: %+v", sig.Params())
		}

		// field := ifaceAst.Methods.List[i]
		// log.Printf("field %d: %#v", i, field)

		results := make([]ifaceRtnVar2, 0, sig.Results().Len())
		for i := 0; i < sig.Results().Len(); i++ {
			v := sig.Results().At(i)
			typ, err := tr.newType(apiName, v.Type(), nil) // todo: get ast!
			if err != nil {
				return interfaceType{}, fmt.Errorf("newType for var %q: %w", v, err)
			}

			results = append(results, ifaceRtnVar2{
				name:          v.Name(),
				implParamName: fmt.Sprintf("_%s_%d_%s", method.Name(), i, v.Name()), //"_" + m.Name() + r.Name(),
				t:             typ})
		}

		fncs = append(fncs, ifaceFunc{
			funcName: method.Name(),
			results:  results,
		})
	}

	// we need to make the type unique within package, because we want each file to be self contained
	// it would be possible to use filename instead of apiName, but that would confuse file renaming. this will do for now
	var implTypeName string
	if _, isNamed := t.(*types.Named); isNamed {
		implTypeName = "_" + name + "_" + apiName + "_irpcInterfaceImpl"
	} else {
		sanitizedName := sanitizeInterfaceName(it)
		implTypeName = "_" + sanitizedName + "_" + apiName + "_irpcInterfaceImpl"
	}
	// log.Printf("implTypeName: %q", implTypeName)
	// log.Printf("interfaceType.name: %q", name)

	return interfaceType{
		name:         name,
		implTypeName: implTypeName,
		fncs:         fncs,
		importSpec:   importSpec,
	}, nil
}

// ImportSpecs implements Type.
func (i interfaceType) ImportSpecs() importSpec {
	return i.importSpec
}

// Name implements Type.
func (i interfaceType) Name() string {
	return i.name
}

// codeblock implements Type.
func (i interfaceType) codeblock() string {
	sb := &strings.Builder{}
	// type declaration
	fmt.Fprintf(sb, "type %s struct {\n", i.implTypeName)
	for _, f := range i.fncs {
		for _, v := range f.results {
			// fmt.Fprintf(sb, "%s %s\n", v.implParamName, v.rtnTypeName)
			fmt.Fprintf(sb, "%s %s\n", v.implParamName, v.t.Name())
		}
	}
	sb.WriteString("}\n")

	// fncs
	for _, f := range i.fncs {
		fmt.Fprintf(sb, "func (i %s)%s()(%s){\n", i.implTypeName, f.funcName, f.rtnParams())
		fmt.Fprintf(sb, "return %s\n", f.retParamsPrefixed("i."))
		sb.WriteString("}\n")
	}

	return sb.String()
}

// decode implements Type.
func (i interfaceType) decode(varId string, existingVars []string, q *qualifier) string {
	sb := &strings.Builder{}
	sb.WriteString("{\n") // separate block
	fmt.Fprintf(sb, `var isNil bool
	%s
	if isNil {
		%s = nil
	} else {
	`, boolEncoder.decode("isNil", existingVars, q), varId)

	fmt.Fprintf(sb, "var impl %s\n", i.implTypeName)
	for _, f := range i.fncs {
		fmt.Fprintf(sb, "{ // %s()\n", f.funcName)
		for _, v := range f.results {
			sb.WriteString(v.t.decode("impl."+v.implParamName, existingVars, q))
		}
		sb.WriteString("}\n")
	}
	fmt.Fprintf(sb, "%s = impl\n", varId)
	sb.WriteString("}\n") // else {
	sb.WriteString("}\n") // separate block

	return sb.String()
}

// encode implements Type.
func (i interfaceType) encode(varId string, existingVars []string) string {
	sb := &strings.Builder{}
	sb.WriteString("{\n") // separate block
	fmt.Fprintf(sb, `var isNil bool
	if %s == nil {
		isNil = true
	}
	%s
	`, varId, boolEncoder.encode("isNil", existingVars))
	sb.WriteString("if !isNil{\n")
	for _, f := range i.fncs {
		fmt.Fprintf(sb, "{ // %s()\n", f.funcName)
		for i, v := range f.results {
			sb.WriteString(v.implParamName)
			if i != len(f.results)-1 {
				sb.WriteString(",")
			}
		}
		fmt.Fprintf(sb, ":= %s.%s()\n", varId, f.funcName)
		for _, v := range f.results {
			sb.WriteString(v.t.encode(v.implParamName, existingVars))
		}
		sb.WriteString("}\n")
	}
	sb.WriteString("}\n") // if !isNil
	sb.WriteString("}\n") // separate block

	return sb.String()
}
