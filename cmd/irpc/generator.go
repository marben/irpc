package main

import (
	"errors"
	"fmt"
	"go/types"
	"io"
	"log"
	"strings"
)

// id len specifies how long our service id's will be. currently the max is 32 bytes as we are using sha256 to generate them
// actual id to negotiate between endpoints desn't have to be full lenght (atm it's only 4 bytes)
const idLen = 32

type generator struct {
	fd       rpcFileDesc
	services []serviceGenerator
	clients  []clientGenerator
	imports  []string
	params   []paramStructGenerator
}

// if fileHash is nil, we generate service id with empty hash - this is only used during dry run (first run of generator)
func newGenerator(fd rpcFileDesc, q types.Qualifier, fileHash []byte) (generator, error) {
	imports := newOrderedSet[string]()
	services := []serviceGenerator{}
	clients := []clientGenerator{}
	paramStructs := []paramStructGenerator{}
	for _, iface := range fd.ifaces {
		methods := []methodGenerator{}
		for i, m := range iface.methods {
			mg, err := newMethodGenerator(iface.name(), i, m, q)
			if err != nil {
				return generator{}, fmt.Errorf("method '%s' generator: %w", m.name, err)
			}
			methods = append(methods, mg)
			imports.add(mg.imports...)
			paramStructs = append(paramStructs, mg.req, mg.resp)
		}

		// we use empty hash when file hash was not provided - during the dry run
		// this allows us to change idLen while keeping the common part of hash the same - not really useful but nice to have
		serviceId := []byte{}
		if fileHash != nil {
			serviceId = generateServiceIdHash(fileHash, iface.name(), idLen)
		}

		sg, err := newServiceGenerator(iface.name()+"IRpcService", iface.name(), methods, serviceId)
		if err != nil {
			return generator{}, fmt.Errorf("service generator for iface: %s: %w", iface.name(), err)
		}
		services = append(services, sg)

		cg, err := newClientGenerator(iface.name()+"IRpcClient", iface.name(), methods, serviceId)
		if err != nil {
			return generator{}, fmt.Errorf("client generator for iface: %s: %w", iface.name(), err)
		}
		clients = append(clients, cg)
	}

	return generator{
		fd:       fd,
		services: services,
		clients:  clients,
		imports:  imports.ordered,
		params:   paramStructs,
	}, nil
}

func (g generator) write(w io.Writer) error {
	genF := newGenFile(g.fd.packageName())
	genF.addImport(g.imports...)

	// SERVICES
	for _, service := range g.services {
		genF.addUniqueBlock(service.code())
		genF.addImport(service.imports...)
	}

	// CLIENTS
	for _, c := range g.clients {
		genF.addUniqueBlock(c.code())
		genF.addImport(c.imports...)
	}

	// PARAM STRUCTS
	for _, p := range g.params {
		// we don't generate empty types (even though the generator is capable of generating them)
		// we use irpcgen.Empty(Ser/Deser) instead
		if !p.isEmpty() {
			genF.addUniqueBlock(p.code())
			for _, e := range p.encoders() {
				genF.addUniqueBlock(e.codeblock())
			}
		}
	}

	file, err := genF.formatted()
	if err != nil {
		var formatErr *formattingErr
		if errors.As(err, &formatErr) {
			log.Println("formatting failed. writing raw code to output file anyway")
			if _, err := w.Write([]byte(formatErr.unformattedCode)); err != nil {
				return fmt.Errorf("failed to write unformatted code to file: %w", err)
			}
		}
		return err
	}

	if _, err := w.Write([]byte(file)); err != nil {
		return fmt.Errorf("copy of generated code to file: %w", err)
	}

	return nil
}

type paramStructGenerator struct {
	typeName string
	params   []funcParam
	imports  []string
}

func newParamStructGenerator(typeName string, params []funcParam) (paramStructGenerator, error) {
	imports := newOrderedSet[string]()
	for _, p := range params {
		imports.add(p.enc.imports()...)
	}

	return paramStructGenerator{
		typeName: typeName,
		params:   params,
		imports:  imports.ordered,
	}, nil
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

// like funcCallParams, but omit's param names, if they were not defined
func (sg paramStructGenerator) returnParams() string {
	b := &strings.Builder{}
	for i, v := range sg.params {
		fmt.Fprintf(b, "%s %s", v.name, v.typeName)
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

// requestParamNames contains all parameter names, including ours
// if our parameter doesn't have a name, we will create one, making suere, we don't overlap with named parameters
func newRequestParam(apiName string, p rpcParam, q types.Qualifier, requestParamNames map[string]struct{}) (funcParam, error) {
	enc, err := varEncoder(apiName, p.typ, q)
	if err != nil {
		return funcParam{}, fmt.Errorf("param field for type '%s': %w", p.typeName(q), err)
	}

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
		typeName:        p.typeName(q),
		enc:             enc,
		structFieldName: fmt.Sprintf("Param%d_%s", p.pos, id),
	}, nil
}

func newResultParam(apiName string, p rpcParam, q types.Qualifier) (funcParam, error) {
	enc, err := varEncoder(apiName, p.typ, q)
	if err != nil {
		return funcParam{}, fmt.Errorf("param field for type '%s': %w", p.typeName(q), err)
	}

	sFieldName := fmt.Sprintf("Param%d", p.pos)
	if p.name != "" {
		sFieldName += "_" + p.name
	}

	return funcParam{
		name:            p.name,
		identifier:      p.name,
		typeName:        p.typeName(q),
		enc:             enc,
		structFieldName: sFieldName,
	}, nil
}

// returns true if field is of type context.Context
func (vf funcParam) isContext() bool {
	return vf.typeName == "context.Context"
}

type methodGenerator struct {
	name      string
	index     int
	req, resp paramStructGenerator
	imports   []string
	ctxVar    string // context used for method call (either there is context param, or we use context.Background() )
}

func newMethodGenerator(ifaceName string, index int, m rpcMethod, q types.Qualifier) (methodGenerator, error) {
	imports := newOrderedSet[string]()

	// REQUEST
	reqStructTypeName := "_Irpc_" + ifaceName + m.name + "Req"

	reqFieldNames := make(map[string]struct{}, len(m.params))
	for _, rf := range m.params {
		// make sure this name is not allocated by another param id
		reqFieldNames[rf.name] = struct{}{}
	}

	reqFields := []funcParam{}
	for _, param := range m.params {
		vf, err := newRequestParam(ifaceName, param, q, reqFieldNames)
		if err != nil {
			return methodGenerator{}, fmt.Errorf("newVarField for param '%s': %w", param.name, err)
		}
		reqFields = append(reqFields, vf)
	}
	req, err := newParamStructGenerator(reqStructTypeName, reqFields)
	if err != nil {
		return methodGenerator{}, fmt.Errorf("new req struct for method '%s': %w", m.name, err)
	}
	imports.add(req.imports...)

	// RESPONSE
	respStructTypeName := "_Irpc_" + ifaceName + m.name + "Resp"
	respFields := []funcParam{}
	for _, result := range m.results {
		vf, err := newResultParam(ifaceName, result, q)
		if err != nil {
			return methodGenerator{}, fmt.Errorf("newVarField for param '%s': %w", result.name, err)
		}
		// we don't support returning context.Context
		if vf.isContext() {
			return methodGenerator{}, fmt.Errorf("unsupported context.Context as return value for varfiled: %s - %s", ifaceName, m.name)
		}
		respFields = append(respFields, vf)
	}
	resp, err := newParamStructGenerator(respStructTypeName, respFields)
	if err != nil {
		return methodGenerator{}, fmt.Errorf("new resp struct for method '%s': %w", m.name, err)
	}
	imports.add(resp.imports...)

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
		return methodGenerator{}, fmt.Errorf("%s - %s : cannot have more than one context parameter", ifaceName, m.name)
	}

	return methodGenerator{
		name:    m.name,
		index:   index,
		req:     req,
		resp:    resp,
		imports: imports.ordered,
		ctxVar:  ctxVarName,
	}, nil
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
	imports         []string
	ifaceName       string
	serviceTypeName string
	methods         []methodGenerator
	serviceId       []byte
}

func newServiceGenerator(serviceTypeName, ifaceTypeName string, methods []methodGenerator, serviceId []byte) (serviceGenerator, error) {
	imports := []string{fmtImport, contextImport}

	return serviceGenerator{
		ifaceName:       ifaceTypeName,
		serviceTypeName: serviceTypeName,
		methods:         methods,
		imports:         imports,
		serviceId:       serviceId,
	}, nil
}

func (sg serviceGenerator) code() string {
	sb := &strings.Builder{}

	// type definition
	fmt.Fprintf(sb, `type %s struct{
		impl %s
		id []byte
	}
	`, sg.serviceTypeName, sg.ifaceName)

	// constructor
	fmt.Fprintf(sb, `func %s (impl %s) *%[3]s {
		return &%[3]s{
			impl:impl,
			id: %s,
		}
	}
	`, generateStructConstructorName(sg.serviceTypeName), sg.ifaceName, sg.serviceTypeName, byteSliceLiteral(sg.serviceId))

	// Id() func
	fmt.Fprintf(sb, `func (s *%s) Id() []byte {
		return s.id
	}
	`, sg.serviceTypeName)

	// Call func call swith
	fmt.Fprintf(sb, `func (s *%s) GetFuncCall(funcId irpcgen.FuncId) (irpcgen.ArgDeserializer, error){
		switch funcId {
			`, sg.serviceTypeName)

	for _, m := range sg.methods {
		fmt.Fprintf(sb, "case %d: // %s\n", m.index, m.name)
		fmt.Fprintf(sb, "return func(d *irpcgen.Decoder) (irpcgen.FuncExecutor, error) {\n")

		// deserialize, if not empty
		if !m.req.isEmpty() {
			fmt.Fprintf(sb, `// DESERIALIZE
		 	var args %s
		 	if err := args.Deserialize(d); err != nil {
		 		return nil, err
		 	}
			`, m.req.typeName)
		}

		fmt.Fprintf(sb, `return %s, nil
		}, nil
		 `, m.executorFuncCode())
	}
	fmt.Fprintf(sb, `default:
			return nil, fmt.Errorf("function '%%d' doesn't exist on service '%%s'", funcId, s.Id())
		}
	}
	`)

	return sb.String()
}

type clientGenerator struct {
	typeName        string
	ifaceName       string
	methods         []methodGenerator
	imports         []string
	fncReceiverName string
	serviceId       []byte
}

func newClientGenerator(clientTypeName string, ifaceTypeName string, methods []methodGenerator, serviceId []byte) (clientGenerator, error) {
	imports := []string{irpcGenImport}
	return clientGenerator{
		typeName:        clientTypeName,
		ifaceName:       ifaceTypeName,
		methods:         methods,
		imports:         imports,
		fncReceiverName: "_c", // todo: must not collide with any of fnc variable names
		serviceId:       serviceId,
	}, nil
}

func (cg clientGenerator) code() string {
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
	`, cg.typeName, generateStructConstructorName(cg.typeName), byteSliceLiteral(cg.serviceId))

	// func calls
	for _, m := range cg.methods {
		// func header
		fmt.Fprintf(b, "func(%s *%s)%s(%s)(%s){\n", cg.fncReceiverName, cg.typeName, m.name, m.req.funcCallParams(), m.resp.funcCallParams())

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
		fmt.Fprintf(b, "if err := %s.endpoint.CallRemoteFunc(%s,%[1]s.id, %[3]d, %s, &%s); err != nil {\n", cg.fncReceiverName, m.ctxVar, m.index, reqVarName, respVarName)
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
