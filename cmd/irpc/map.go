package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
)

var _ Type = mapType{}

type mapType struct {
	lenEnc   encoder
	key, val Type
	ni       *namedInfo
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
	// log.Printf("keyT: %q", keyT)
	valT, err := tr.newType(apiName, t.Elem(), valAst)
	if err != nil {
		return mapType{}, fmt.Errorf("newType() for map value %q: %w", valAst, err)
	}
	// log.Printf("valT: %q", valT)

	return mapType{
		lenEnc: uint64Encoder,
		key:    keyT,
		val:    valT,
		ni:     ni,
	}, nil
}

// Name implements Type.
func (m mapType) Name(q *qualifier) string {
	if m.ni != nil {
		return m.ni.qualifiedName(q)
	}
	return "map[" + m.key.Name(q) + "]" + m.val.Name(q)
}

// codeblock implements Type.
func (m mapType) codeblock(q *qualifier) string {
	return ""
}

// decode implements Type.
func (m mapType) decode(varId string, existingVars varNames, q *qualifier) string {
	sb := &strings.Builder{}

	// length
	fmt.Fprintf(sb, "{ // %s %s\n", varId, m.Name(q))
	sb.WriteString("var ul uint64\n")
	sb.WriteString(m.lenEnc.decode("ul", existingVars, q))
	sb.WriteString("var l int = int(ul)\n")
	existingVars = append(existingVars, "ul", "l")

	// fmt.Fprintf(sb, "%s = make(map[%s]%s, l)\n", varId, m.key.Name(), m.val.Name())
	fmt.Fprintf(sb, "%s = make(%s, l)\n", varId, m.Name(q))
	sb.WriteString("for range l {\n")

	fmt.Fprintf(sb, "var k %s\n", m.key.Name(q))
	existingVars = append(existingVars, "k")
	fmt.Fprintf(sb, "%s\n", m.key.decode("k", existingVars, q))

	fmt.Fprintf(sb, "var v %s\n", m.val.Name(q))
	existingVars = append(existingVars, "v")
	fmt.Fprintf(sb, "%s\n", m.val.decode("v", existingVars, q))

	fmt.Fprintf(sb, "%s[k] = v", varId)
	sb.WriteString("}\n")
	sb.WriteString("}\n") // end of block
	return sb.String()
}

// encode implements Type.
func (m mapType) encode(varId string, existingVars varNames, q *qualifier) string {
	sb := &strings.Builder{}
	// length
	fmt.Fprintf(sb, "{ // %s %s\n", varId, m.Name(q))
	fmt.Fprintf(sb, "var l int = len(%s)\n", varId)
	sb.WriteString(m.lenEnc.encode("uint64(l)", existingVars, q))
	existingVars = append(existingVars, "l")

	keyIt, valIt := existingVars.generateKeyValueIteratorNames()

	// for loop
	fmt.Fprintf(sb, "for %s, %s := range %s {", keyIt, valIt, varId)
	sb.WriteString(m.key.encode(keyIt, existingVars, q))
	sb.WriteString(m.val.encode(valIt, existingVars, q))
	sb.WriteString("}\n") // end of for loop

	sb.WriteString("}\n") // end of block
	return sb.String()
}
