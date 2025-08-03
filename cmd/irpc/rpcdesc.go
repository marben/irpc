package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// comma separated list of variable names and types. ex: "a int, b float64"
func printParamList(q types.Qualifier, list []rpcParam) string {
	s := ""
	for i, p := range list {
		s += fmt.Sprintf("%s %s", p.name, p.tDesc.qualifiedTypeName)
		if i != len(list)-1 {
			s += ","
		}
	}
	return s
}

// represents function parameters/return value
type rpcParam struct {
	pos   int // position in field
	name  string
	imp   *importSpec
	tDesc typeDesc
}

type importInfo struct {
	importSpec *ast.ImportSpec // todo: remove, once we know what exactly we need
	pkgName    *types.PkgName  // todo: remove, once we know what exactly we need
}

func (iif importInfo) Path() string {
	return iif.pkgName.Imported().Path()
}

func (iif importInfo) Alias() string {
	return iif.importSpec.Name.Name
}

func (iif importInfo) String() string {
	return fmt.Sprintf("importSpec.Name: %q, importSpec.Path: %q, pkgName.Name: %q, pkgName.Imported.Path: %q, pkgName.Pkg().Path(): %q",
		iif.importSpec.Name.Name, iif.importSpec.Path.Value, iif.pkgName.Name(), iif.pkgName.Imported().Path(), iif.pkgName.Pkg().Path())
}

type importsList struct {
	imports []importInfo
}

func newImportsList() *importsList {
	return &importsList{}
}

func (il *importsList) add(spec *ast.ImportSpec, pkgName *types.PkgName) {
	iif := importInfo{spec, pkgName}
	il.imports = append(il.imports, iif)
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
