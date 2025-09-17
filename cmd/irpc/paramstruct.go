package main

import (
	"fmt"
	"strings"
)

type paramStructGenerator struct {
	typeName string
	params   []funcParam
}

func newParamStructGenerator(typeName string, params []funcParam) (paramStructGenerator, error) {
	return paramStructGenerator{
		typeName: typeName,
		params:   params,
	}, nil
}

func (sg paramStructGenerator) code(q *qualifier) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "type %s struct{\n", sg.typeName)
	for _, p := range sg.params {
		if p.isContext() {
			// we comment out context var as it is not filled anyway
			sb.WriteString("//")
		}
		sb.WriteString(p.structFieldName + " " + p.typ.Name(q) + "\n")
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
	fmt.Fprintf(sb, "func (s %s)Serialize(e *irpcgen.Encoder) error {\n", sg.typeName)
	if len(sg.params) > 0 {
		for _, p := range sg.params {
			sb.WriteString(p.typ.encode("s."+p.structFieldName, nil, q))
		}
	}
	sb.WriteString("return nil\n}")

	return sb.String()
}

func (sg paramStructGenerator) deserializeFunc(q *qualifier) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "func (s *%s)Deserialize(d *irpcgen.Decoder) error {\n", sg.typeName)
	if len(sg.params) > 0 {
		for _, p := range sg.params {
			sb.WriteString(p.typ.decode("s."+p.structFieldName, nil, q))
		}
	}
	sb.WriteString("return nil\n}")

	return sb.String()
}

// comma separated list of variable names and types. ex: "a int, b float64"
func (sg paramStructGenerator) funcCallParams(q *qualifier) string {
	b := &strings.Builder{}
	for i, v := range sg.params {
		fmt.Fprintf(b, "%s %s", v.identifier, v.typ.Name(q))
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
	return last.typ.Name(q) == "error"
	// log.Printf("last type: %+v", last.param.typ.String())
}

func (sg paramStructGenerator) encoders() []encoder {
	encs := []encoder{}
	for _, p := range sg.params {
		encs = append(encs, p.typ)
	}

	return encs
}
