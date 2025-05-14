package main

import (
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"
	"unicode"
)

const (
	irpcImport    = "github.com/marben/irpc"
	binaryImport  = "encoding/binary"
	fmtImport     = "fmt"
	ioImport      = "io"
	bytesImport   = "bytes"
	mathImport    = "math"
	contextImport = "context"
)

// generates unique service hash based on generated code's hash (without the hash;) and service name
func generateServiceIdHash(fileHash []byte, serviceName string, len int) []byte {
	input := append(fileHash, []byte(serviceName)...)
	hsh := sha256.Sum256(input)
	return hsh[:len]
}

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
		if vf.name() == varName {
			varName += "_"
			goto loop
		}
	}
	return varName
}

// encoding/decoding of slices requires unique iterator names
// we generate them from variable name, but we need to remove some characters
// replaces '[' ']' '.' with underscore
func generateSliceIteratorName(existingVars []string) string {
	possibleNames := []string{"i", "j", "k", "l", "m", "n"}
	for {
		for i, n := range possibleNames {
			if !slices.Contains(existingVars, n) {
				return n
			}
			possibleNames[i] = n + "_"
		}
	}
}

// returns byte slice definition ex: "[]byte{127, 3, 255}"
func byteSliceLiteral(in []byte) string {
	sb := strings.Builder{}
	for i, b := range in {
		sb.WriteString(fmt.Sprintf("%d", b))
		if i != len(in)-1 {
			sb.WriteString(", ")
		}
	}
	return fmt.Sprintf("[]byte{%s}", sb.String())
}
