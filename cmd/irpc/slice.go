package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
)

var _ Type = sliceType{}

// []byte
func (tr *typeResolver) newByteSliceType(ni *namedInfo) (Type, error) {
	return newDirectCallType("ByteSlice", "ByteSlice", "[]byte", ni), nil
}

// []bool
func (tr *typeResolver) newBoolSliceType(ni *namedInfo) (Type, error) {
	return newDirectCallType("BoolSlice", "BoolSlice", "[]bool", ni), nil
}

// generic slice encoder
type sliceType struct {
	elemT Type
	lenT  Type
	ni    *namedInfo
}

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
		elemT: elemT,
		lenT:  tr.lenType,
		ni:    ni,
	}, nil
}

func (st sliceType) name(q *qualifier) string {
	if st.ni != nil {
		return q.qualifyNamedInfo(*st.ni)
	}
	return "[]" + st.elemT.name(q)
}

// encode implements encoder.
func (st sliceType) encode(varId string, existingVars varNames, q *qualifier) string {
	sb := &strings.Builder{}

	// length
	fmt.Fprintf(sb, "{ // %s %s\n", varId, st.name(q))
	sb.WriteString(st.lenT.encode("len("+varId+")", existingVars, q))

	// for loop
	existingVars = append(existingVars, "v")
	fmt.Fprintf(sb, "for _, v := range %s {\n", varId)
	sb.WriteString(st.elemT.encode("v", existingVars, q))
	sb.WriteString("}")
	sb.WriteString("}\n")

	return sb.String()
}

// decode implements encoder.
func (st sliceType) decode(varId string, existingVars varNames, q *qualifier) string {
	sb := &strings.Builder{}

	// length
	fmt.Fprintf(sb, "{ // %s %s\n", varId, st.name(q))
	sb.WriteString("var l int\n")
	sb.WriteString(st.lenT.decode("l", existingVars, q))
	existingVars = append(existingVars, "l")

	// for loop
	itName := existingVars.generateIteratorName()
	fmt.Fprintf(sb, "%s = make(%s, l)\n", varId, st.name(q))
	fmt.Fprintf(sb, "for %s := range l {", itName)
	sb.WriteString(st.elemT.decode(varId+"["+itName+"]", existingVars, q))
	sb.WriteString("}\n")
	sb.WriteString("}\n")

	return sb.String()
}

// codeblock implements encoder.
func (st sliceType) codeblock(q *qualifier) string {
	return ""
}
