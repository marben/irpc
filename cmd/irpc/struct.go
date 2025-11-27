package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
)

var _ Type = structType{}

// returns nil, if ast is nil
// if i is out of bounds, panics
func structAstFieldOrNil(ast *ast.StructType, i int) ast.Expr {
	if ast == nil {
		return nil
	}

	return astTypeFieldFromFieldList(ast.Fields, i)
}

// panics if ast is nil or index is out of bounds
// return's Type part of the field
func astTypeFieldFromFieldList(astFieldList *ast.FieldList, i int) ast.Expr {
	var curr int
	for _, f := range astFieldList.List {
		typeExpr := f.Type
		if f.Names == nil {
			if curr == i {
				return typeExpr
			}
			curr++
		} else {
			for range f.Names {
				if curr == i {
					return typeExpr
				}
				curr++
			}
		}
	}

	panic(fmt.Errorf("there is no ast field indexed %d in %#v", i, astFieldList))
}

type structType struct {
	fields []field
	ni     *namedInfo
}

func (tr *typeResolver) newStructType(apiName string, ni *namedInfo, t *types.Struct, astExpr ast.Expr) (structType, error) {
	var structAst *ast.StructType
	if astExpr != nil {
		var ok bool
		structAst, ok = astExpr.(*ast.StructType)
		if !ok {
			return structType{}, fmt.Errorf("provided ast is not *ast.StructType, but %T", astExpr)
		}
	}
	fields := []field{}
	for i := 0; i < t.NumFields(); i++ {
		f := t.Field(i)
		// log.Printf("%d: field: %q", i, f.Name())
		fieldAst := structAstFieldOrNil(structAst, i)
		ft, err := tr.newType(apiName, f.Type(), fieldAst)
		if err != nil {
			return structType{}, fmt.Errorf("create Type for field %q: %w", f, err)
		}
		sf := newField(f.Name(), ft)
		fields = append(fields, sf)
	}

	return structType{
		ni:     ni,
		fields: fields,
	}, nil
}

// name implements Type.
func (s structType) name(q *qualifier) string {
	if s.ni != nil {
		return s.ni.qualifiedName(q)
	}

	sb := strings.Builder{}
	sb.WriteString("struct{")
	for _, f := range s.fields {
		sb.WriteString(f.name + " " + f.t.name(q))
		sb.WriteString(";")
	}
	sb.WriteString("}")
	return sb.String()
}

// codeblock implements Type.
func (s structType) codeblocks(q *qualifier) []string {
	var cb []string
	for _, f := range s.fields {
		cb = append(cb, f.t.codeblocks(q)...)
	}
	return cb
}

// genEncFunc implements Type.
func (s structType) genEncFunc(_ string, q *qualifier) string {
	lq := q.copy()
	sb := &strings.Builder{}
	// todo: 's' could be named like the serialize/deserialize struct param. we would need to revam the encFunc definition
	fmt.Fprintf(sb, "func(enc *irpcgen.Encoder, s %s) error {\n", s.name(q))
	for _, f := range s.fields {
		fmt.Fprintf(sb, `if err := %s(enc, s.%s); err != nil {
			return fmt.Errorf("serialize s.%s of type %s: %%w", err)
		}
		 `, f.t.genEncFunc("enc", q), f.name, f.name, f.t.name(lq))
	}
	sb.WriteString("return nil\n")
	sb.WriteString("}")
	return sb.String()
}

// genDecFunc implements Type.
func (s structType) genDecFunc(decoderVarName string, q *qualifier) string {
	lq := q.copy()
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "func(dec *irpcgen.Decoder, s *%s) error {\n", s.name(q))
	for _, f := range s.fields {
		fmt.Fprintf(sb, `if err := %s(dec, &s.%s); err != nil {
			return fmt.Errorf("deserialize s.%s of type %s: %%w", err)
		}
		`, f.t.genDecFunc("dec", q), f.name, f.name, f.t.name(lq))
	}
	sb.WriteString("return nil\n")
	sb.WriteString("}")

	return sb.String()
}
