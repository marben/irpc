package main

import (
	"fmt"
	"go/ast"
	"log"
	"strings"
)

type apiGenerator struct {
	apiName string
	methods []methodGenerator
}

func newApiGenerator(g *generator, tr *typeResolver, apiName string, astIface *ast.InterfaceType) (*apiGenerator, error) {
	log.Printf("creating apiGenerator with apiName %q", apiName)
	methods := []methodGenerator{}
	for i, methodField := range astIface.Methods.List {
		method, err := newMethodGenerator(g, tr, apiName, methodField, i)
		if err != nil {
			return nil, fmt.Errorf("newMethodGenerator(): %w", err)
		}
		methods = append(methods, method)
	}

	return &apiGenerator{
		apiName: apiName,
		methods: methods,
	}, nil
}

func (ag *apiGenerator) paramStructs() []paramStructGenerator {
	paramStructs := make([]paramStructGenerator, 0, len(ag.methods)*2)
	for _, method := range ag.methods {
		paramStructs = append(paramStructs, method.req, method.resp)
	}
	return paramStructs
}

func (ag *apiGenerator) clientCode(hash []byte, q *qualifier) string {
	clientTypeName := ag.apiName + "IRpcClient"
	fncReceiverName := "_c" // todo: must not collide with any of fnc variable names

	q.addUsedImport(irpcGenImport, fmtImport) // todo: remove
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
	`, clientTypeName, generateStructConstructorName(clientTypeName), byteSliceLiteral(ag.serviceId(hash)))

	// func calls
	for _, m := range ag.methods {
		// func header
		fmt.Fprintf(b, "func(%s *%s)%s(%s)(%s){\n", fncReceiverName, clientTypeName, m.name, m.req.funcCallParams(q), m.resp.funcCallParams(q))

		var allVarIds varNameList
		for _, p := range m.req.params {
			allVarIds.addVarName(p.identifier)
		}
		for _, p := range m.resp.params {
			allVarIds.addVarName(p.identifier)
		}

		// request
		var reqVarName string
		if m.req.isEmpty() {
			reqVarName = "irpcgen.EmptySerializable{}"
		} else {
			reqVarName = allVarIds.generateUniqueVarName("req")
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
			// respVarName = generateUniqueVarnameForFuncParams("resp", allVarIds)
			respVarName = allVarIds.generateUniqueVarName("resp")
			fmt.Fprintf(b, "var %s %s\n", respVarName, m.resp.typeName)
		}

		// func call
		fmt.Fprintf(b, "if err := %s.endpoint.CallRemoteFunc(%s,%[1]s.id, %[3]d, %s, &%s); err != nil {\n", fncReceiverName, m.ctxVar, m.index, reqVarName, respVarName)
		if m.resp.isLastTypeError(q) {
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

func (ag *apiGenerator) serviceCode(hash []byte, q *qualifier) string {
	serviceTypeName := ag.apiName + "IRpcService"
	w := &strings.Builder{}

	// type definition
	fmt.Fprintf(w, `type %s struct{
		impl %s
		id []byte
	}
	`, serviceTypeName, ag.apiName)

	// constructor
	fmt.Fprintf(w, `func %s (impl %s) *%[3]s {
		return &%[3]s{
			impl:impl,
			id: %s,
		}
	}
	`, generateStructConstructorName(serviceTypeName), ag.apiName, serviceTypeName, byteSliceLiteral(ag.serviceId(hash)))

	// Id() func
	fmt.Fprintf(w, `func (s *%s) Id() []byte {
		return s.id
	}
	`, serviceTypeName)

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

func (ag *apiGenerator) serviceId(hash []byte) []byte {
	// we use empty hash when file hash was not provided - during the dry run
	// this allows us to change idLen while keeping the common part of hash the same - not really useful but nice to have
	if hash == nil {
		return nil
	}

	return generateServiceIdHash(hash, ag.apiName, idLen)
}
