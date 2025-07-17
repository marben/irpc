package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// representation of irpc service/client defining interface (marked with //go:generate)
type rpcInterface struct {
	typeSpec *ast.TypeSpec
	astIface *ast.InterfaceType
	methods  []rpcMethod
}

func newRpcInterface(typesInfo *types.Info, typeSpec *ast.TypeSpec, astIface *ast.InterfaceType) (rpcInterface, error) {
	ms := []rpcMethod{}
	for _, method := range astIface.Methods.List {
		m, err := newRpcMethod(typesInfo, method)
		if err != nil {
			return rpcInterface{}, fmt.Errorf("newRpcMethod: %w", err)
		}
		ms = append(ms, m)
	}

	ifaceTypeObject := typesInfo.Defs[typeSpec.Name]
	if ifaceTypeObject == nil {
		return rpcInterface{}, fmt.Errorf("couldn't find interface '%s' type definition", typeSpec.Name.Name)
	}

	return rpcInterface{typeSpec: typeSpec, astIface: astIface, methods: ms}, nil
}

func (i rpcInterface) name() string {
	return i.typeSpec.Name.String()
}

func (i rpcInterface) print(q types.Qualifier, prefix string) string {
	if len(i.methods) == 0 {
		return fmt.Sprintf("%stype %s interface{}\n", prefix, i.name())
	}

	s := fmt.Sprintf("%stype %s interface{\n", prefix, i.name())
	for _, m := range i.methods {
		s += fmt.Sprintf("%s\t%s\n", prefix, m.print(q))
	}
	s += fmt.Sprintf("%s}\n", prefix)

	return s
}

type rpcMethod struct {
	name            string
	astField        *ast.Field
	params, results []rpcParam
}

func newRpcMethod(typesInfo *types.Info, astField *ast.Field) (rpcMethod, error) {
	astFuncType, ok := astField.Type.(*ast.FuncType)
	if !ok {
		return rpcMethod{}, fmt.Errorf("*ast.Field %v is not *ast.FuncType", astField)
	}

	var methodName string
	if len(astField.Names) == 0 {
		methodName = "<no name>"
	} else {
		methodName = astField.Names[0].Name
	}

	params, err := loadRpcParamList(typesInfo, astFuncType.Params.List)
	if err != nil {
		return rpcMethod{}, fmt.Errorf("params list load for %s: %w", methodName, err)
	}

	var results []rpcParam
	if astFuncType.Results != nil {
		results, err = loadRpcParamList(typesInfo, astFuncType.Results.List)
		if err != nil {
			return rpcMethod{}, fmt.Errorf("results list load for %s: %w", methodName, err)
		}
	}
	return rpcMethod{name: methodName, astField: astField, params: params, results: results}, nil
}

func (m rpcMethod) print(q types.Qualifier) string {
	params := "(" + printParamList(q, m.params) + ")"
	results := printParamList(q, m.results)
	if len(m.results) > 1 {
		results = "(" + results + ")"
	}
	return fmt.Sprintf("%s%s%s", m.name, params, results)
}

// comma separated list of variable names and types. ex: "a int, b float64"
func printParamList(q types.Qualifier, list []rpcParam) string {
	s := ""
	for i, p := range list {
		s += fmt.Sprintf("%s %s", p.name, p.typeName(q))
		if i != len(list)-1 {
			s += ","
		}
	}
	return s
}

func loadRpcParamList(typesInfo *types.Info, list []*ast.Field) ([]rpcParam, error) {
	params := []rpcParam{}
	for pos, field := range list {
		tv, ok := typesInfo.Types[field.Type]
		if !ok {
			fmt.Printf("couldn't determine fileld's %v type and value", field)
			continue
		}
		if field.Names == nil {
			// parameter doesn't have name, just a type (typically function returns)
			param, err := newRpcParam(pos, "", tv.Type)
			if err != nil {
				return nil, fmt.Errorf("newRpcParam on pos %d: %w", pos, err)
			}
			params = append(params, param)
		} else {
			for _, name := range field.Names {
				param, err := newRpcParam(pos, name.Name, tv.Type)
				if err != nil {
					return nil, fmt.Errorf("newRpcParam on pos %d: %w", pos, err)
				}
				params = append(params, param)
			}
		}
	}
	return params, nil
}

// represents function parameters/return value
type rpcParam struct {
	pos  int // position in field
	name string
	typ  types.Type
}

func newRpcParam(position int, name string, typ types.Type) (rpcParam, error) {
	return rpcParam{
		pos:  position,
		name: name,
		typ:  typ,
	}, nil
}

func (rp rpcParam) typeName(q types.Qualifier) string {
	// log.Printf("type: %#v", rp.typ)
	// log.Printf("string: %s", rp.typ.String())
	// log.Printf("TypeString: %s", rp.typ.String())
	return types.TypeString(rp.typ, q)
}

type importInfo struct {
	importSpec *ast.ImportSpec // todo: remove, once we know what exactly we need
	pkgName    *types.PkgName  // todo: remove, once we know what exactly we need
}

type importsList struct {
	imports []importInfo
}

func (il *importsList) add(spec *ast.ImportSpec, pkgName *types.PkgName) {
	il.imports = append(il.imports, importInfo{spec, pkgName})
}

func newImportsList() *importsList {
	return &importsList{}
}

type rpcFileDesc struct {
	filename    string
	packageName string
	packagePath string // our package path (within module) (not tested outside module)
	ifaces      []rpcInterface
	imports     *importsList
}

func loadRpcFileDesc(filename string) (rpcFileDesc, error) {
	absFilePath, err := filepath.Abs(filename)
	if err != nil {
		return rpcFileDesc{}, fmt.Errorf("filepath.Abs(): %w", err)
	}

	dir := filepath.Dir(absFilePath)

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedDeps | packages.NeedImports | packages.NeedSyntax | packages.NeedTypesInfo |
			packages.NeedFiles | packages.NeedName | packages.NeedCompiledGoFiles | packages.NeedExportFile | packages.NeedSyntax |
			packages.NeedModule,
		Dir: dir,
	}

	// we need to load all the files in directory, otherwise we get "command-line-arguments" as pkg paths
	// todo: maybe we need to use ./... or base it at the root of our module, to get all the deps? need to test/figure out
	packages, err := packages.Load(cfg, ".")
	if err != nil {
		return rpcFileDesc{}, fmt.Errorf("packages.Load(): %w", err)
	}

	// packages.Load() seems to be designed to parse multiple files (passed in go command style (./... etc))
	// we only care about one file though, therefore it should always be the first in the array in following code

	if len(packages) != 1 {
		return rpcFileDesc{}, fmt.Errorf("unexpectedly %d packages returned for file %q", len(packages), filename)
	}

	pkg := packages[0]

	fileAst, err := findASTForFile(pkg, filename)
	if err != nil {
		return rpcFileDesc{}, fmt.Errorf("couldn't find ast for given file %s", filename)
	}

	imports := newImportsList()
	for _, impSpec := range fileAst.Imports {
		// todo: impSpec seems to have a value of 'name', which can be '.', nil etc...we should use it to make imports in generated files
		pkgName := pkg.TypesInfo.PkgNameOf(impSpec)
		imports.add(impSpec, pkgName)
	}

	ifaces, err := loadRpcInterfaces(fileAst, pkg.TypesInfo)
	if err != nil {
		return rpcFileDesc{}, fmt.Errorf("loadRpcInterfaces: %w", err)
	}

	return rpcFileDesc{
		filename:    filename,
		packageName: pkg.Name,
		packagePath: pkg.PkgPath,
		ifaces:      ifaces,
		imports:     imports,
	}, nil
}

// func (fd *rpcFileDesc) packageName() string {
// 	return fd.pkg.Name
// }

func loadRpcInterfaces(fileAst *ast.File, tInfo *types.Info) ([]rpcInterface, error) {
	ifaces := []rpcInterface{}
	for _, decl := range fileAst.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			it, ok := ts.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}
			iface, err := newRpcInterface(tInfo, ts, it)
			if err != nil {
				return nil, fmt.Errorf("newRpcInterface %s: %w", ts.Name.String(), err)
			}
			ifaces = append(ifaces, iface)
		}
	}

	return ifaces, nil
}

func (fd *rpcFileDesc) print(q types.Qualifier) string {
	s := fmt.Sprintf("%s:\n", fd.filename)
	for _, i := range fd.ifaces {
		s += i.print(q, "\t")
	}
	return s
}

func findASTForFile(pkg *packages.Package, targetFile string) (*ast.File, error) {
	absTarget, err := filepath.Abs(targetFile)
	if err != nil {
		return nil, err
	}

	for i, f := range pkg.CompiledGoFiles {
		absFile, _ := filepath.Abs(f)
		if absFile == absTarget {
			return pkg.Syntax[i], nil
		}
	}

	return nil, fmt.Errorf("file %s not found in package", targetFile)
}
