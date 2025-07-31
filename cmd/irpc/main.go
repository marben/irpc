package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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

func generateFile(fd rpcFileDesc, hash []byte) (string, error) {
	gen, err := newGenerator(fd, hash, fd.typesInfo)
	if err != nil {
		return "", fmt.Errorf("failed to create generator")
	}
	for _, iface := range fd.ifaces {
		if err := gen.addInterface(iface); err != nil {
			return "", fmt.Errorf("addInterface(%s):%w", iface.name(), err)
		}
	}

	buf := bytes.NewBuffer(nil)
	gen.write(buf)

	return buf.String(), nil
}

func processInputFile(inputFile string) error {
	fd, err := loadRpcFileDesc(inputFile)
	if err != nil {
		return fmt.Errorf("loadRpcFileDesc for file '%s': %w", inputFile, err)
	}
	// fmt.Print(fd.print())

	unshashed, err := generateFile(fd, nil)
	if err != nil {
		return fmt.Errorf("genFile2(): %w", err)
	}

	hasher := sha256.New()
	hasher.Write([]byte(unshashed))
	hash := hasher.Sum(nil)

	hashed, err := generateFile(fd, hash)
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
