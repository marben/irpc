package main

import (
	"fmt"
	"io"
	"strings"
)

type generator struct {
	fd       rpcFileDesc
	services []serviceGenerator
	clients  []clientGenerator
	imports  []string
	params   []paramStructGenerator
}

func newGenerator(fd rpcFileDesc) (generator, error) {
	imports := newOrderedSet[string]()
	services := []serviceGenerator{}
	clients := []clientGenerator{}
	paramStructs := []paramStructGenerator{}
	for _, iface := range fd.ifaces {
		methods := []methodGenerator{}
		for _, m := range iface.methods {
			mg, err := newMethodGenerator(iface.name(), m)
			if err != nil {
				return generator{}, fmt.Errorf("method '%s' generator: %w", m.name, err)
			}
			methods = append(methods, mg)
			imports.add(mg.imports...)
			paramStructs = append(paramStructs, mg.req, mg.resp)
		}

		serviceId := iface.name() + "RpcService"

		sg, err := newServiceGenerator(iface.name()+"RpcService", iface.name(), serviceId, methods)
		if err != nil {
			return generator{}, fmt.Errorf("service generator for iface: %s: %w", iface.name(), err)
		}
		services = append(services, sg)

		cg, err := newClientGenerator(iface.name()+"RpcClient", serviceId, methods)
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
	genF := newGenFile(g.fd.packageName)
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
		return fmt.Errorf("formatted output: %w", err)
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
	imports.add(ioImport) // every Serialize() Deserialize() needs io import
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
	def := fmt.Sprintf("type %s struct{\n", sg.typeName)
	for _, p := range sg.params {
		def += p.structFieldName + " " + p.typeName + "\n"
	}
	def += "\n}\n"
	def += sg.serializeFunc() + "\n"
	def += sg.deserializeFunc()

	return def
}

func (sg paramStructGenerator) serializeFunc() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "func (s %s)Serialize(w io.Writer) error {\n", sg.typeName)
	if len(sg.params) > 0 {
		for _, p := range sg.params {
			fmt.Fprintf(sb, "{ // %s\n", p.typeName)
			sb.WriteString(p.enc.encode("s." + p.structFieldName))
			sb.WriteString("}\n")
		}
	}
	sb.WriteString("return nil\n}")

	return sb.String()
}

func (sg paramStructGenerator) deserializeFunc() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "func (s *%s)Deserialize(r io.Reader) error {\n", sg.typeName)
	if len(sg.params) > 0 {
		for _, p := range sg.params {
			fmt.Fprintf(sb, "{ // %s\n", p.typeName)
			sb.WriteString(p.enc.decode("s." + p.structFieldName))
			sb.WriteString("}\n")
		}
	}
	sb.WriteString("return nil\n}")

	return sb.String()
}

// comma separated list of variable names and types. ex: "a int, b float64"
func (sg paramStructGenerator) funcCallParams() string {
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

func (sg paramStructGenerator) encoders() []encoder {
	encs := []encoder{}
	for _, p := range sg.params {
		encs = append(encs, p.enc)
	}

	return encs
}

type varField struct {
	name            string
	typeName        string
	structFieldName string
	enc             encoder
}

func newVarField(p rpcParam) (varField, error) {
	structFieldName := fmt.Sprintf("Param%d_%s", p.pos, p.name)
	enc, err := varEncoder(p.typeAndValue.Type)
	if err != nil {
		return varField{}, fmt.Errorf("param field for type '%s': %w", p.typeName, err)
	}

	return varField{
		name:            p.name,
		typeName:        p.typeName,
		structFieldName: structFieldName,
		enc:             enc,
	}, nil
}

type methodGenerator struct {
	name      string
	req, resp paramStructGenerator
	imports   []string
}

func newMethodGenerator(ifaceName string, m rpcMethod) (methodGenerator, error) {
	imports := newOrderedSet[string]()
	reqStructTypeName := "_Irpc_" + ifaceName + m.name + "Req"
	reqFields := []varField{}
	for _, param := range m.params {
		rf, err := newVarField(param)
		if err != nil {
			return methodGenerator{}, fmt.Errorf("newVarField for param '%s': %w", param.name, err)
		}
		reqFields = append(reqFields, rf)
	}
	req, err := newParamStructGenerator(reqStructTypeName, reqFields)
	if err != nil {
		return methodGenerator{}, fmt.Errorf("new req struct for method '%s': %w", m.name, err)
	}
	imports.add(req.imports...)

	respStructTypeName := "_Irpc_" + ifaceName + m.name + "Resp"
	respFields := []varField{}
	for _, result := range m.results {
		rf, err := newVarField(result)
		if err != nil {
			return methodGenerator{}, fmt.Errorf("newVarField for param '%s': %w", result.name, err)
		}
		respFields = append(respFields, rf)
	}
	resp, err := newParamStructGenerator(respStructTypeName, respFields)
	if err != nil {
		return methodGenerator{}, fmt.Errorf("new resp struct for method '%s': %w", m.name, err)
	}
	imports.add(resp.imports...)

	return methodGenerator{
		name:    m.name,
		req:     req,
		resp:    resp,
		imports: imports.ordered,
	}, nil
}

type serviceGenerator struct {
	imports         []string
	ifaceName       string
	serviceTypeName string
	serviceId       string
	methods         []methodGenerator
}

func newServiceGenerator(serviceTypeName, ifaceTypeName, serviceId string, methods []methodGenerator) (serviceGenerator, error) {
	imports := []string{bytesImport, fmtImport}

	if len(methods) > 0 {
		// every method's call function creates a bytes.Buffer
		imports = append(imports, bytesImport)
	}

	return serviceGenerator{
		ifaceName:       ifaceTypeName,
		serviceTypeName: serviceTypeName,
		serviceId:       serviceId,
		methods:         methods,
		imports:         imports,
	}, nil
}

func (sg serviceGenerator) code() string {
	funcReceiverName := "s"

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
	fmt.Fprintf(sb, `func (%s) Id() string {
		return "%s"
	}
	`, sg.serviceTypeName, sg.serviceId)

	// Call func call swith
	fmt.Fprintf(sb, `func (%s *%s) CallFunc(funcName string, args []byte) ([]byte, error){
		switch funcName {
			`, funcReceiverName, sg.serviceTypeName)
	for _, m := range sg.methods {
		fmt.Fprintf(sb, `case "%[1]s":
			return %[2]s.call%[1]s(args)
			`, m.name, funcReceiverName)
	}
	fmt.Fprintf(sb, `default:
			return nil, fmt.Errorf("function '%%s' doesn't exist on service '%%s'", funcName, %s.Id())
		}
	}
	`, funcReceiverName)

	// methods
	for _, m := range sg.methods {
		fmt.Fprintf(sb, `func (%[7]s *%[1]s) call%[2]s(params []byte)([]byte, error) {
			r := bytes.NewBuffer(params)
			var req %s
			if err := req.Deserialize(r); err != nil {
				return nil, fmt.Errorf("failed to deserialize %[2]s: %%w", err)
			}
			var resp %[4]s
			%[5]s = %[7]s.impl.%[2]s(%[6]s)
			b := bytes.NewBuffer(nil)
			err := resp.Serialize(b)
			if err != nil {
				return nil, fmt.Errorf("response serialization failed: %%w", err)
			}
			return b.Bytes(), nil
		}
		`, sg.serviceTypeName, m.name, m.req.typeName, m.resp.typeName, m.resp.paramListPrefixed("resp."), m.req.paramListPrefixed("req."), funcReceiverName)
	}

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
	}

	func %[2]s(endpoint *irpc.Endpoint) *%[1]s {
		return &%[1]s{endpoint: endpoint}
	}
	`, cg.typeName, generateStructConstructorName(cg.typeName))

	// func calls
	for _, m := range cg.methods {
		// func header
		fmt.Fprintf(b, "func(%s *%s)%s(%s)(%s){\n", cg.fncReceiverName, cg.typeName, m.name, m.req.funcCallParams(), m.resp.funcCallParams())

		// request construction
		fmt.Fprintf(b, "var req = %s {\n", m.req.typeName)
		for _, p := range m.req.params {
			fmt.Fprintf(b, "%s: %s,\n", p.structFieldName, p.name)
		}
		fmt.Fprintf(b, "}\n") // end struct assignment

		// func call
		fmt.Fprintf(b, "var resp %s\n", m.resp.typeName)
		s := `if err := %s.endpoint.CallRemoteFunc("%s", "%s", req, &resp); err != nil {
			panic(err)
		}
		`
		fmt.Fprintf(b, s, cg.fncReceiverName, cg.serviceId, m.name)

		// return values
		fmt.Fprintf(b, "return ")
		for i, f := range m.resp.params {
			fmt.Fprintf(b, "resp.%s", f.structFieldName)
			if i != len(m.resp.params)-1 {
				fmt.Fprintf(b, ",")
			}
		}
		fmt.Fprintf(b, "\n")

		fmt.Fprintf(b, "}\n") // end of func
	}

	return b.String()
}
