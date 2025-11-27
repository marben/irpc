package main

import (
	"fmt"
	"go/ast"
	"go/types"
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
		return newDirectCallType("irpcgen.EncByteSlice", "irpcgen.DecByteSlice", "[]uint8", ni), nil
	}

	return tr.newGenericSliceType(apiName, ni, st, astExpr)
}

// []byte
func (tr *typeResolver) newByteSliceType(ni *namedInfo) (Type, error) {
	return newDirectCallType("irpcgen.EncByteSlice", "irpcgen.DecByteSlice", "[]byte", ni), nil
}

// []bool
func (tr *typeResolver) newBoolSliceType(ni *namedInfo) (Type, error) {
	return newDirectCallType("irpcgen.EncBoolSlice", "irpcgen.DecBoolSlice", "[]bool", ni), nil
}

var _ Type = genericSliceType{}

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

// genEncFunc implements Type.
func (st genericSliceType) genEncFunc(q *qualifier) string {
	lq := q.copy()
	return fmt.Sprintf(`func(enc *irpcgen.Encoder, sl %s) error{
		return irpcgen.EncSlice(enc, %q, %s, sl)
	}`, st.name(q), st.elemT.name(lq), st.elemT.genEncFunc(q))
}

// genDecFunc implements Type.
func (st genericSliceType) genDecFunc(q *qualifier) string {
	lq := q.copy()
	return fmt.Sprintf(`func(dec *irpcgen.Decoder, sl *%s) error {
		return irpcgen.DecSlice(dec, %q, %s, sl)
	}`, st.name(q), st.elemT.name(lq), st.elemT.genDecFunc(q))
}

// codeblock implements encoder.
func (st genericSliceType) codeblocks(q *qualifier) []string {
	return st.elemT.codeblocks(q)
}
