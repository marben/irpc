package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/token"
	"go/types"
	"io"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

// id len specifies how long our service's id will be. currently the max is 32 bytes as we are using sha256 to generate them
// actual id to negotiate between endpoints desn't have to be full lenght (currently it's only 4 bytes)
const idLen = 32

type generator struct {
	inputPkg *packages.Package   // todo:remove
	allPkgs  []*packages.Package // todo:remove
	services []*apiGenerator
	// imports  orderedSet[importSpec] // todo:remove
	qual *qualifier
	tr   *typeResolver
}

func newGenerator(filename string) (*generator, error) {
	absFilePath, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("filepath.Abs(): %w", err)
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
		return nil, fmt.Errorf("packages.Load(): %w", err)
	}

	targetPkg, err := findPackageForFile(allPackages, filename)
	if err != nil {
		return nil, err
	}

	fileAst, err := findASTForFile(targetPkg, filename)
	if err != nil {
		return nil, fmt.Errorf("couldn't find ast for given file %s", filename)
	}

	// IMPORTS
	srcImports := newOrderedSet[importSpec]()
	for _, is := range fileAst.Imports {
		// log.Printf("file import spec: %q  %q", is.Name, is.Path.Value)
		var alias string
		if is.Name != nil {
			alias = is.Name.Name
		}
		path, err := strconv.Unquote(is.Path.Value)
		if err != nil {
			return nil, fmt.Errorf("strconv.Unquote(%q): %w", is.Path.Value, err)
		}

		if pkg, ok := findPackageForPackagePath(allPackages, path); ok {
			// package was parsed by go/packages library
			spec := importSpec{alias, path, pkg.Name}
			srcImports.add(spec)
			// log.Printf("added import: %+v", spec)
		} else {
			// must be from stdlib, which isn't provided by the packages lib
			imp := importer.Default()
			pkg, err := imp.Import(path)
			if err != nil {
				return nil, fmt.Errorf("importer.Import(%q): %w", path, err)
			}
			spec := importSpec{alias: alias, path: pkg.Path(), pkgName: pkg.Name()}
			srcImports.add(spec)
			// log.Printf("added import: %+v", spec)
		}
	}

	gen := &generator{
		inputPkg: targetPkg,
		allPkgs:  allPackages,
		// imports:  newOrderedSet[importSpec](),
	}

	tr, err := newTypeResolver(gen, targetPkg, allPackages)
	if err != nil {
		return nil, fmt.Errorf("newTypeResolver(): %w", err)
	}

	gen.qual = newQualifier(gen, tr, srcImports) // todo: ugly and wrong

	gen.tr = tr

	// INTERFACES
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
			if iface, ok := ts.Type.(*ast.InterfaceType); ok {
				api, err := newApiGenerator(gen, tr, ts.Name.String(), iface)
				if err != nil {
					return nil, fmt.Errorf("new apiGenerator: %w", err)
				}

				// todo: this way of adding imports is not cool
				// gen.addImport(fmtImport, irpcGenImport)
				// if len(api.methods) > 0 {
				// every FuncExecutor uses context
				// gen.addImport(contextImport)
				// }

				gen.services = append(gen.services, api)
			}
		}
	}

	return gen, nil
}

func newRpcParam(position int, name string, t Type) (rpcParam, error) {
	return rpcParam{
		pos:  position,
		name: name,
		typ:  t,
	}, nil
}

// todo: this is a leftover from when we were relient on ast only
// todo: this may come in handy later, but for now, it's here just for reference
func (g *generator) qualifiedTypeName(astExpr ast.Expr) (string, error) {
	var qualifier string
	if selExpr, ok := astExpr.(*ast.SelectorExpr); ok {
		if ident, ok := selExpr.X.(*ast.Ident); ok {
			qualifier = ident.Name
		}
	}

	qf := func(pkg *types.Package) string {
		return qualifier
	}

	tv, err := g.tr.findTypeAndValueForAst(astExpr)
	if err != nil {
		return "", fmt.Errorf("findTypeAndValueForAst: %w", err)
	}

	qualifiedTypeName := types.TypeString(tv.Type, qf)

	return qualifiedTypeName, nil
}

// if hash is nil, we generate service id with empty hash
//   - this is used during first run of generator
func (g *generator) generate(w io.Writer, hash []byte) error {
	codeBlocks := newOrderedSet[string]()

	// todo: service/client code is weirdly separated from each other
	// todo: shoul interwine them, make only one loop. but only after major refactor i am working on now....

	paramStructs := []paramStructGenerator{}

	// SERVICES
	for _, service := range g.services {
		codeBlocks.add(service.serviceCode(hash, g.qual))
		paramStructs = append(paramStructs, service.paramStructs()...)
	}

	// CLIENTS
	for _, service := range g.services {
		codeBlocks.add(service.clientCode(hash, g.qual))
	}

	// PARAM STRUCTS
	for _, p := range paramStructs {
		// we don't generate empty types (even though the generator is capable of generating them)
		// we use irpcgen.Empty(Ser/Deser) instead
		if !p.isEmpty() {
			codeBlocks.add(p.code(g.qual))
			for _, e := range p.encoders() {
				codeBlocks.add(e.codeblock())
			}
		}
	}

	// GENERATE
	rawOutput := g.genRaw(codeBlocks)

	// FORMAT
	formatted, err := format.Source([]byte(rawOutput))
	if err != nil {
		log.Println("formatting failed. writing raw code to output file anyway")
		if _, err := w.Write([]byte(rawOutput)); err != nil {
			return fmt.Errorf("writing unformatted code to file: %w", err)
		}
	}

	if _, err := w.Write([]byte(formatted)); err != nil {
		return fmt.Errorf("copy of generated code to file: %w", err)
	}

	return nil
}

/*
// todo: unused?
func (g *generator) addImport(imps ...importSpec) {
	// for _, imp := range imps {
	// 	if imp.path == g.inputPkg.PkgPath {
	// 		// we don't want to import our own directory
	// 		continue
	// 	}
	// 	g.imports.add(imp)
	// }
}
*/

func (g *generator) genRaw(codeBlocks orderedSet[string]) string {
	sb := &strings.Builder{}
	// HEADER
	headerStr := `// Code generated by irpc generator; DO NOT EDIT
	package %s
	`
	fmt.Fprintf(sb, headerStr, g.inputPkg.Name)

	if g.qual.usedImports.len() != 0 {
		sb.WriteString("import(\n")
		for _, imp := range g.qual.usedImports.ordered {
			fmt.Fprintf(sb, "%s \"%s\"\n", imp.alias, imp.path)
		}
		sb.WriteString("\n)\n")
	}

	// UNIQUE BLOCKS
	for _, b := range codeBlocks.ordered {
		fmt.Fprintf(sb, "\n%s\n", b)
	}

	return sb.String()
}
