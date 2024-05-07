package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
)

// representation of irpc service/client defining interface (marked with //go:generate)
type rpcInterface struct {
	typeSpec *ast.TypeSpec
	astIface *ast.InterfaceType
	// typesInterface *types.Interface // todo: probably unnecessary
	methods []rpcMethod
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
	// typesInterface, ok := ifaceTypeObject.Type().(*types.Interface)
	// if !ok {
	// 	return rpcInterface{}, fmt.Errorf("iface's '%s' type definition is not *types.Interface. it is: %s", typeSpec.Name.Name, ifaceTypeObject.Type())
	// }

	return rpcInterface{typeSpec: typeSpec, astIface: astIface, methods: ms}, nil
}

func (i rpcInterface) name() string {
	return i.typeSpec.Name.String()
}

func (i rpcInterface) print(prefix string) string {
	if len(i.methods) == 0 {
		return fmt.Sprintf("%stype %s interface{}\n", prefix, i.name())
	}

	s := fmt.Sprintf("%stype %s interface{\n", prefix, i.name())
	for _, m := range i.methods {
		s += fmt.Sprintf("%s\t%s\n", prefix, m.print())
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
	results, err := loadRpcParamList(typesInfo, astFuncType.Results.List)
	if err != nil {
		return rpcMethod{}, fmt.Errorf("results list load for %s: %w", methodName, err)
	}
	return rpcMethod{name: methodName, astField: astField, params: params, results: results}, nil
}

func (m rpcMethod) print() string {
	params := "(" + printParamList(m.params) + ")"
	results := printParamList(m.results)
	if len(m.results) > 1 {
		results = "(" + results + ")"
	}
	return fmt.Sprintf("%s%s%s", m.name, params, results)
}

// comma separated list of variable names and types. ex: "a int, b float64"
func printParamList(list []rpcParam) string {
	s := ""
	for i, p := range list {
		s += fmt.Sprintf("%s %s", p.name, p.typeName)
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
			param, err := newRpcParam(pos, "", tv)
			if err != nil {
				return nil, fmt.Errorf("newRpcParam on pos %d: %w", pos, err)
			}
			params = append(params, param)
		} else {
			for _, name := range field.Names {
				param, err := newRpcParam(pos, name.Name, tv)
				if err != nil {
					return nil, fmt.Errorf("newRpcParam on pos %d: %w", pos, err)
				}
				params = append(params, param)
			}
		}
	}
	return params, nil
}

type rpcParam struct {
	pos          int
	name         string
	typeName     string
	typeAndValue types.TypeAndValue // todo: is it needed?
}

func newRpcParam(position int, name string, tv types.TypeAndValue) (rpcParam, error) {
	var typeName string
	switch t := tv.Type.(type) {
	case *types.Basic:
		typeName = t.Name()
	case *types.Named:
		typeName = t.Obj().Name()
	case *types.Slice:
		typeName = t.String()
	default:
		return rpcParam{}, fmt.Errorf("unsupported param type %T", t)
	}

	return rpcParam{
		pos:          position,
		name:         name,
		typeName:     typeName,
		typeAndValue: tv,
	}, nil
}

type rpcFileDesc struct {
	filename    string
	packageName string
	ifaces      []rpcInterface
}

func newRpcFileDesc(filename string) (rpcFileDesc, error) {
	fSet := token.NewFileSet()
	fileAst, err := parser.ParseFile(fSet, filename, nil, parser.AllErrors)
	if err != nil {
		return rpcFileDesc{}, fmt.Errorf("parse file '%s': %w", filename, err)
	}

	// print AST
	// ast.Print(fSet, fileAst)

	conf := types.Config{Importer: importer.Default()}
	info := &types.Info{
		Defs:  make(map[*ast.Ident]types.Object),
		Types: make(map[ast.Expr]types.TypeAndValue),
	}
	pkg, err := conf.Check("", fSet, []*ast.File{fileAst}, info)
	if err != nil {
		return rpcFileDesc{}, fmt.Errorf("conf.Check(): %w", err)
	}
	fmt.Printf("loaded package: %s\n", pkg.Name())

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
			iface, err := newRpcInterface(info, ts, it)
			if err != nil {
				return rpcFileDesc{}, fmt.Errorf("newRpcInterface %s: %w", ts.Name.String(), err)
			}
			ifaces = append(ifaces, iface)
		}
	}

	return rpcFileDesc{
		filename:    filename,
		ifaces:      ifaces,
		packageName: fileAst.Name.Name}, nil
}

func (f rpcFileDesc) print() string {
	s := fmt.Sprintf("%s:\n", f.filename)
	for _, i := range f.ifaces {
		s += i.print("\t")
	}
	return s
}
