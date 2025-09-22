package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/types"
	"path/filepath"
	"strconv"

	"golang.org/x/tools/go/packages"
)

type typeResolver struct {
	inputPkg                     *packages.Package
	allPkgs                      []*packages.Package
	srcFileAst                   *ast.File
	srcImports                   orderedSet[importSpec] // imports from the src file
	binMarshaler, binUnmarshaler *types.Interface
}

// todo: make value type?
func newTypeResolver(filename string /*, inputPkg *packages.Package, allPkgs []*packages.Package*/) (typeResolver, error) {
	absFilePath, err := filepath.Abs(filename)
	if err != nil {
		return typeResolver{}, fmt.Errorf("filepath.Abs(): %w", err)
	}

	dir := filepath.Dir(absFilePath)

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedDeps | packages.NeedImports | packages.NeedSyntax | packages.NeedTypesInfo |
			packages.NeedFiles | packages.NeedName | packages.NeedCompiledGoFiles | packages.NeedExportFile | packages.NeedSyntax |
			packages.NeedModule,
		Dir: dir,
	}

	allPackages, err := packages.Load(cfg, "./...") // todo: start at module root
	if err != nil {
		return typeResolver{}, fmt.Errorf("packages.Load(): %w", err)
	}

	targetPkg, err := findPackageForFile(allPackages, filename)
	if err != nil {
		return typeResolver{}, err
	}

	fileAst, err := findASTForFile(targetPkg, filename)
	if err != nil {
		return typeResolver{}, fmt.Errorf("couldn't find ast for given file %s", filename)
	}

	srcImports := newOrderedSet[importSpec]()
	for _, is := range fileAst.Imports {
		// log.Printf("file import spec: %q  %q", is.Name, is.Path.Value)
		var alias string
		if is.Name != nil {
			alias = is.Name.Name
		}
		path, err := strconv.Unquote(is.Path.Value)
		if err != nil {
			return typeResolver{}, fmt.Errorf("strconv.Unquote(%q): %w", is.Path.Value, err)
		}

		if pkg, ok := findPackageForPackagePath(allPackages, path); ok {
			// package was parsed by go/packages library
			spec := importSpec{alias, path, pkg.Name}
			srcImports.add(spec)
			// log.Printf("added src import: %+v", spec)
		} else {
			// must be from stdlib, which isn't provided by the packages lib
			imp := importer.Default()
			pkg, err := imp.Import(path)
			if err != nil {
				return typeResolver{}, fmt.Errorf("importer.Import(%q): %w", path, err)
			}
			spec := importSpec{alias: alias, path: pkg.Path(), pkgName: pkg.Name()}
			srcImports.add(spec)
			// log.Printf("added src import: %+v", spec)
		}
	}

	imp := importer.Default()
	encodingPkg, err := imp.Import("encoding")
	if err != nil {
		return typeResolver{}, fmt.Errorf("importer.Import(\"encoding\"): %w", err)
	}
	binMarshaler, ok := encodingPkg.Scope().Lookup("BinaryMarshaler").Type().Underlying().(*types.Interface)
	if !ok {
		return typeResolver{}, fmt.Errorf("failed to find encoding.BinaryMarshaller type")
	}
	binUnmarshaler, ok := encodingPkg.Scope().Lookup("BinaryUnmarshaler").Type().Underlying().(*types.Interface)
	if !ok {
		return typeResolver{}, fmt.Errorf("failed to find encoding.BinaryUnmarshaller type")
	}

	// contextPkg, err := imp.Import("context")
	// if err != nil {
	// 	return typeResolver{}, fmt.Errorf("importer.Import(\"context\"): %w", err)
	// }
	// contextObj := contextPkg.Scope().Lookup("Context")
	// if contextObj == nil {
	// 	return typeResolver{}, fmt.Errorf("failed to find context.Context object")
	// }

	return typeResolver{
		inputPkg:       targetPkg,
		allPkgs:        allPackages,
		srcFileAst:     fileAst,
		srcImports:     srcImports,
		binMarshaler:   binMarshaler,
		binUnmarshaler: binUnmarshaler,
	}, nil
}

func (tr typeResolver) findTypeAndValueForAst(expr ast.Expr) (types.TypeAndValue, error) {
	for _, pkg := range tr.allPkgs {
		if tv, ok := pkg.TypesInfo.Types[expr]; ok {
			return tv, nil
		}
	}
	return types.TypeAndValue{}, fmt.Errorf("typeAndValue for expr %T:%v not found", expr, expr)
}

func (tr typeResolver) loadRpcParamList(apiName string, list []*ast.Field) ([]rpcParam, error) {
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
			params = append(params, param)
		} else {
			for _, name := range field.Names {
				param, err := newRpcParam(pos, name.Name, t)
				if err != nil {
					return nil, fmt.Errorf("newRpcParam on pos %d: %w", pos, err)
				}
				params = append(params, param)
			}
		}
	}
	return params, nil
}

func (tr typeResolver) typeIsContext(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}

	if named.Obj().Name() != "Context" {
		return false
	}

	pkg := named.Obj().Pkg()
	if pkg == nil || pkg.Path() != "context" {
		return false
	}

	return true
}

func (tr typeResolver) newType(apiName string, t types.Type, astExpr ast.Expr) (Type, error) {
	ni, utAst, err := tr.unwrapNamedOrPassThrough(t, astExpr)
	if err != nil {
		return nil, fmt.Errorf("unwrapNamedOrPassThrough(): %w", err)
	}

	if tr.typeIsContext(t) {
		if ni == nil {
			return nil, fmt.Errorf("namedInfo for context is nil")
		}
		return newContextType(*ni), nil
	}

	if types.Implements(t, tr.binMarshaler) && types.Implements(types.NewPointer(t), tr.binUnmarshaler) {
		return tr.newBinaryMarshalerType(ni)
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

func (tr typeResolver) unwrapNamedOrPassThrough(t types.Type, astExpr ast.Expr) (*namedInfo, ast.Expr, error) {
	var is importSpec
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
		// we try to find the "pkgName"  in pkgName.VarType
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

// returns ex: "time" if ast.Expr is at "time.Time". otherwise ""
func packagePrefixFromAst(astExpr ast.Expr) string {
	if selExpr, ok := astExpr.(*ast.SelectorExpr); ok {
		if ident, ok := selExpr.X.(*ast.Ident); ok {
			return ident.Name
		}
	}
	return ""
}
