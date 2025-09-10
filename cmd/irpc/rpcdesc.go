package main

import (
	"fmt"
	"go/ast"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// represents function parameters/return value
// todo: seems unnecessary. could construct funcParam directly
type rpcParam struct {
	pos  int // position in field
	name string
	typ  Type
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

func findPackageForFile(pkgs []*packages.Package, file string) (*packages.Package, error) {
	abs, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}
	for _, pkg := range pkgs {
		for _, f := range pkg.GoFiles {
			if sameFile(f, abs) {
				return pkg, nil
			}
		}
	}
	return nil, fmt.Errorf("couldn't find *packages.Package for file %q", abs)
}

func (g *generator) findPackageForPackagePath(pkgPath string) (*packages.Package, error) {
	/*
		log.Printf("looking for package with path %q", pkgPath)
		for _, pkg := range g.allPkgs {
			param := pkg.PkgPath
			log.Printf("looking at pkg %q", param)
			if param == pkgPath {
				log.Printf("and the winner is!: %q", pkg.PkgPath)
				return pkg, nil
			}
		}

		return nil, fmt.Errorf("package with path %q not found", pkgPath)
	*/
	pkg, ok := findPackageForPackagePath(g.allPkgs, pkgPath)
	if !ok {
		return nil, fmt.Errorf("couldn't find type's package")
	}
	return pkg, nil
}

func findPackageForPackagePath(pkgs []*packages.Package, pkgPath string) (*packages.Package, bool) {
	// log.Printf("looking for package with path %q", pkgPath)
	for _, pkg := range pkgs {
		param := pkg.PkgPath
		// log.Printf("looking at pkg %q", param)
		if param == pkgPath {
			// log.Printf("and the winner is!: %q", pkg.PkgPath)
			return pkg, true
		}
	}

	return nil, false
}

func sameFile(a, b string) bool {
	// log.Printf("comparing file %q and %q", a, b)
	ra, err1 := filepath.EvalSymlinks(a)
	rb, err2 := filepath.EvalSymlinks(b)
	if err1 == nil && err2 == nil {
		return ra == rb
	}
	return a == b
}
