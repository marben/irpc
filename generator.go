package main

import (
	"errors"
	"fmt"
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

func newGenerator(fd rpcFileDesc) (generator, error) {
	imports := newOrderedSet[string]()
	services := []serviceGenerator{}
	clients := []clientGenerator{}
	paramStructs := []paramStructGenerator{}
	for _, iface := range fd.ifaces {
		methods := []methodGenerator{}
		for i, m := range iface.methods {
			mg, err := newMethodGenerator(iface.name(), i, m)
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
	fmt.Fprintf(sb, "func (s %s)Serialize(e *irpc.Encoder) error {\n", sg.typeName)
	if len(sg.params) > 0 {
		for _, p := range sg.params {
			sb.WriteString(p.enc.encode("s." + p.structFieldName))
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
			sb.WriteString(p.enc.decode("s." + p.structFieldName))
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

func newVarField(apiName string, p rpcParam) (varField, error) {
	structFieldName := fmt.Sprintf("Param%d_%s", p.pos, p.name)
	enc, err := varEncoder(apiName, p.typeAndValue.Type)
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
	index     int
	req, resp paramStructGenerator
	imports   []string
}

func newMethodGenerator(ifaceName string, index int, m rpcMethod) (methodGenerator, error) {
	imports := newOrderedSet[string]()
	reqStructTypeName := "_Irpc_" + ifaceName + m.name + "Req"
	reqFields := []varField{}
	for _, param := range m.params {
		rf, err := newVarField(ifaceName, param)
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
		rf, err := newVarField(ifaceName, result)
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
		index:   index,
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
	imports := []string{fmtImport}

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
		 	var args %[1]s
		 	if err := args.Deserialize(d); err != nil {
		 		return nil, err
		 	}
			return func() irpc.Serializable {
				// EXECUTE
				var resp %[2]s
				%[3]s = s.impl.%[4]s(%[5]s)
				return resp
			}, nil
		}, nil
		 `, m.req.typeName, m.resp.typeName, m.resp.paramListPrefixed("resp."), m.name, m.req.paramListPrefixed("args."))
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
		id irpc.RegisteredServiceId
	}

	func %[2]s(endpoint *irpc.Endpoint) (*%[1]s, error) {
		id, err := endpoint.RegisterClient([]byte("%[3]s"))
		if err != nil {
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
			fmt.Fprintf(b, "%s: %s,\n", p.structFieldName, p.name)
		}
		fmt.Fprintf(b, "}\n") // end struct assignment

		// func call
		fmt.Fprintf(b, "var %s %s\n", respVarName, m.resp.typeName)
		s := `if err := %s.endpoint.CallRemoteFunc(%[1]s.id, %d, %s, &%s); err != nil {
			panic(err)
		}
		`
		fmt.Fprintf(b, s, cg.fncReceiverName, m.index, reqVarName, respVarName)

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
