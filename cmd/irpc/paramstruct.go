package main

import (
	"fmt"
	"strings"
)

type paramStructGenerator struct {
	structName string
	params     []genParam
}

func newReqRespStructsGenerator(apiName, methodName string, reqParams, respParams []rpcParam) (req, resp paramStructGenerator, err error) {
	var paramNames varNames = make([]string, 0, len(reqParams)+len(respParams))

	for _, p := range reqParams {
		paramNames.addVarName(p.name)
	}
	for _, p := range respParams {
		paramNames.addVarName(p.name)
	}

	// req
	reqStructParams := make([]genParam, 0, len(reqParams))
	for _, p := range reqParams {
		id := p.name
		if id == "" || id == "_" {
			id = paramNames.generateUniqueVarName(fmt.Sprintf("p%d", p.pos))
		}
		reqStructParams = append(reqStructParams, genParam{
			identifier:      id,
			structFieldName: id,
			typ:             p.typ,
		})
	}

	req = paramStructGenerator{
		structName: "_irpc_" + apiName + "_" + methodName + "Req",
		params:     reqStructParams,
	}

	// resp
	respStructParams := make([]genParam, 0, len(respParams))
	for _, p := range respParams {
		id := p.name
		if id == "" || id == "_" {
			id = paramNames.generateUniqueVarName(fmt.Sprintf("p%d", p.pos))
		}
		fp := genParam{
			identifier:      p.name,
			structFieldName: id,
			typ:             p.typ,
		}

		if fp.isContext() {
			return paramStructGenerator{}, paramStructGenerator{}, fmt.Errorf("unsupported context.Context as return value for varfiled: %s - %s", apiName, methodName)
		}

		respStructParams = append(respStructParams, fp)
	}

	resp = paramStructGenerator{
		structName: "_irpc_" + apiName + "_" + methodName + "Resp",
		params:     respStructParams,
	}

	return req, resp, nil
}

func (sg paramStructGenerator) code(q *qualifier) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "type %s struct{\n", sg.structName)
	for _, p := range sg.params {
		if p.isContext() {
			// we comment out context var as it is not filled anyway
			sb.WriteString("//")
		}
		sb.WriteString(p.structFieldName + " " + p.typ.name(q) + "\n")
	}
	sb.WriteString("\n}\n")
	sb.WriteString(sg.serializeFunc(q) + "\n")
	sb.WriteString(sg.deserializeFunc(q))

	return sb.String()
}

func (sg paramStructGenerator) isEmpty() bool {
	return len(sg.params) == 0
}

func (sg paramStructGenerator) serializeFunc(q *qualifier) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "func (s %s)Serialize(e *irpcgen.Encoder) error {\n", sg.structName)
	if len(sg.params) > 0 {
		for _, p := range sg.params {
			varId := "s." + p.structFieldName
			encFunc := p.typ.genEncFunc(q)
			if encFunc == "" {
				// some types skip encoding (currently it's context)
				continue
			}
			fmt.Fprintf(sb, "if err := %s(e, %s); err != nil{\n", encFunc, varId)
			if p.identifier != "" {
				fmt.Fprintf(sb, "return fmt.Errorf(\"serialize \\\"%s\\\" of type %s: %%w\", err)\n", p.identifier, p.typ.name(q.copy()))
			} else {
				fmt.Fprintf(sb, "return fmt.Errorf(\"serialize type %s: %%w\", err)\n", p.typ.name(q.copy()))
			}
			sb.WriteString("}\n")
		}
	}
	sb.WriteString("return nil\n}") // end of Serialize

	return sb.String()
}

func (sg paramStructGenerator) deserializeFunc(q *qualifier) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "func (s *%s)Deserialize(d *irpcgen.Decoder) error {\n", sg.structName)
	if len(sg.params) > 0 {
		for _, p := range sg.params {
			varId := "s." + p.structFieldName
			decFunc := p.typ.genDecFunc(q)
			if decFunc == "" {
				continue
			}
			fmt.Fprintf(sb, "if err := %s(d, &%s); err != nil {\n", decFunc, varId)
			if p.identifier != "" {
				fmt.Fprintf(sb, "return fmt.Errorf(\"deserialize %s of type %s: %%w\", err)\n", p.identifier, p.typ.name(q.copy()))
			} else {
				fmt.Fprintf(sb, "return fmt.Errorf(\"deserialize type %s: %%w\", err)\n", p.typ.name(q.copy()))
			}
			sb.WriteString("}\n")
		}
	}
	sb.WriteString("return nil\n}")

	return sb.String()
}

// comma separated list of variable names and types. ex: "a int, b float64"
func (sg paramStructGenerator) funcCallParams(q *qualifier) string {
	b := &strings.Builder{}
	for i, v := range sg.params {
		fmt.Fprintf(b, "%s %s", v.identifier, v.typ.name(q))
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

func (sg paramStructGenerator) isLastTypeError(q *qualifier) bool {
	if len(sg.params) == 0 {
		return false
	}

	// todo: reimplement with better error recognition?
	last := sg.params[len(sg.params)-1]
	return last.typ.name(q) == "error"
	// log.Printf("last type: %+v", last.param.typ.String())
}

func (sg paramStructGenerator) types() []Type {
	types := []Type{}
	for _, p := range sg.params {
		types = append(types, p.typ)
	}

	return types
}

// genParam descibes a function parameter and it's representation in our req/resp structure
type genParam struct {
	// identifier. eg: 'a' in func(a int){}. if it's omitted or '_', we generate a name, because we must allow it in our client
	// unique within func definition
	identifier string

	// structFieldName eq: "a" in reqStruct.a	stores the value of our parameter (the argument)
	structFieldName string

	// typ contains all we need for encoding/decoding of this type
	typ Type
}

// returns true if field is of type context.Context
func (fp genParam) isContext() bool {
	if _, ok := fp.typ.(contextType); ok {
		return true
	}
	return false
}
