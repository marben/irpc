package main

import (
	"fmt"
	"go/ast"
	"strings"
)

type apiGenerator struct {
	apiName          string
	goDoc            string
	serviceIdVarName string
	// methods callable over the network: the user's methods without channels,
	// plus the wire methods of the methods with channels (streams)
	methods []methodGenerator
	streams []streamMethodGenerator
}

func newApiGenerator(tr typeResolver, apiName string, astIface *ast.InterfaceType, godocCg *ast.CommentGroup) (apiGenerator, error) {
	methods := []methodGenerator{}
	streams := []streamMethodGenerator{}

	// the user's methods keep their index (funcId) - the additional wire methods of streams are numbered after them
	nextAuxIndex := len(astIface.Methods.List)

	for i, methodField := range astIface.Methods.List {
		method, err := newMethodGenerator(tr, apiName, methodField, i)
		if err != nil {
			return apiGenerator{}, fmt.Errorf("newMethodGenerator(): %w", err)
		}

		if !method.hasChannels() {
			methods = append(methods, method)
			continue
		}

		stream, err := newStreamMethodGenerator(tr, apiName, method, &nextAuxIndex)
		if err != nil {
			pos := tr.srcPkg.Fset.Position(methodField.Pos())
			return apiGenerator{}, fmt.Errorf("%s: %s.%s: %w", pos, apiName, method.name, err)
		}
		streams = append(streams, stream)
		methods = append(methods, stream.wireMethods()...)
	}

	return apiGenerator{
		apiName:          apiName,
		goDoc:            godocFromAstCommentGroup(godocCg),
		serviceIdVarName: fmt.Sprintf("_%sIrpcId", apiName),
		methods:          methods,
		streams:          streams,
	}, nil
}

func (ag apiGenerator) hasStreams() bool {
	return len(ag.streams) > 0
}

func (ag apiGenerator) streamsTypeName() string {
	return "_irpc_" + ag.apiName + "Streams"
}

func (ag apiGenerator) paramStructs() []paramStructGenerator {
	paramStructs := make([]paramStructGenerator, 0, len(ag.methods)*2)
	for _, method := range ag.methods {
		paramStructs = append(paramStructs, method.req, method.resp)
	}
	return paramStructs
}

func (ag apiGenerator) serviceIdVarDefinition(hash []byte) string {
	return fmt.Sprintf("var %s = irpcgen.ServiceId(%s)", ag.serviceIdVarName, serviceIdLiteralFromBytes(ag.serviceId(hash)))
}

func (ag apiGenerator) clientTypeName() string {
	return ag.apiName + "IrpcClient"
}

func (ag apiGenerator) serviceTypeName() string {
	return ag.apiName + "IrpcService"
}

func (ag apiGenerator) clientCode(q *qualifier) string {
	sb := &strings.Builder{}

	// GoDoc comment
	fmt.Fprintf(sb, "// %s implements [%s] interface. It by forwards calls over network to [%s] that provides the implementation.\n", ag.clientTypeName(), ag.apiName, ag.serviceTypeName())
	if ag.goDoc != "" {
		sb.WriteString("// \n")
		sb.WriteString(ag.goDoc)
	}

	// type definition
	if !ag.hasStreams() {
		fmt.Fprintf(sb, `type %[1]s struct {
			endpoint irpcgen.Endpoint
		}

		func %[2]s(endpoint irpcgen.Endpoint) (*%[1]s, error) {
			if err := endpoint.RegisterClient(%[3]s); err != nil {
				return nil, fmt.Errorf("register failed: %%w", err)
			}
			return &%[1]s{endpoint: endpoint}, nil
		}
		`, ag.clientTypeName(), generateStructConstructorName(ag.clientTypeName()), ag.serviceIdVarName)
	} else {
		q.addUsedImport(contextImport)
		fmt.Fprintf(sb, `type %[1]s struct {
			endpoint irpcgen.Endpoint
			// unlimited sends the calls of streams: they must not wait for one of the slots
			// that the other calls of the same stream are holding
			unlimited irpcgen.UnlimitedFuncCaller
		}

		func %[2]s(endpoint irpcgen.Endpoint) (*%[1]s, error) {
			unlimited, ok := endpoint.(irpcgen.UnlimitedFuncCaller)
			if !ok {
				return nil, fmt.Errorf("endpoint %%T doesn't implement irpcgen.UnlimitedFuncCaller, required by the streams of %[4]s", endpoint)
			}
			if err := endpoint.RegisterClient(%[3]s); err != nil {
				return nil, fmt.Errorf("register failed: %%w", err)
			}
			return &%[1]s{endpoint: endpoint, unlimited: unlimited}, nil
		}
		`, ag.clientTypeName(), generateStructConstructorName(ag.clientTypeName()), ag.serviceIdVarName, ag.apiName)
	}

	// func calls
	for _, m := range ag.methods {
		var allVarIds varNames
		for _, p := range m.req.params {
			allVarIds.addVarName(p.identifier)
		}
		for _, p := range m.resp.params {
			allVarIds.addVarName(p.identifier)
		}

		fncReceiverName := allVarIds.generateUniqueVarName("_c")

		// func header
		if m.isWire {
			fmt.Fprintf(sb, "// %s is a stream call (see [irpcgen.StreamSession])\n", m.name)
		} else {
			fmt.Fprintf(sb, "// %s implements [%s]\n//\n", m.name, ag.apiName)
			sb.WriteString(m.goDoc)
		}
		fmt.Fprintf(sb, "func(%s *%s)%s(%s)(%s){\n", fncReceiverName, ag.clientTypeName(), m.name, m.req.funcCallParams(q), m.resp.funcCallParams(q))

		// request
		var reqVarName string
		if m.req.isEmpty() {
			reqVarName = "irpcgen.EmptySerializable{}"
		} else {
			reqVarName = allVarIds.generateUniqueVarName("req")
			// request construction
			fmt.Fprintf(sb, "var %s = %s {\n", reqVarName, m.req.structName)
			for _, p := range m.req.params {
				if p.isContext() {
					// we skip contexts, as they are treated special
					sb.WriteString("// ")
				}
				fmt.Fprintf(sb, "%s: %s,\n", p.structFieldName, p.identifier)
			}
			fmt.Fprintf(sb, "}\n") // end struct assignment
		}

		// response
		var respVarName string
		if m.resp.isEmpty() {
			respVarName = "irpcgen.EmptyDeserializable{}"
		} else {
			respVarName = allVarIds.generateUniqueVarName("resp")
			fmt.Fprintf(sb, "var %s %s\n", respVarName, m.resp.structName)
		}

		// func call
		callFunc := "endpoint.CallRemoteFunc"
		if m.unlimited || m.isWire {
			// every call of a stream is unlimited on the client: we only limit ourselves, and a stream's
			// calls must not wait for slots held by the other calls of the same stream
			callFunc = "unlimited.CallRemoteFuncUnlimited"
		}
		fmt.Fprintf(sb, "if err := %s.%s(%s,%s, %d, %s, &%s); err != nil {\n", fncReceiverName, callFunc, m.ctxVar, ag.serviceIdVarName, m.index, reqVarName, respVarName)
		if m.resp.isLastTypeError(q) {
			// declare zero var, because i don't know, how to directly instantiate zero values
			if len(m.resp.params) > 1 {
				fmt.Fprintf(sb, "var zero %s\n", m.resp.structName)
			}
			fmt.Fprintf(sb, "return ")
			for i := 0; i < len(m.resp.params)-1; i++ {
				p := m.resp.params[i]
				fmt.Fprintf(sb, "%s.%s,", "zero", p.structFieldName)
			}
			fmt.Fprintf(sb, "err\n")
		} else {
			fmt.Fprintf(sb, "panic(err) // to avoid panic, make your func return error and regenerate irpc code\n")
		}
		fmt.Fprintf(sb, "}\n")

		// return values
		if !m.resp.isEmpty() {
			fmt.Fprintf(sb, "return ")
			for i, f := range m.resp.params {
				fmt.Fprintf(sb, "%s.%s", respVarName, f.structFieldName)
				if i != len(m.resp.params)-1 {
					fmt.Fprintf(sb, ",")
				}
			}
			fmt.Fprintf(sb, "\n")
		}

		fmt.Fprintf(sb, "}\n") // end of func
	}

	for _, s := range ag.streams {
		sb.WriteString(s.clientCode(q, ag.clientTypeName()))
	}

	return sb.String()
}

func (ag apiGenerator) serviceCode(q *qualifier) string {
	w := &strings.Builder{}

	if !ag.hasStreams() {
		// type definition
		fmt.Fprintf(w, `// %s provides [%[2]s] interface over irpc 
		type %[1]s struct{
			impl %s
		}
		`, ag.serviceTypeName(), ag.apiName)

		// constructor
		fmt.Fprintf(w, `// %s returns new [irpcgen.Service] forwarding [%[2]s] network calls to impl
		func %[1]s (impl %[2]s) *%[3]s {
			return &%[3]s{
				impl:impl,
			}
		}
		`, generateStructConstructorName(ag.serviceTypeName()), ag.apiName, ag.serviceTypeName())
	} else {
		q.addUsedImport(contextImport)
		fmt.Fprintf(w, `// %s provides [%[2]s] interface over irpc 
		type %[1]s struct{
			impl %[2]s
			streams *%[3]s // implements the calls of the methods with channels
		}

		// %[4]s returns new [irpcgen.Service] forwarding [%[2]s] network calls to impl
		func %[4]s (impl %[2]s) *%[1]s {
			return &%[1]s{
				impl:impl,
				streams: &%[3]s{impl: impl},
			}
		}
		`, ag.serviceTypeName(), ag.apiName, ag.streamsTypeName(), generateStructConstructorName(ag.serviceTypeName()))

		// IsUnlimitedFunc
		// "case 1, // name\n 2: // name"
		unlimitedMethods := []methodGenerator{}
		for _, m := range ag.methods {
			if m.unlimited {
				unlimitedMethods = append(unlimitedMethods, m)
			}
		}
		unlimited := []string{}
		for i, m := range unlimitedMethods {
			sep := ","
			if i == len(unlimitedMethods)-1 {
				sep = ":"
			}
			unlimited = append(unlimited, fmt.Sprintf("%d%s // %s", m.index, sep, m.name))
		}
		fmt.Fprintf(w, `// IsUnlimitedFunc implements [irpcgen.UnlimitedFuncService] interface.
		// The calls of streams that can last for the whole stream don't take a worker slot.
		func (s *%s) IsUnlimitedFunc(funcId irpcgen.FuncId) bool {
			switch funcId {
			case %s
				return true
			default:
				return false
			}
		}
		`, ag.serviceTypeName(), strings.Join(unlimited, "\n"))
	}

	// Id() func
	fmt.Fprintf(w, `// Id implements [irpcgen.Service] interface.
	func (s *%s) Id() irpcgen.ServiceId {
		return %s
	}
	`, ag.serviceTypeName(), ag.serviceIdVarName)

	// Call func call switch
	fmt.Fprintf(w, `// GetFuncCall implements [irpcgen.Service] interface
	func (s *%s) GetFuncCall(funcId irpcgen.FuncId) (irpcgen.ArgDeserializer, error){
		switch funcId {
			`, ag.serviceTypeName())

	q.addUsedImport(irpcGenImport, fmtImport)

	for _, m := range ag.methods {
		fmt.Fprintf(w, "case %d: // %s\n", m.index, m.name)
		fmt.Fprintf(w, "return func(d *irpcgen.Decoder) (irpcgen.FuncExecutor, error) {\n")

		// deserialize, if not empty
		if !m.req.isEmpty() {
			fmt.Fprintf(w, `var args %s
		 	if err := args.Deserialize(d); err != nil {
		 		return nil, err
		 	}
			`, m.req.structName)
		}

		fmt.Fprintf(w, `return %s, nil
		}, nil
		 `, m.executorFuncCode(q))
	}
	fmt.Fprintf(w, `default:
			return nil, fmt.Errorf("function '%%d' doesn't exist on service '%%s'", funcId, s.Id())
		}
	}
	`)

	if ag.hasStreams() {
		// the struct implementing the wire methods of the streams
		fmt.Fprintf(w, "// %s implements the service side of the methods with channels of [%s]\n", ag.streamsTypeName(), ag.apiName)
		fmt.Fprintf(w, "type %s struct {\nimpl %s\n", ag.streamsTypeName(), ag.apiName)
		for _, s := range ag.streams {
			fmt.Fprintf(w, "%s irpcgen.StreamSessions[%s]\n", s.sessionsFieldName(), s.sessionTypeName())
		}
		w.WriteString("}\n\n")
		for _, s := range ag.streams {
			w.WriteString(s.serverCode(q, ag.streamsTypeName()))
		}
	}
	return w.String()
}

func (ag apiGenerator) serviceId(hash []byte) []byte {
	// we use empty hash when file hash was not provided - during the dry run
	// this allows us to change idLen while keeping the common part of hash the same - not really useful but nice to have
	if hash == nil {
		return nil
	}

	return generateServiceIdHash(hash, ag.apiName, generatedIdLen)
}
