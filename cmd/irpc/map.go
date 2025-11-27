package main

import (
	"fmt"
	"go/ast"
	"go/types"
)

var _ Type = mapType{}

type mapType struct {
	lenT       Type
	keyT, valT Type
	ni         *namedInfo
}

func (tr *typeResolver) newMapType(apiName string, ni *namedInfo, t *types.Map, astExpr ast.Expr) (mapType, error) {
	var keyAst, valAst ast.Expr
	if astExpr != nil {
		mapAst, ok := astExpr.(*ast.MapType)
		if !ok {
			return mapType{}, fmt.Errorf("provided ast.Expr is not *ast.MapType but %T", astExpr)
		}
		keyAst = mapAst.Key
		valAst = mapAst.Value
	}

	keyT, err := tr.newType(apiName, t.Key(), keyAst)
	if err != nil {
		return mapType{}, fmt.Errorf("newType() for map key %q: %w", keyAst, err)
	}
	valT, err := tr.newType(apiName, t.Elem(), valAst)
	if err != nil {
		return mapType{}, fmt.Errorf("newType() for map value %q: %w", valAst, err)
	}

	return mapType{
		lenT: tr.lenType,
		keyT: keyT,
		valT: valT,
		ni:   ni,
	}, nil
}

// name implements Type.
func (m mapType) name(q *qualifier) string {
	if m.ni != nil {
		return m.ni.qualifiedName(q)
	}
	return "map[" + m.keyT.name(q) + "]" + m.valT.name(q)
}

// codeblock implements Type.
func (m mapType) codeblocks(q *qualifier) []string {
	return append(m.keyT.codeblocks(q), m.valT.codeblocks(q)...)
}

// genEncFunc implements Type.
func (m mapType) genEncFunc(encoderVarName string, q *qualifier) string {
	lq := q.copy()
	return fmt.Sprintf(`func(enc *irpcgen.Encoder, m %s)error{
		return irpcgen.EncMap(enc, m, %q, %s, %q, %s)
	}`, m.name(q), m.keyT.name(lq), m.keyT.genEncFunc("enc", q), m.valT.name(lq), m.valT.genEncFunc("enc", q))
}

// genDecFunc implements Type.
func (m mapType) genDecFunc(decoderVarName string, q *qualifier) string {
	lq := q.copy()
	return fmt.Sprintf(`func(dec *irpcgen.Decoder, m *%s) error {
		return irpcgen.DecMap(dec, m, %q, %s, %q, %s)
	}`, m.name(q), m.keyT.name(lq), m.keyT.genDecFunc("dec", q), m.valT.name(lq), m.valT.genDecFunc("dec", q))
}
