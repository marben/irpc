package main

import (
	"fmt"
	"go/ast"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// represents function parameters/return value
type rpcParam struct {
	pos   int // position in field
	name  string
	tDesc typeDesc
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
