package main

import (
	"fmt"
	"go/ast"
	"go/types"
)

var _ Type = chanType{}

// chanType is a channel used directly as a method parameter or result (a stream).
// It is never serialized itself: the generator turns it into stream calls moving []elem batches.
type chanType struct {
	dir     types.ChanDir
	elemT   Type       // element's Type, for the generated code
	elem    types.Type // element's go type, used to synthesize the []elem batch type
	elemAst ast.Expr   // element's ast (may be nil)
}

// newChanType is called only for channels used directly as method parameters/results.
// Channels anywhere else (struct fields, slices, ...) are rejected by newType.
func (tr *typeResolver) newChanType(apiName string, ct *types.Chan, astExpr ast.Expr) (chanType, error) {
	var elemAst ast.Expr
	if chanAst, ok := astExpr.(*ast.ChanType); ok {
		elemAst = chanAst.Value
	}

	if _, isChan := ct.Elem().Underlying().(*types.Chan); isChan {
		return chanType{}, fmt.Errorf("channel of channels (%s) is not supported", ct)
	}
	if tr.typeIsContext(ct.Elem()) {
		return chanType{}, fmt.Errorf("channel of context.Context (%s) is not supported", ct)
	}

	elemT, err := tr.newType(apiName, ct.Elem(), elemAst)
	if err != nil {
		return chanType{}, fmt.Errorf("channel element type %s: %w", ct.Elem(), err)
	}

	return chanType{
		dir:     ct.Dir(),
		elemT:   elemT,
		elem:    ct.Elem(),
		elemAst: elemAst,
	}, nil
}

// batchType returns the []elem Type used to move the values of the stream
func (tr *typeResolver) batchType(apiName string, ct chanType) (Type, error) {
	var sliceAst ast.Expr
	if ct.elemAst != nil {
		// keeps the element's import alias (e.g. []myTime.Time)
		sliceAst = &ast.ArrayType{Elt: ct.elemAst}
	}
	return tr.newType(apiName, types.NewSlice(ct.elem), sliceAst)
}

func (c chanType) name(q *qualifier) string {
	switch c.dir {
	case types.RecvOnly:
		return "<-chan " + c.elemT.name(q)
	case types.SendOnly:
		return "chan<- " + c.elemT.name(q)
	default:
		return "chan " + c.elemT.name(q)
	}
}

// genEncFunc implements Type. Channels are never serialized directly.
func (c chanType) genEncFunc(q *qualifier) string { return "" }

// genDecFunc implements Type. Channels are never serialized directly.
func (c chanType) genDecFunc(q *qualifier) string { return "" }

// codeblocks implements Type. The element's codeblocks come with the batch type.
func (c chanType) codeblocks(q *qualifier) []string { return nil }

func isChanType(t Type) bool {
	_, ok := t.(chanType)
	return ok
}
