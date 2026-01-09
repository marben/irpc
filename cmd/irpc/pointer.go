package main

import (
	"fmt"
	"go/ast"
	"go/types"
)

var _ Type = pointerType{}

type pointerType struct {
	elemT Type
	ni    *namedInfo
}

func (tr *typeResolver) newPointerType(apiName string, ni *namedInfo, pt *types.Pointer, astExpr ast.Expr) (pointerType, error) {
	var elemAst ast.Expr
	if astExpr != nil {
		starExpr, ok := astExpr.(*ast.StarExpr)
		if !ok {
			return pointerType{}, fmt.Errorf("pointer's astExpr is not *ast.StarExpr, but %T of value %#[1]v", astExpr)
		}
		elemAst = starExpr.X
	}

	elemT, err := tr.newType(apiName, pt.Elem(), elemAst)
	if err != nil {
		return pointerType{}, fmt.Errorf("newType() for pointer element: %q: %w", pt.Elem(), err)
	}

	return pointerType{
		elemT: elemT,
		ni:    ni,
	}, nil
}

// codeblocks implements Type.
func (pt pointerType) codeblocks(q *qualifier) []string {
	return pt.elemT.codeblocks(q)
}

// genEncFunc implements Type.
func (pt pointerType) genEncFunc(q *qualifier) string {
	lq := q.copy()
	return fmt.Sprintf(`func(enc *irpcgen.Encoder, pt %s) error{
		return irpcgen.EncPointer(enc, pt, %q, %s)
	}`, pt.name(q), pt.elemT.name(lq), pt.elemT.genEncFunc(q))
}

// genDecFunc implements Type.
func (pt pointerType) genDecFunc(q *qualifier) string {
	lq := q.copy()
	return fmt.Sprintf(`func(dec *irpcgen.Decoder, pt *%s) error {
		return irpcgen.DecPointer(dec, pt, %q, %s)
	}`, pt.name(q), pt.elemT.name(lq), pt.elemT.genDecFunc(q))
}

// name implements Type.
func (pt pointerType) name(q *qualifier) string {
	if pt.ni != nil {
		return q.qualifyNamedInfo(*pt.ni)
	}
	return "*" + pt.elemT.name(q)
}
