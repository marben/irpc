package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

type Type interface {
	encoder
	Name(q *qualifier) string
}

type namedInfo struct {
	namedName  string
	importSpec importSpec
}

func (ni namedInfo) qualifiedName(q *qualifier) string {
	return q.qualifyNamedInfo(ni)
}

type basicType struct {
	enc            directCallEncoder
	underlyingName string
	namedInfo      *namedInfo
}

func (bt basicType) Name(q *qualifier) string {
	if bt.namedInfo == nil {
		return bt.underlyingName
	}
	return q.qualifyNamedInfo(*bt.namedInfo)
}

var _ Type = basicType{}

// ast can be nil if not available
func (tr *typeResolver) newBasicType(bt *types.Basic, ni *namedInfo) (Type, error) {
	var irpcFuncName string
	switch bt.Kind() {
	case types.Bool:
		irpcFuncName = "Bool"
	case types.Int:
		irpcFuncName = "VarInt"
	case types.Uint:
		irpcFuncName = "UvarInt"
	case types.Int8:
		irpcFuncName = "Int8"
	case types.Uint8: // serves 'types.Byte' as well
		irpcFuncName = "Uint8"
	case types.Int16:
		irpcFuncName = "VarInt16"
	case types.Uint16:
		irpcFuncName = "UvarInt16"
	case types.Int32: // serves 'types.Rune' as well
		irpcFuncName = "VarInt32"
	case types.Uint32:
		irpcFuncName = "UvarInt32"
	case types.Int64:
		irpcFuncName = "VarInt64"
	case types.Uint64:
		irpcFuncName = "UvarInt64"
	case types.Float32:
		irpcFuncName = "Float32le"
	case types.Float64:
		irpcFuncName = "Float64le"
	case types.String:
		irpcFuncName = "String"
	default:
		return basicType{}, fmt.Errorf("unsupported basic type %q", bt.Name())
	}

	var needsCasting bool
	if ni != nil {
		needsCasting = true
	}

	enc := directCallEncoder{
		encFuncName:        irpcFuncName,
		decFuncName:        irpcFuncName,
		underlyingTypeName: bt.Name(),
		needsCasting:       needsCasting,
	}

	return basicType{
		enc:            enc,
		underlyingName: bt.Name(),
		namedInfo:      ni,
	}, nil
}

func (bt basicType) codeblock(q *qualifier) string {
	return bt.enc.codeblock(q)
}

func (bt basicType) decode(varId string, existingVars varNameList, q *qualifier) string {
	return bt.enc.decode(varId, existingVars, q)
}

func (bt basicType) encode(varId string, existingVars varNameList, q *qualifier) string {
	return bt.enc.encode(varId, existingVars, q)
}

func (bt basicType) Encoder() encoder {
	return bt.enc
}

type sliceType struct {
	elem   Type
	lenEnc encoder
	ni     *namedInfo
}

func (g *generator) findAstTypeSpec(named *types.Named) (*ast.TypeSpec, error) {
	obj := named.Obj()
	if obj.Pkg() == nil {
		return nil, fmt.Errorf("no obj.Pkg() for named type %q", named)
	}
	pkg, err := g.findPackageForPackagePath(obj.Pkg().Path())
	if err != nil {
		return nil, fmt.Errorf("no pkg for file path: %q", obj.Pkg().Path())
	}
	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}

			// log.Printf("current genDecl: %v", gen.Specs)
			for _, spec := range gen.Specs {
				ts := spec.(*ast.TypeSpec)
				// log.Printf("looking at typespec: %v", ts.Name)
				if pkg.TypesInfo.Defs[ts.Name] == obj {
					return ts, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("findTypeSpec(): found no spec for obj: %#v", obj)
}

// todo: use field for signatures too?
type field struct {
	name string
	t    Type
}
