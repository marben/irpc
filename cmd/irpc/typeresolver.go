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
	ut := t.Underlying()

	switch ut := ut.(type) {
	case *types.Basic:
		return tr.newBasicTypeT(t, ut, astExpr)
	case *types.Slice:
		return tr.newSliceTypeT(apiName, t, ut, astExpr)
	case *types.Map: // todo: test maps using http.Header named map (doesn't have ast etc..)
		return tr.newMapType(apiName, astExpr, t)
	case *types.Struct:
		return tr.newStructTypeT(apiName, t, ut, astExpr)
	case *types.Interface:
		return tr.newInterfaceTypeT(apiName, t, ut, astExpr)
	default:
		return nil, fmt.Errorf("unsupported Type %T", ut)
	}
}

// if type is named, resolves it's name and import specs
// extracts package prefix from astExpr (if it's ast.SelectorExpr)
// astExpr can be nil in which case the alias in import spec will be empty
func (tr *typeResolver) typeNameAndImport(t types.Type, astExpr ast.Expr) (name string, is importSpec) {
	// var is importSpec
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

		return namedName, is
	}
	// todo: inline interfaces are not yet supported. it seems like we will need
	// our own implementation of TypeString(). or at least a good qualifier
	// we have inline type. can be interface, struct slice
	// needs to be sanitized
	// unsanitized := t.String()
	// log.Printf("unsanitized: %q", unsanitized)
	// sanitized := "this_is_sanitized_interface_name"
	// return sanitized, importSpec{}

	return types.TypeString(t, nil), importSpec{}
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
