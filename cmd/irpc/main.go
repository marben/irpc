package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("error: %+v\n", err)
		os.Exit(1)
	}
}

func run() error {
	flag.Parse()

	var inputFiles []string
	if flag.NArg() > 0 {
		// input files were passed as arguments
		inputFiles = flag.Args()
	} else {
		// we should be run by "//go generate", which sets some env variables for us ("GOARCH", "GOOS", "GOFILE", "GOLINE", "GOPACKAGE", "DOLLAR")
		f := os.Getenv("GOFILE")
		if f == "" {
			return fmt.Errorf("either specify the input file as command parameter, or run the irpc command using `go generate`")
		}
		inputFiles = append(inputFiles, f)
	}

	for _, f := range inputFiles {
		if err := processInputFile(f); err != nil {
			return fmt.Errorf("processing file %q: %w", f, err)
		}
	}

	return nil
}

func generateFile(filename string, hash []byte) (string, error) {
	absFilePath, err := filepath.Abs(filename)
	if err != nil {
		return "", fmt.Errorf("filepath.Abs(): %w", err)
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
		return "", fmt.Errorf("packages.Load(): %w", err)
	}

	// packages.Load() seems to be designed to parse multiple files (passed in go command style (./... etc))
	// we only care about one file though, therefore it should always be the first in the array in following code

	if len(packages) != 1 {
		return "", fmt.Errorf("unexpectedly %d packages returned for file %q", len(packages), filename)
	}

	pkg := packages[0]

	fileAst, err := findASTForFile(pkg, filename)
	if err != nil {
		return "", fmt.Errorf("couldn't find ast for given file %s", filename)
	}

	imports := newImportsList()
	for _, impSpec := range fileAst.Imports {
		// todo: impSpec seems to have a value of 'name', which can be '.', nil etc...we should use it to make imports in generated files
		pkgName := pkg.TypesInfo.PkgNameOf(impSpec)
		imports.add(impSpec, pkgName)
	}

	gen, err := newGenerator(pkg.Name, pkg.PkgPath, hash, pkg.TypesInfo)
	if err != nil {
		return "", fmt.Errorf("failed to create generator")
	}

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
				gen.addInterface(ts.Name.String(), iface)
			}
		}
	}

	buf := bytes.NewBuffer(nil)
	gen.write(buf)

	return buf.String(), nil
}

func processInputFile(inputFile string) error {
	unshashed, err := generateFile(inputFile, nil)
	if err != nil {
		return fmt.Errorf("genFile2(): %w", err)
	}

	hasher := sha256.New()
	hasher.Write([]byte(unshashed))
	hash := hasher.Sum(nil)

	hashed, err := generateFile(inputFile, hash)
	if err != nil {
		return fmt.Errorf("hashed enerator for file %q: %w", inputFile, err)
	}

	// OUTPUT FILE
	genFileName, err := generatedFileName(inputFile)
	if err != nil {
		return fmt.Errorf("figure out generated file name: %w", err)
	}
	fmt.Println("generating file:", genFileName)
	outFile, err := os.Create(genFileName)
	if err != nil {
		return fmt.Errorf("create generated file '%s': %w", genFileName, err)
	}
	defer outFile.Close()

	if _, err := outFile.Write([]byte(hashed)); err != nil {
		return fmt.Errorf("write to file %q: %w", genFileName, err)
	}

	return nil
}

func generatedFileName(inputGoFile string) (string, error) {
	goSuffix := ".go"
	n, found := strings.CutSuffix(inputGoFile, goSuffix)
	if !found {
		return "", fmt.Errorf("not a go file. '%s' doesn't end with '%s' suffix", inputGoFile, goSuffix)
	}

	return n + "_irpc" + goSuffix, nil
}
