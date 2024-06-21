package main

import (
	"unicode"
)

const (
	irpcImport   = "github.com/marben/irpc/pkg/irpc"
	binaryImport = "encoding/binary"
	fmtImport    = "fmt"
	ioImport     = "io"
	bytesImport  = "bytes"
	mathImport   = "math"
)

// generates the 'NewSomething(){}' function name
func generateStructConstructorName(structName string) string {
	runesIn := []rune(structName)
	if len(runesIn) == 0 {
		panic("cannot generate struct constructor name from empty string")
	}

	var rtn []rune
	if unicode.IsUpper(runesIn[0]) {
		rtn = []rune("New")
	} else {
		rtn = []rune("new")
	}

	rtn = append(rtn, unicode.ToUpper(runesIn[0]))
	rtn = append(rtn, runesIn[1:]...)
	return string(rtn)
}

// makes sure varname is unique among existing vars. extends it with enough "_" if necessary.
// returns the new unique var name
func generateUniqueVarname(varName string, existingVars []varField) string {
loop:
	for _, vf := range existingVars {
		if vf.name == varName {
			varName += "_"
			goto loop
		}
	}
	return varName
}
