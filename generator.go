package main

import (
	"errors"
	"fmt"
	"go/types"
	"io"
	"log"
	"strings"
)

type generator struct {
	fd       rpcFileDesc
	services []serviceGenerator
	clients  []clientGenerator
	imports  []string
	params   []paramStructGenerator
}

func newGenerator(fd rpcFileDesc, q types.Qualifier) (generator, error) {
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

		serviceId := iface.name() + "IRpcService"

		sg, err := newServiceGenerator(iface.name()+"IRpcService", iface.name(), serviceId, methods)
		if err != nil {
			return generator{}, fmt.Errorf("service generator for iface: %s: %w", iface.name(), err)
		}
		services = append(services, sg)

		cg, err := newClientGenerator(iface.name()+"IRpcClient", serviceId, methods)
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
		genF.addUniqueBlock(p.code())
		for _, e := range p.encoders() {
			genF.addUniqueBlock(e.codeblock())
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
	params   []varField
	imports  []string
}

func newParamStructGenerator(typeName string, params []varField) (paramStructGenerator, error) {
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
		sb.WriteString(p.structFieldName() + " " + p.typeName() + "\n")
	}
	sb.WriteString("\n}\n")
	sb.WriteString(sg.serializeFunc() + "\n")
	sb.WriteString(sg.deserializeFunc())

	return sb.String()
}

func (sg paramStructGenerator) serializeFunc() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "func (s %s)Serialize(e *irpc.Encoder) error {\n", sg.typeName)
	if len(sg.params) > 0 {
		for _, p := range sg.params {
			sb.WriteString(p.enc.encode("s."+p.structFieldName(), nil))
		}
	}
	sb.WriteString("return nil\n}")

	return sb.String()
}

func (sg paramStructGenerator) deserializeFunc() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "func (s *%s)Deserialize(d *irpc.Decoder) error {\n", sg.typeName)
	if len(sg.params) > 0 {
		for _, p := range sg.params {
			sb.WriteString(p.enc.decode("s."+p.structFieldName(), nil))
		}
	}
	sb.WriteString("return nil\n}")

	return sb.String()
}

// comma separated list of variable names and types. ex: "a int, b float64"
func (sg paramStructGenerator) funcCallParams() string {
	b := &strings.Builder{}
	for i, v := range sg.params {
		fmt.Fprintf(b, "%s %s", v.name(), v.typeName())
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
		fmt.Fprintf(sb, "%s%s", prefix, p.structFieldName())
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
	return last.param.typ.String() == "error"
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
type varField struct {
	param rpcParam
	enc   encoder
	q     types.Qualifier
}

func newVarField(apiName string, p rpcParam, q types.Qualifier) (varField, error) {
	enc, err := varEncoder(apiName, p.typ, q)
	if err != nil {
		return varField{}, fmt.Errorf("param field for type '%s': %w", p.typeName(q), err)
	}

	return varField{
		param: p,
		enc:   enc,
		q:     q,
	}, nil
}

func (vf varField) structFieldName() string {
	return fmt.Sprintf("Param%d_%s", vf.param.pos, vf.param.name)
}

func (vf varField) name() string {
	return vf.param.name
}

func (vf varField) typeName() string {
	return vf.param.typeName(vf.q)
}

// returns true if field is of type context.Context
func (vf varField) isContext() bool {
	return vf.typeName() == "context.Context"
}

type methodGenerator struct {
	name      string
	index     int
	req, resp paramStructGenerator
	imports   []string
	q         types.Qualifier
	ctxVar    string // context used for method call (either there is context param, or we use context.Background() )
}

func newMethodGenerator(ifaceName string, index int, m rpcMethod, q types.Qualifier) (methodGenerator, error) {
	imports := newOrderedSet[string]()
	reqStructTypeName := "_Irpc_" + ifaceName + m.name + "Req"
	reqFields := []varField{}
	for _, param := range m.params {
		vf, err := newVarField(ifaceName, param, q)
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

	respStructTypeName := "_Irpc_" + ifaceName + m.name + "Resp"
	respFields := []varField{}
	for _, result := range m.results {
		vf, err := newVarField(ifaceName, result, q)
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
	ctxParams := []varField{}
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
		ctxVarName = ctxParams[0].name()
	default:
		return methodGenerator{}, fmt.Errorf("%s - %s : cannot have more than one context parameter", ifaceName, m.name)
	}

	return methodGenerator{
		name:    m.name,
		index:   index,
		req:     req,
		resp:    resp,
		imports: imports.ordered,
		q:       q,
		ctxVar:  ctxVarName,
	}, nil
}

func (mg methodGenerator) executorFuncCode() string {
	return fmt.Sprintf(`func(ctx context.Context) irpc.Serializable {
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
			fmt.Fprintf(sb, "%s%s", prefix, p.structFieldName())
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
	serviceId       string
	methods         []methodGenerator
}

func newServiceGenerator(serviceTypeName, ifaceTypeName, serviceId string, methods []methodGenerator) (serviceGenerator, error) {
	imports := []string{fmtImport, contextImport}

	return serviceGenerator{
		ifaceName:       ifaceTypeName,
		serviceTypeName: serviceTypeName,
		serviceId:       serviceId,
		methods:         methods,
		imports:         imports,
	}, nil
}

func (sg serviceGenerator) code() string {
	sb := &strings.Builder{}

	// type definition
	fmt.Fprintf(sb, `type %s struct{
		impl %s
	}
	`, sg.serviceTypeName, sg.ifaceName)

	// constructor
	fmt.Fprintf(sb, `func %s (impl %s) *%[3]s {
		return &%[3]s{impl:impl}
	}
	`, generateStructConstructorName(sg.serviceTypeName), sg.ifaceName, sg.serviceTypeName)

	// Id() func
	fmt.Fprintf(sb, `func (%s) Hash() []byte {
		return []byte("%s")
	}
	`, sg.serviceTypeName, sg.serviceId)

	// Call func call swith
	fmt.Fprintf(sb, `func (s *%s) GetFuncCall(funcId irpc.FuncId) (irpc.ArgDeserializer, error){
		switch funcId {
			`, sg.serviceTypeName)

	for _, m := range sg.methods {
		fmt.Fprintf(sb, "case %d: // %s\n", m.index, m.name)
		fmt.Fprintf(sb, `return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
		 	var args %s
		 	if err := args.Deserialize(d); err != nil {
		 		return nil, err
		 	}
			return %s, nil
		}, nil
		 `, m.req.typeName, m.executorFuncCode())
	}
	fmt.Fprintf(sb, `default:
			return nil, fmt.Errorf("function '%%d' doesn't exist on service '%%s'", funcId, string(s.Hash()))
		}
	}
	`)

	return sb.String()
}

type clientGenerator struct {
	typeName        string
	serviceId       string
	methods         []methodGenerator
	imports         []string
	fncReceiverName string
}

func newClientGenerator(clientTypeName string, serviceId string, methods []methodGenerator) (clientGenerator, error) {
	imports := []string{irpcImport}
	return clientGenerator{
		typeName:        clientTypeName,
		serviceId:       serviceId,
		methods:         methods,
		imports:         imports,
		fncReceiverName: "_c", // todo: must not collide with any of fnc variable names
	}, nil
}

func (cg clientGenerator) code() string {
	b := &strings.Builder{}
	// type definition
	fmt.Fprintf(b, `
	type %[1]s struct {
		endpoint *irpc.Endpoint
		id string
	}

	func %[2]s(endpoint *irpc.Endpoint) (*%[1]s, error) {
		id := "%[3]s"
		if err := endpoint.RegisterClient(id); err != nil {
			return nil, fmt.Errorf("register failed: %%w", err)
		}
		return &%[1]s{endpoint: endpoint, id: id}, nil
	}
	`, cg.typeName, generateStructConstructorName(cg.typeName), cg.serviceId)

	// func calls
	for _, m := range cg.methods {
		// func header
		fmt.Fprintf(b, "func(%s *%s)%s(%s)(%s){\n", cg.fncReceiverName, cg.typeName, m.name, m.req.funcCallParams(), m.resp.funcCallParams())

		allVars := append(m.req.params, m.resp.params...)
		reqVarName := generateUniqueVarname("req", allVars)
		respVarName := generateUniqueVarname("resp", allVars)
		// request construction
		fmt.Fprintf(b, "var %s = %s {\n", reqVarName, m.req.typeName)
		for _, p := range m.req.params {
			if p.isContext() {
				// we skip contexts, as they are treated special
				b.WriteString("// ")
			}
			fmt.Fprintf(b, "%s: %s,\n", p.structFieldName(), p.name())
		}
		fmt.Fprintf(b, "}\n") // end struct assignment

		// response
		fmt.Fprintf(b, "var %s %s\n", respVarName, m.resp.typeName)

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
				fmt.Fprintf(b, "%s.%s,", "zero", p.structFieldName())
			}
			fmt.Fprintf(b, "err\n")
		} else {
			fmt.Fprintf(b, "panic(err) // to avoid panic, make your func return error and regenerate the code\n")
		}
		fmt.Fprintf(b, "}\n")

		// return values
		fmt.Fprintf(b, "return ")
		for i, f := range m.resp.params {
			fmt.Fprintf(b, "%s.%s", respVarName, f.structFieldName())
			if i != len(m.resp.params)-1 {
				fmt.Fprintf(b, ",")
			}
		}
		fmt.Fprintf(b, "\n")

		fmt.Fprintf(b, "}\n") // end of func
	}

	return b.String()
}
