package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	// fmt.Printf("Running %s go on %s\n", os.Args[0], os.Getenv("GOFILE"))

	// cwd, err := os.Getwd()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("  cwd = %s\n", cwd)
	// fmt.Printf("  os.Args = %#v\n", os.Args)

	// for _, ev := range []string{"GOARCH", "GOOS", "GOFILE", "GOLINE", "GOPACKAGE", "DOLLAR"} {
	// 	fmt.Println("  ", ev, "=", os.Getenv(ev))
	// }

	if err := run(); err != nil {
		log.Fatalf("run failed: %+v", err)
	}
}

func run() error {
	srcFileName := os.Getenv("GOFILE")
	if srcFileName == "" {
		return fmt.Errorf("$GOFILE doesn't contain the requested parse file")
	}

	fd, err := newRpcFileDesc(srcFileName)
	if err != nil {
		return fmt.Errorf("newRpcFileDesc for file '%s': %w", srcFileName, err)
	}
	fmt.Print(fd.print())

	g, err := newGenerator(fd)
	if err != nil {
		return fmt.Errorf("newGenerator for file '%s': %w", srcFileName, err)
	}

	// OUTPUT FILE
	genFileName, err := generatedFileName(srcFileName)
	if err != nil {
		return fmt.Errorf("figure out generated file name: %w", err)
	}
	log.Println("generating file:", genFileName)
	outFile, err := os.Create(genFileName)
	if err != nil {
		return fmt.Errorf("create generated file '%s': %w", genFileName, err)
	}
	defer outFile.Close()

	if err := g.write(outFile); err != nil {
		return fmt.Errorf("generator write to '%s': %w", genFileName, err)
	}

	return nil
}

func generatedFileName(inputGoFile string) (string, error) {
	goSuffix := ".go"
	n, found := strings.CutSuffix(inputGoFile, goSuffix)
	if !found {
		return "", fmt.Errorf("not a go file. '%s' doesn't end with '%s' suffix", inputGoFile, goSuffix)
	}

	return n + "_gen" + goSuffix, nil
}
