package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"io"
	"log"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// id len specifies how long our service id's will be. currently the max is 32 bytes as we are using sha256 to generate them
// actual id to negotiate between endpoints desn't have to be full lenght (currently it's only 4 bytes)
const idLen = 32

type generator struct {
	inputPkg     *packages.Package
	services     []serviceGenerator
	paramStructs []paramStructGenerator
	typesInfo    *types.Info // can we simply use the inputPkg?
	imports      orderedSet[importSpec]
}

func newGenerator(filename string) (*generator, error) {
	absFilePath, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("filepath.Abs(): %w", err)
	}

	dir := filepath.Dir(absFilePath)

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedDeps | packages.NeedImports | packages.NeedSyntax | packages.NeedTypesInfo |
			packages.NeedFiles | packages.NeedName | packages.NeedCompiledGoFiles | packages.NeedExportFile | packages.NeedSyntax |
			packages.NeedModule,
		Dir: dir,
	}

	// we need to load all the files in directory, otherwise we get "command-line-arguments" as pkg paths
	// todo: maybe we need to use ./... or base it at the root of our module, to get all the deps? need to test/figure out
	packages, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, fmt.Errorf("packages.Load(): %w", err)
	}

	// packages.Load() seems to be designed to parse multiple files (passed in go command style (./... etc))
	// we only care about one file though, therefore it should always be the first in the array in following code

	if len(packages) != 1 {
		return nil, fmt.Errorf("unexpectedly %d packages returned for file %q", len(packages), filename)
	}

	pkg := packages[0]

	gen := &generator{
		typesInfo: pkg.TypesInfo,
		inputPkg:  pkg,
		imports:   newOrderedSet[importSpec](),
	}

	fileAst, err := findASTForFile(pkg, filename)
	if err != nil {
		return nil, fmt.Errorf("couldn't find ast for given file %s", filename)
	}

	for _, decl := range fileAst.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if iface, ok := ts.Type.(*ast.InterfaceType); ok {
				gen.addInterface(ts.Name.String(), iface)
			}
		}
	}

	return gen, nil
}

func (g *generator) addInterface(ifaceName string, astIface *ast.InterfaceType) error {
	methods := []methodGenerator{}
	for i, methodField := range astIface.Methods.List {
		if len(methodField.Names) == 0 {
			return fmt.Errorf("method of interface %q has no name", ifaceName)
		}
		methodName := methodField.Names[0].Name

		astFuncType, ok := methodField.Type.(*ast.FuncType)
		if !ok {
			return fmt.Errorf("*ast.Field %v is not *ast.FuncType", methodField)
		}

		params, err := g.loadRpcParamList(ifaceName, astFuncType.Params.List)
		if err != nil {
			return fmt.Errorf("params list load for %s: %w", methodName, err)
		}

		var results []rpcParam
		if astFuncType.Results != nil {
			results, err = g.loadRpcParamList(ifaceName, astFuncType.Results.List)
			if err != nil {
				return fmt.Errorf("results list load for %s: %w", methodName, err)
			}
		}

		mg, err := g.newMethodGenerator(ifaceName, i, methodName, params, results)
		if err != nil {
			return fmt.Errorf("method '%s' generator: %w", methodName, err)
		}
		methods = append(methods, mg)
	}

	if err := g.addServiceGenerator(ifaceName, methods); err != nil {
		return fmt.Errorf("addServiceGenerator(): %w", err)
	}
	return nil
}

func (g *generator) newMethodGenerator(ifaceName string, index int, methodName string, params, results []rpcParam) (methodGenerator, error) {
	// REQUEST
	reqStructTypeName := "_Irpc_" + ifaceName + methodName + "Req"

	reqFieldNames := make(map[string]struct{}, len(params))
	for _, rf := range params {
		// make sure this name is not allocated by another param id
		reqFieldNames[rf.name] = struct{}{}
	}

	reqParams := []funcParam{}
	for _, param := range params {
		rp, err := g.newRequestParam(param, reqFieldNames)
		if err != nil {
			return methodGenerator{}, fmt.Errorf("newRequestParam '%s': %w", param.name, err)
		}
		reqParams = append(reqParams, rp)
	}
	req, err := g.addParamStructGenerator(reqStructTypeName, reqParams)
	if err != nil {
		return methodGenerator{}, fmt.Errorf("new req struct for method '%s': %w", methodName, err)
	}

	// RESPONSE
	respStructTypeName := "_Irpc_" + ifaceName + methodName + "Resp"
	respParams := []funcParam{}
	for _, result := range results {
		rp, err := g.newResultParam(result)
		if err != nil {
			return methodGenerator{}, fmt.Errorf("newResultParam '%s': %w", result.name, err)
		}
		// we don't support returning context.Context
		if rp.isContext() {
			return methodGenerator{}, fmt.Errorf("unsupported context.Context as return value for varfiled: %s - %s", ifaceName, methodName)
		}
		respParams = append(respParams, rp)
	}
	resp, err := g.addParamStructGenerator(respStructTypeName, respParams)
	if err != nil {
		return methodGenerator{}, fmt.Errorf("new resp struct for method '%s': %w", methodName, err)
	}

	// context
	// we currently only support one or no context var
	// multiple ctx vars could be combined, but it doesn't make much sense and i cannot be bothered atm
	ctxParams := []funcParam{}
	for _, p := range req.params {
		if p.isContext() {
			ctxParams = append(ctxParams, p)
		}
	}
	var ctxVarName string
	switch len(ctxParams) {
	case 0:
		ctxVarName = "context.Background()"
	case 1:
		ctxVarName = ctxParams[0].identifier
	default:
		return methodGenerator{}, fmt.Errorf("%s - %s : cannot have more than one context parameter", ifaceName, methodName)
	}

	return methodGenerator{
		name:   methodName,
		index:  index,
		req:    req,
		resp:   resp,
		ctxVar: ctxVarName,
	}, nil
}

func (g *generator) loadRpcParamList(interfaceName string, list []*ast.Field) ([]rpcParam, error) {
	params := []rpcParam{}
	for pos, field := range list {
		// try to get qualifier, if there is one
		var qualifier string
		if selExpr, ok := field.Type.(*ast.SelectorExpr); ok {
			if ident, ok := selExpr.X.(*ast.Ident); ok {
				qualifier = ident.Name
			}

		}

		tv, ok := g.typesInfo.Types[field.Type]
		if !ok {
			fmt.Printf("couldn't determine fileld's %v type and value", field)
			continue
		}

		if field.Names == nil {
			// parameter doesn't have name, just a type (typically function returns)
			param, err := g.newRpcParam(interfaceName, pos, "", tv.Type, qualifier)
			if err != nil {
				return nil, fmt.Errorf("newRpcParam on pos %d: %w", pos, err)
			}
			params = append(params, param)
		} else {
			for _, name := range field.Names {
				// obj := typesInfo.ObjectOf(name)
				param, err := g.newRpcParam(interfaceName, pos, name.Name, tv.Type, qualifier)
				if err != nil {
					return nil, fmt.Errorf("newRpcParam on pos %d: %w", pos, err)
				}
				params = append(params, param)
			}
		}
	}
	return params, nil
}

func (g *generator) newRpcParam(interfaceName string, position int, name string, typ types.Type, qualifier string) (rpcParam, error) {
	tDesc, err := g.newTypeDesc(interfaceName, typ, qualifier)
	if err != nil {
		return rpcParam{}, fmt.Errorf("newTypeDesc(): %w", err)
	}

	if named, ok := typ.(*types.Named); ok {
		if ok {
			obj := named.Obj()
			if pkg := obj.Pkg(); pkg != nil {
				var alias string
				if qualifier != pkg.Name() {
					alias = qualifier
				}
				g.addImport(importSpec{alias: alias, path: pkg.Path()})
			}
		}
	}

	return rpcParam{
		pos:   position,
		name:  name,
		tDesc: tDesc,
	}, nil
}

func (g *generator) newTypeDesc(apiName string, t types.Type, qualifier string) (typeDesc, error) {
	/*
		var enc encoder
		switch t := t.Underlying().(type) {
		case *types.Basic:
			switch t.Kind() {
			case types.Bool:
				enc = boolEncoder
			case types.Int:
				enc = intEncoder
			case types.Uint:
				enc = uintEncoder
			case types.Int8:
				enc = int8Encoder
			case types.Uint8: // serves 'types.Byte' as well
				enc = uint8Encoder
			case types.Int16:
				enc = int16Encoder
			case types.Uint16:
				enc = uint16Encoder
			case types.Int32: // serves 'types.Rune' as well
				enc = int32Encoder
			case types.Uint32:
				enc = uint32Encoder
			case types.Int64:
				enc = int64Encoder
			case types.Uint64:
				enc = uint64Encoder
			case types.Float32:
				enc = float32Encoder
			case types.Float64:
				enc = float64Encoder
			case types.String:
				enc = stringEncoder
			default:
				return typeDesc{}, fmt.Errorf("unsupported basic type '%s'", t.Name())
			}
		}
	*/

	qf := func(pkg *types.Package) string {
		return qualifier
	}

	encGen, err := newEncoderResolver(apiName, qf, g.typesInfo)
	if err != nil {
		return typeDesc{}, fmt.Errorf("newEncoderResolver(): %w", err)
	}

	enc, err := encGen.varEncoder(t)
	if err != nil {
		return typeDesc{}, fmt.Errorf("varEncoder(): %w", err)
	}

	return typeDesc{
		tt:                t,
		qualifiedTypeName: types.TypeString(t, qf),
		enc:               enc,
	}, nil
}

// if hash is nil, we generate service id with empty hash
//   - this is used during first run of generator
func (g *generator) generate(w io.Writer, hash []byte) error {
	codeBlocks := newOrderedSet[string]()

	// SERVICES
	for _, service := range g.services {
		codeBlocks.add(service.serviceCode(hash))
	}

	// CLIENTS
	for _, service := range g.services {
		codeBlocks.add(service.clientCode(hash))
	}

	// PARAM STRUCTS
	for _, p := range g.paramStructs {
		// we don't generate empty types (even though the generator is capable of generating them)
		// we use irpcgen.Empty(Ser/Deser) instead
		if !p.isEmpty() {
			codeBlocks.add(p.code())
			for _, e := range p.encoders() {
				codeBlocks.add(e.codeblock())
			}
		}
	}

	// GENERATE
	rawOutput := g.genRaw(codeBlocks)

	// FORMAT
	formatted, err := format.Source([]byte(rawOutput))
	if err != nil {
		log.Println("formatting failed. writing raw code to output file anyway")
		if _, err := w.Write([]byte(rawOutput)); err != nil {
			return fmt.Errorf("writing unformatted code to file: %w", err)
		}
	}

	if _, err := w.Write([]byte(formatted)); err != nil {
		return fmt.Errorf("copy of generated code to file: %w", err)
	}

	return nil
}

type paramStructGenerator struct {
	typeName string
	params   []funcParam
}

func (g *generator) addParamStructGenerator(typeName string, params []funcParam) (paramStructGenerator, error) {
	sg := paramStructGenerator{
		typeName: typeName,
		params:   params,
	}
	g.paramStructs = append(g.paramStructs, sg)

	return sg, nil
}

func (sg paramStructGenerator) code() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "type %s struct{\n", sg.typeName)
	for _, p := range sg.params {
		if p.isContext() {
			// we comment out context var as it is not filled anyway
			sb.WriteString("//")
		}
		sb.WriteString(p.structFieldName + " " + p.typeName + "\n")
	}
	sb.WriteString("\n}\n")
	sb.WriteString(sg.serializeFunc() + "\n")
	sb.WriteString(sg.deserializeFunc())

	return sb.String()
}

func (sg paramStructGenerator) isEmpty() bool {
	return len(sg.params) == 0
}

func (sg paramStructGenerator) serializeFunc() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "func (s %s)Serialize(e *irpcgen.Encoder) error {\n", sg.typeName)
	if len(sg.params) > 0 {
		for _, p := range sg.params {
			sb.WriteString(p.enc.encode("s."+p.structFieldName, nil))
		}
	}
	sb.WriteString("return nil\n}")

	return sb.String()
}

func (sg paramStructGenerator) deserializeFunc() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "func (s *%s)Deserialize(d *irpcgen.Decoder) error {\n", sg.typeName)
	if len(sg.params) > 0 {
		for _, p := range sg.params {
			sb.WriteString(p.enc.decode("s."+p.structFieldName, nil))
		}
	}
	sb.WriteString("return nil\n}")

	return sb.String()
}

// comma separated list of variable names and types. ex: "a int, b float64"
func (sg paramStructGenerator) funcCallParams() string {
	b := &strings.Builder{}
	for i, v := range sg.params {
		fmt.Fprintf(b, "%s %s", v.identifier, v.typeName)
		if i != len(sg.params)-1 {
			b.WriteString(",")
		}
	}
	return b.String()
}

// generates comma separated list of variable names each prefixed with 'prefix'
// for use in returning deserialized struct values
// as if they were normal return values from client function
// ex: return resp.Param0_v, resp.Param1_err
func (sg paramStructGenerator) paramListPrefixed(prefix string) string {
	sb := &strings.Builder{}
	for i, p := range sg.params {
		fmt.Fprintf(sb, "%s%s", prefix, p.structFieldName)
		if i != len(sg.params)-1 {
			sb.WriteString(",")
		}
	}
	return sb.String()
}

func (sg paramStructGenerator) isLastTypeError() bool {
	if len(sg.params) == 0 {
		return false
	}

	last := sg.params[len(sg.params)-1]
	return last.typeName == "error"
	// log.Printf("last type: %+v", last.param.typ.String())
}

func (sg paramStructGenerator) encoders() []encoder {
	encs := []encoder{}
	for _, p := range sg.params {
		encs = append(encs, p.enc)
	}

	return encs
}

// represents a variable in param struct which in turn represents function parameter/return value
type funcParam struct {
	name            string // original name as defined in the interface. can be ""
	identifier      string // identifier we use for this field. it's either param.name or if there is none, we generate it
	typeName        string
	structFieldName string
	enc             encoder
}

func (g *generator) addImport(imps ...importSpec) {
	for _, imp := range imps {
		if imp.path == g.inputPkg.PkgPath {
			// we don't want to import our own directory
			continue
		}
		g.imports.add(imp)
	}
}

// requestParamNames contains all parameter names, including ours
// if our parameter doesn't have a name, we will create one, making suere, we don't overlap with named parameters
func (g *generator) newRequestParam(p rpcParam, requestParamNames map[string]struct{}) (funcParam, error) {
	// figure out a unique id
	id := p.name
	if id == "" || id == "_" {
		id = fmt.Sprintf("p%d", p.pos)
		for {
			if _, exists := requestParamNames[id]; exists {
				id += "_"
			} else {
				break
			}
		}
	}
	requestParamNames[id] = struct{}{}

	return funcParam{
		name:            p.name,
		identifier:      id,
		typeName:        p.tDesc.qualifiedTypeName,
		enc:             p.tDesc.enc,
		structFieldName: fmt.Sprintf("Param%d_%s", p.pos, id),
	}, nil
}

func (g *generator) newResultParam(p rpcParam) (funcParam, error) {
	sFieldName := fmt.Sprintf("Param%d", p.pos)
	if p.name != "" {
		sFieldName += "_" + p.name
	}

	return funcParam{
		name:            p.name,
		identifier:      p.name,
		typeName:        p.tDesc.qualifiedTypeName,
		enc:             p.tDesc.enc,
		structFieldName: sFieldName,
	}, nil
}

func (g *generator) genRaw(codeBlocks orderedSet[string]) string {
	sb := &strings.Builder{}
	// HEADER
	headerStr := `// Code generated by irpc generator; DO NOT EDIT
	package %s
	`
	fmt.Fprintf(sb, headerStr, g.inputPkg.Name)

	// IMPORTS
	if g.imports.len() != 0 {
		sb.WriteString("import(\n")
		for _, imp := range g.imports.ordered {
			fmt.Fprintf(sb, "%s \"%s\"\n", imp.alias, imp.path)
		}
		sb.WriteString("\n)\n")
	}

	// UNIQUE BLOCKS
	for _, b := range codeBlocks.ordered {
		fmt.Fprintf(sb, "\n%s\n", b)
	}

	return sb.String()
}

// returns true if field is of type context.Context
func (vf funcParam) isContext() bool {
	return vf.typeName == "context.Context"
}

type methodGenerator struct {
	name      string
	index     int
	req, resp paramStructGenerator
	ctxVar    string // context used for method call (either there is context param, or we use context.Background() )
}

func (mg methodGenerator) executorFuncCode() string {
	if mg.resp.isEmpty() {
		return fmt.Sprintf(`func(ctx context.Context) irpcgen.Serializable {
				// EXECUTE
				s.impl.%[2]s(%[3]s)
				return irpcgen.EmptySerializable{}
			}`, mg.resp.typeName, mg.name, mg.requestParamsListPrefixed("args.", "ctx"))
	}

	return fmt.Sprintf(`func(ctx context.Context) irpcgen.Serializable {
				// EXECUTE
				var resp %[1]s
				%[2]s = s.impl.%[3]s(%[4]s)
				return resp
			}`, mg.resp.typeName, mg.resp.paramListPrefixed("resp."), mg.name, mg.requestParamsListPrefixed("args.", "ctx"))
}

// creates method call list with each var prefixed with 'prefix'
// replaces any parameter of type context.Context with 'ctxVarName'
func (mg methodGenerator) requestParamsListPrefixed(prefix, ctxVarName string) string {
	sb := &strings.Builder{}
	for i, p := range mg.req.params {
		if p.isContext() {
			sb.WriteString(ctxVarName)
		} else {
			fmt.Fprintf(sb, "%s%s", prefix, p.structFieldName)
		}
		if i != len(mg.req.params) {
			sb.WriteString(",")
		}
	}
	return sb.String()
}

type serviceGenerator struct {
	ifaceName string
	methods   []methodGenerator
}

func (g *generator) addServiceGenerator(ifaceName string, methods []methodGenerator) error {
	g.addImport(fmtImport)
	g.addImport(irpcGenImport)
	if len(methods) > 0 {
		// every FuncExecutor uses context
		g.addImport(contextImport)
	}

	sg2 := serviceGenerator{
		ifaceName: ifaceName,
		methods:   methods,
	}
	g.services = append(g.services, sg2)

	return nil
}

func (sg serviceGenerator) serviceId(hash []byte) []byte {
	// we use empty hash when file hash was not provided - during the dry run
	// this allows us to change idLen while keeping the common part of hash the same - not really useful but nice to have
	if hash == nil {
		return nil
	}

	return generateServiceIdHash(hash, sg.ifaceName, idLen)
}

func (sg serviceGenerator) serviceCode(hash []byte) string {
	serviceTypeName := sg.ifaceName + "IRpcService"
	w := &strings.Builder{}

	// type definition
	fmt.Fprintf(w, `type %s struct{
		impl %s
		id []byte
	}
	`, serviceTypeName, sg.ifaceName)

	// constructor
	fmt.Fprintf(w, `func %s (impl %s) *%[3]s {
		return &%[3]s{
			impl:impl,
			id: %s,
		}
	}
	`, generateStructConstructorName(serviceTypeName), sg.ifaceName, serviceTypeName, byteSliceLiteral(sg.serviceId(hash)))

	// Id() func
	fmt.Fprintf(w, `func (s *%s) Id() []byte {
		return s.id
	}
	`, serviceTypeName)

	// Call func call swith
	fmt.Fprintf(w, `func (s *%s) GetFuncCall(funcId irpcgen.FuncId) (irpcgen.ArgDeserializer, error){
		switch funcId {
			`, serviceTypeName)

	for _, m := range sg.methods {
		fmt.Fprintf(w, "case %d: // %s\n", m.index, m.name)
		fmt.Fprintf(w, "return func(d *irpcgen.Decoder) (irpcgen.FuncExecutor, error) {\n")

		// deserialize, if not empty
		if !m.req.isEmpty() {
			fmt.Fprintf(w, `// DESERIALIZE
		 	var args %s
		 	if err := args.Deserialize(d); err != nil {
		 		return nil, err
		 	}
			`, m.req.typeName)
		}

		fmt.Fprintf(w, `return %s, nil
		}, nil
		 `, m.executorFuncCode())
	}
	fmt.Fprintf(w, `default:
			return nil, fmt.Errorf("function '%%d' doesn't exist on service '%%s'", funcId, s.Id())
		}
	}
	`)
	return w.String()
}

func (sg serviceGenerator) clientCode(hash []byte) string {
	clientTypeName := sg.ifaceName + "IRpcClient"
	fncReceiverName := "_c" // todo: must not collide with any of fnc variable names

	b := &strings.Builder{}
	// type definition
	fmt.Fprintf(b, `
	type %[1]s struct {
		endpoint irpcgen.Endpoint
		id []byte
	}

	func %[2]s(endpoint irpcgen.Endpoint) (*%[1]s, error) {
		id := %[3]s
		if err := endpoint.RegisterClient(id); err != nil {
			return nil, fmt.Errorf("register failed: %%w", err)
		}
		return &%[1]s{endpoint: endpoint, id: id}, nil
	}
	`, clientTypeName, generateStructConstructorName(clientTypeName), byteSliceLiteral(sg.serviceId(hash)))

	// func calls
	for _, m := range sg.methods {
		// func header
		fmt.Fprintf(b, "func(%s *%s)%s(%s)(%s){\n", fncReceiverName, clientTypeName, m.name, m.req.funcCallParams(), m.resp.funcCallParams())

		allVars := append(m.req.params, m.resp.params...)

		// request
		var reqVarName string
		if m.req.isEmpty() {
			reqVarName = "irpcgen.EmptySerializable{}"
		} else {
			reqVarName = generateUniqueVarname("req", allVars)
			// request construction
			fmt.Fprintf(b, "var %s = %s {\n", reqVarName, m.req.typeName)
			for _, p := range m.req.params {
				if p.isContext() {
					// we skip contexts, as they are treated special
					b.WriteString("// ")
				}
				fmt.Fprintf(b, "%s: %s,\n", p.structFieldName, p.identifier)
			}
			fmt.Fprintf(b, "}\n") // end struct assignment
		}

		// response
		var respVarName string
		if m.resp.isEmpty() {
			respVarName = "irpcgen.EmptyDeserializable{}"
		} else {
			respVarName = generateUniqueVarname("resp", allVars)
			fmt.Fprintf(b, "var %s %s\n", respVarName, m.resp.typeName)
		}

		// func call
		fmt.Fprintf(b, "if err := %s.endpoint.CallRemoteFunc(%s,%[1]s.id, %[3]d, %s, &%s); err != nil {\n", fncReceiverName, m.ctxVar, m.index, reqVarName, respVarName)
		if m.resp.isLastTypeError() {
			// declare zero var, because i don't know, how to directly instantiate zero values
			if len(m.resp.params) > 1 {
				fmt.Fprintf(b, "var zero %s\n", m.resp.typeName)
			}
			fmt.Fprintf(b, "return ")
			for i := 0; i < len(m.resp.params)-1; i++ {
				p := m.resp.params[i]
				fmt.Fprintf(b, "%s.%s,", "zero", p.structFieldName)
			}
			fmt.Fprintf(b, "err\n")
		} else {
			fmt.Fprintf(b, "panic(err) // to avoid panic, make your func return error and regenerate the code\n")
		}
		fmt.Fprintf(b, "}\n")

		// return values
		fmt.Fprintf(b, "return ")
		for i, f := range m.resp.params {
			fmt.Fprintf(b, "%s.%s", respVarName, f.structFieldName)
			if i != len(m.resp.params)-1 {
				fmt.Fprintf(b, ",")
			}
		}
		fmt.Fprintf(b, "\n")

		fmt.Fprintf(b, "}\n") // end of func
	}

	return b.String()
}
