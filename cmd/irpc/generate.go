package main

import (
	"crypto/sha256"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

var (
	irpcGenImport = importSpec{path: "github.com/marben/irpc/irpcgen", pkgName: "irpcgen"}
	fmtImport     = importSpec{path: "fmt", pkgName: "fmt"}
	contextImport = importSpec{path: "context", pkgName: "context"}
)

// generates unique service hash based on generated code's hash (without the hash;) and service name
func generateServiceIdHash(fileHash []byte, serviceName string, maxLen int) []byte {
	input := append(fileHash, []byte(serviceName)...)
	hsh := sha256.Sum256(input)
	l := min(maxLen, len(hsh))
	return hsh[:l]
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

// not a pointer type! - passing down the stack doesn't alter caller
type varNames []string

func (l *varNames) generateUniqueVarName(idealName string) string {
	if !l.contains(idealName) {
		l.addVarName(idealName)
		return idealName
	}
	for i := 2; ; i++ {
		testVar := idealName + strconv.Itoa(i)
		if !l.contains(testVar) {
			l.addVarName(testVar)
			return testVar
		}
	}
}

func (l *varNames) addVarName(vn ...string) {
	*l = append(*l, vn...)
}

func (l varNames) contains(vn string) bool {
	return slices.Contains(l, vn)
}

func (l *varNames) generateIteratorName() string {
	possibleNames := []string{"i", "j", "k", "l", "m", "n"}
	appendix := 2
	for {
		for i, n := range possibleNames {
			if !l.contains(n) {
				l.addVarName(n)
				return n
			}
			possibleNames[i] = n + strconv.Itoa(appendix)
		}
		appendix++
	}
}

func (l *varNames) generateKeyValueIteratorNames() (kIt, vIt string) {
	// start with k, v ; if those are taken, continue to k2, v2; k3, v3 etc

	if !l.contains("k") && !l.contains("v") {
		l.addVarName("k", "v")
		return "k", "v"
	}

	// we start adding numbers starting with 2
	for i := 2; ; i++ {
		k := fmt.Sprintf("k%d", i)
		v := fmt.Sprintf("v%d", i)
		if !l.contains(k) && !l.contains(v) {
			l.addVarName(k, v)
			return k, v
		}
	}
}

// returns byte slice definition in hex with max of 8 words pe line
// ex: "[]byte{0x12,0x03,0xfe}"
func byteSliceLiteral(in []byte) string {
	sb := strings.Builder{}
	if len(in) <= 8 {
		// all in just one line
		for i, v := range in {
			fmt.Fprintf(&sb, "%#02x", v)
			if i != len(in)-1 {
				sb.WriteString(",")
			}
		}
		return fmt.Sprintf("[]byte{%s}", sb.String())
	} else {
		// we make a block of max 8 nums
		for i, v := range in {
			fmt.Fprintf(&sb, "%#02x,", v)
			if (i+1)%8 == 0 {
				sb.WriteByte('\n')
			}
		}
		return fmt.Sprintf("[]byte{\n%s\n}", sb.String())
	}
}
