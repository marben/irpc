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
	methods          []methodGenerator
}

func newApiGenerator(tr typeResolver, apiName string, astIface *ast.InterfaceType, godocCg *ast.CommentGroup) (apiGenerator, error) {
	methods := []methodGenerator{}
	for i, methodField := range astIface.Methods.List {
		method, err := newMethodGenerator(tr, apiName, methodField, i)
		if err != nil {
			return apiGenerator{}, fmt.Errorf("newMethodGenerator(): %w", err)
		}
		methods = append(methods, method)
	}

	return apiGenerator{
		apiName:          apiName,
		goDoc:            godocFromAstCommentGroup(godocCg),
		serviceIdVarName: fmt.Sprintf("_%sIrpcId", apiName),
		methods:          methods,
	}, nil
}

func (ag apiGenerator) paramStructs() []paramStructGenerator {
	paramStructs := make([]paramStructGenerator, 0, len(ag.methods)*2)
	for _, method := range ag.methods {
		paramStructs = append(paramStructs, method.req, method.resp)
	}
	return paramStructs
}

func (ag apiGenerator) serviceIdVarDefinition(hash []byte) string {
	return fmt.Sprintf("var %s = %s", ag.serviceIdVarName, byteSliceLiteral(ag.serviceId(hash)))
}

func (ag apiGenerator) clientCode(q *qualifier) string {
	clientTypeName := ag.apiName + "IrpcClient"
	sb := &strings.Builder{}

	// GoDoc comment
	fmt.Fprintf(sb, "// %s implements %s\n", clientTypeName, ag.apiName)
	if ag.goDoc != "" {
		sb.WriteString("// \n")
		sb.WriteString(ag.goDoc)
	}

	// type definition
	fmt.Fprintf(sb, `type %[1]s struct {
		endpoint irpcgen.Endpoint
	}

	func %[2]s(endpoint irpcgen.Endpoint) (*%[1]s, error) {
		if err := endpoint.RegisterClient(%[3]s); err != nil {
			return nil, fmt.Errorf("register failed: %%w", err)
		}
		return &%[1]s{endpoint: endpoint}, nil
	}
	`, clientTypeName, generateStructConstructorName(clientTypeName), ag.serviceIdVarName)

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
		sb.WriteString(m.goDoc)
		fmt.Fprintf(sb, "func(%s *%s)%s(%s)(%s){\n", fncReceiverName, clientTypeName, m.name, m.req.funcCallParams(q), m.resp.funcCallParams(q))

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
		fmt.Fprintf(sb, "if err := %s.endpoint.CallRemoteFunc(%s,%s, %d, %s, &%s); err != nil {\n", fncReceiverName, m.ctxVar, ag.serviceIdVarName, m.index, reqVarName, respVarName)
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

	return sb.String()
}

func (ag apiGenerator) serviceCode(q *qualifier) string {
	w := &strings.Builder{}

	serviceTypeName := ag.apiName + "IrpcService"

	// type definition
	fmt.Fprintf(w, `type %s struct{
		impl %s
	}
	`, serviceTypeName, ag.apiName)

	// constructor
	fmt.Fprintf(w, `func %s (impl %s) *%[3]s {
		return &%[3]s{
			impl:impl,
		}
	}
	`, generateStructConstructorName(serviceTypeName), ag.apiName, serviceTypeName)

	// Id() func
	fmt.Fprintf(w, `func (s *%s) Id() []byte {
		return %s
	}
	`, serviceTypeName, ag.serviceIdVarName)

	// Call func call switch
	fmt.Fprintf(w, `func (s *%s) GetFuncCall(funcId irpcgen.FuncId) (irpcgen.ArgDeserializer, error){
		switch funcId {
			`, serviceTypeName)

	q.addUsedImport(irpcGenImport, fmtImport)

	for _, m := range ag.methods {
		fmt.Fprintf(w, "case %d: // %s\n", m.index, m.name)
		fmt.Fprintf(w, "return func(d *irpcgen.Decoder) (irpcgen.FuncExecutor, error) {\n")

		// deserialize, if not empty
		if !m.req.isEmpty() {
			fmt.Fprintf(w, `// DESERIALIZE
		 	var args %s
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
