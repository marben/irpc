package main

import (
	"fmt"
	"go/format"
	"strings"
)

type formattingErr struct {
	formattingError error
	unformattedCode string
}

func (e *formattingErr) Error() string {
	return e.formattingError.Error()
}

func (e *formattingErr) Unwrap() error {
	return e.formattingError
}

type genFile struct {
	pkg          string
	imports      orderedSet[string]
	uniqueBlocks orderedSet[string]
}

func newGenFile(pkg string) *genFile {
	return &genFile{
		pkg:          pkg,
		imports:      newOrderedSet[string](),
		uniqueBlocks: newOrderedSet[string](),
	}
}

func (gen *genFile) addImport(imp ...string) {
	for _, i := range imp {
		// skip import of our own package
		if i == gen.pkg {
			continue
		}

		gen.imports.add(i)
	}
}

func (gen *genFile) addUniqueBlock(block string) {
	gen.uniqueBlocks.add(block)
}

// returns formatted code.
// on formatting error, returns the raw version for review (although syntactically incorrect) and error
func (gen *genFile) formatted() (string, error) {
	raw := gen.raw()
	formatted, err := format.Source([]byte(raw))
	if err != nil {
		return "", &formattingErr{formattingError: err, unformattedCode: raw}
	}

	return string(formatted), nil
}

// returns unformatted code
func (gen *genFile) raw() string {
	sb := &strings.Builder{}
	// HEADER
	headerStr := `// Code generated by irpc generator; DO NOT EDIT
	package %s
	`
	fmt.Fprintf(sb, headerStr, gen.pkg)

	// IMPORTS
	if gen.imports.len() != 0 {
		sb.WriteString("import(\n")
		for _, imp := range gen.imports.ordered {
			fmt.Fprintf(sb, "\"%s\"\n", imp)
		}
		sb.WriteString("\n)\n")
	}

	// UNIQUE BLOCKS
	for _, b := range gen.uniqueBlocks.ordered {
		fmt.Fprintf(sb, "\n%s\n", b)
	}

	return sb.String()
}
