package main

import (
	"unicode"
)

const (
	irpcImport   = "tigershare/irpc"
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
