package main

import (
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

func processInputFile(inputFile string) error {
	gen, err := newGenerator(inputFile)
	if err != nil {
		return fmt.Errorf("newGenerator(): %w", err)
	}

	// calculate hash of generated file
	hasher := sha256.New()
	if err := gen.generate(hasher, nil); err != nil {
		return fmt.Errorf("gen.generate(): %w", err)
	}
	hash := hasher.Sum(nil)

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

	// write the generated file to output file, but with hash filled in
	if err := gen.generate(outFile, hash); err != nil {
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
