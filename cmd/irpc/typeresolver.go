package main

import (
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type typeResolver struct {
	g        *generator // todo: get rid of
	inputPkg *packages.Package
	allPkgs  []*packages.Package
}

// todo: make value type?
func newTypeResolver(g *generator, inputPkg *packages.Package, allPkgs []*packages.Package) (*typeResolver, error) {
	return &typeResolver{
		g:        g,
		inputPkg: inputPkg,
		allPkgs:  allPkgs,
	}, nil
}

func (tr *typeResolver) findTypeAndValueForAst(expr ast.Expr) (types.TypeAndValue, error) {
	for _, pkg := range tr.allPkgs {
		if tv, ok := pkg.TypesInfo.Types[expr]; ok {
			return tv, nil
		}
	}
	return types.TypeAndValue{}, fmt.Errorf("typeAndValue for expr %T:%v not found", expr, expr)
}

func (tr *typeResolver) loadRpcParamList(apiName string, list []*ast.Field) ([]rpcParam, error) {
	params := []rpcParam{}
	for pos, field := range list {
		astExpr := field.Type
		tv, err := tr.findTypeAndValueForAst(astExpr)
		if err != nil {
			return nil, fmt.Errorf("couldn't determine ast's %T fileld's type and value: %w", astExpr, err)
		}
		t, err := tr.newType(apiName, tv.Type, astExpr)
		if err != nil {
			return nil, fmt.Errorf("newType(): %w", err)
		}

		if field.Names == nil {
			// parameter doesn't have name, just a type (typically function returns)
			param, err := newRpcParam(pos, "", t)
			if err != nil {
				return nil, fmt.Errorf("newRpcParam on pos %d: %w", pos, err)
			}
			// tr.addImport(param.typ.ImportSpecs()) // todo: remove?
			params = append(params, param)
		} else {
			for _, name := range field.Names {
				param, err := newRpcParam(pos, name.Name, t)
				if err != nil {
					return nil, fmt.Errorf("newRpcParam on pos %d: %w", pos, err)
				}
				// tr.addImport(param.typ.ImportSpecs()) // todo: remove?
				params = append(params, param)
			}
		}
	}
	return params, nil
}

func (tr *typeResolver) newType(apiName string, t types.Type, astExpr ast.Expr) (Type, error) {
	ni, utAst, err := tr.unwrapNamedOrPassThrough(t, astExpr)
	if err != nil {
		return nil, fmt.Errorf("unwrapNamedOrPassThrough(): %w", err)
	}

	switch ut := t.Underlying().(type) {
	case *types.Basic:
		return tr.newBasicType(ut, ni)
	case *types.Slice:
		return tr.newSliceType(apiName, ni, ut, utAst)
	case *types.Map: // todo: test maps using http.Header named map (doesn't have ast etc..)
		return tr.newMapType(apiName, ni, ut, utAst)
	case *types.Struct:
		return tr.newStructType(apiName, ni, ut, utAst)
	case *types.Interface:
		return tr.newInterfaceType(apiName, ni, ut, utAst)
	default:
		return nil, fmt.Errorf("unsupported type: %T", ut)
	}
}

func (tr *typeResolver) unwrapNamedOrPassThrough(t types.Type, astExpr ast.Expr) (*namedInfo, ast.Expr, error) {
	var is importSpec
	// var namedInfo namedInfo
	named, isNamed := t.(*types.Named)
	if isNamed {
		obj := named.Obj()
		namedName := obj.Name()
		// log.Printf("namedName: %q", namedName)
		pkg := obj.Pkg()
		// log.Printf("pkg: %v", pkg)
		if pkg != nil {
			is.path = pkg.Path()
			is.pkgName = pkg.Name()
		}
		// we try to find the "pkgName"  int pkgName.VarType
		// it can differ from pkg.Name() in case of alias and we would like to honour it
		if astExpr != nil {
			packagePrefix := packagePrefixFromAst(astExpr)
			// log.Printf("setting qualifier: %q", packagePrefix)
			is.alias = packagePrefix
		}

		return &namedInfo{
			namedName:  namedName,
			importSpec: is,
		}, nil, nil
	}

	// pass through
	return nil, astExpr, nil
}

// if type is named, it tries to find it's ast in the package. returns nil if not found
// if type is not named, it assumes passed in astExpr is valid astExpr of type and returns it
// todo: maybe we can squeee it inside typeNameAndImport() ?
func (tr *typeResolver) unwrapTypeAst(t types.Type, astExpr ast.Expr) ast.Expr {
	named, isNamed := t.(*types.Named)
	if isNamed {
		typeSpec, err := tr.g.findAstTypeSpec(named)
		if err != nil {
			return nil
		}
		return typeSpec.Type
	}
	return astExpr
}

// returns ex: "time" if ast.Expr is at "time.Time". otherwise ""
func packagePrefixFromAst(astExpr ast.Expr) string {
	if selExpr, ok := astExpr.(*ast.SelectorExpr); ok {
		if ident, ok := selExpr.X.(*ast.Ident); ok {
			return ident.Name
		}
	}
	return ""
}
