package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
)

func (tr *typeResolver) newSliceType(apiName string, ni *namedInfo, st *types.Slice, astExpr ast.Expr) (Type, error) {
	switch st.Elem().Underlying().String() {
	case "byte":
		return tr.newByteSliceType(ni)
	case "bool":
		return tr.newBoolSliceType(ni)
	}
	if st.String() == "[]uint8" {
		// []uint8 is interchangeable with []byte. but not if named
		return newDirectCallType("ByteSlice", "ByteSlice", "[]uint8", ni), nil
	}

	return tr.newGenericSliceType(apiName, ni, st, astExpr)
}

// []byte
func (tr *typeResolver) newByteSliceType(ni *namedInfo) (Type, error) {
	return newDirectCallType("ByteSlice", "ByteSlice", "[]byte", ni), nil
}

// []bool
func (tr *typeResolver) newBoolSliceType(ni *namedInfo) (Type, error) {
	return newDirectCallType("BoolSlice", "BoolSlice", "[]bool", ni), nil
}

type genericSliceType struct {
	elemT Type
	lenT  Type
	ni    *namedInfo
}

func (tr *typeResolver) newGenericSliceType(apiName string, ni *namedInfo, st *types.Slice, astExpr ast.Expr) (genericSliceType, error) {
	var elemAst ast.Expr
	if astExpr != nil {
		arrayAst, ok := astExpr.(*ast.ArrayType)
		if !ok {
			return genericSliceType{}, fmt.Errorf("slice's astExpression is not *ast.ArrayType, but %T of value %#[1]v", astExpr)
		}
		elemAst = arrayAst.Elt
	}

	elemT, err := tr.newType(apiName, st.Elem(), elemAst)
	if err != nil {
		return genericSliceType{}, fmt.Errorf("newType() for slices element %q: %w", st.Elem(), err)
	}

	return genericSliceType{
		elemT: elemT,
		lenT:  tr.lenType,
		ni:    ni,
	}, nil
}

func (st genericSliceType) name(q *qualifier) string {
	if st.ni != nil {
		return q.qualifyNamedInfo(*st.ni)
	}
	return "[]" + st.elemT.name(q)
}

// encode implements encoder.
func (st genericSliceType) encode(varId string, existingVars varNames, q *qualifier) string {
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
func (st genericSliceType) decode(varId string, existingVars varNames, q *qualifier) string {
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
func (st genericSliceType) codeblocks(q *qualifier) []string {
	return st.elemT.codeblocks(q)
}
