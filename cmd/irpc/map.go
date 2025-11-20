package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
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

// decode implements Type.
func (m mapType) decode(varId string, existingVars varNames, q *qualifier) string {
	sb := &strings.Builder{}

	// length
	fmt.Fprintf(sb, "{ // %s %s\n", varId, m.name(q))
	sb.WriteString("var l int\n")
	sb.WriteString(m.lenT.decode("l", existingVars, q))
	existingVars = append(existingVars, "l")

	fmt.Fprintf(sb, "%s = make(%s, l)\n", varId, m.name(q))
	sb.WriteString("for range l {\n")

	fmt.Fprintf(sb, "var k %s\n", m.keyT.name(q))
	existingVars = append(existingVars, "k")
	fmt.Fprintf(sb, "%s\n", m.keyT.decode("k", existingVars, q))

	fmt.Fprintf(sb, "var v %s\n", m.valT.name(q))
	existingVars = append(existingVars, "v")
	fmt.Fprintf(sb, "%s\n", m.valT.decode("v", existingVars, q))

	fmt.Fprintf(sb, "%s[k] = v", varId)
	sb.WriteString("}\n")
	sb.WriteString("}\n") // end of block
	return sb.String()
}

// encode implements Type.
func (m mapType) encode(varId string, existingVars varNames, q *qualifier) string {
	sb := &strings.Builder{}
	// length
	fmt.Fprintf(sb, "{ // %s %s\n", varId, m.name(q))
	sb.WriteString(m.lenT.encode("len("+varId+")", existingVars, q))

	keyIt, valIt := existingVars.generateKeyValueIteratorNames()

	// for loop
	fmt.Fprintf(sb, "for %s, %s := range %s {", keyIt, valIt, varId)
	sb.WriteString(m.keyT.encode(keyIt, existingVars, q))
	sb.WriteString(m.valT.encode(valIt, existingVars, q))
	sb.WriteString("}\n") // end of for loop

	sb.WriteString("}\n") // end of block
	return sb.String()
}
