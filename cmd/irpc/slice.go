package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
)

var _ Type = sliceType{}

func (tr *typeResolver) newSliceType(apiName string, ni *namedInfo, st *types.Slice, astExpr ast.Expr) (sliceType, error) {
	var elemAst ast.Expr
	if astExpr != nil {
		arrayAst, ok := astExpr.(*ast.ArrayType)
		if !ok {
			return sliceType{}, fmt.Errorf("slice's astExpression is not *ast.ArrayType, but %T of value %#[1]v", astExpr)
		}
		elemAst = arrayAst.Elt
	}

	elemT, err := tr.newType(apiName, st.Elem(), elemAst)
	if err != nil {
		return sliceType{}, fmt.Errorf("newType() for slices element %q: %w", st.Elem(), err)
	}

	return sliceType{
		elem:   elemT,
		lenEnc: uint64Encoder,
		ni:     ni,
	}, nil
}

func (st sliceType) Name(q *qualifier) string {
	if st.ni != nil {
		return q.qualifyNamedInfo(*st.ni)
	}
	return "[]" + st.elem.Name(q)
}

// encode implements encoder.
func (st sliceType) encode(varId string, existingVars varNameList, q *qualifier) string {
	sb := &strings.Builder{}

	// length
	fmt.Fprintf(sb, "{ // %s %s\n", varId, st.Name(q))
	fmt.Fprintf(sb, "var l int = len(%s)\n", varId)
	sb.WriteString(st.lenEnc.encode("uint64(l)", existingVars, q))
	existingVars = append(existingVars, "l")

	// for loop
	existingVars = append(existingVars, "v")
	fmt.Fprintf(sb, "for _, v := range %s {\n", varId)
	sb.WriteString(st.elem.encode("v", existingVars, q))
	sb.WriteString("}")
	sb.WriteString("}\n")

	return sb.String()
}

// decode implements encoder.
func (st sliceType) decode(varId string, existingVars varNameList, q *qualifier) string {
	sb := &strings.Builder{}

	// length
	fmt.Fprintf(sb, "{ // %s %s\n", varId, st.Name(q))
	sb.WriteString("var ul uint64\n")
	sb.WriteString(st.lenEnc.decode("ul", existingVars, q))
	sb.WriteString("var l int = int(ul)\n")
	existingVars = append(existingVars, "l", "ul")

	// for loop
	itName := generateIteratorName(existingVars)
	existingVars = append(existingVars, itName)
	fmt.Fprintf(sb, "%s = make(%s, l)\n", varId, st.Name(q))
	fmt.Fprintf(sb, "for %s := range l {", itName)
	sb.WriteString(st.elem.decode(varId+"["+itName+"]", existingVars, q))
	sb.WriteString("}\n")
	sb.WriteString("}\n")

	return sb.String()
}

// codeblock implements encoder.
func (st sliceType) codeblock(q *qualifier) string {
	return ""
}
