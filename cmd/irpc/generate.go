package main

import (
	"crypto/sha256"
	"fmt"
	"go/types"
	"regexp"
	"slices"
	"strings"
	"unicode"
)

var (
	irpcGenImport = importSpec{path: "github.com/marben/irpc/irpcgen", pkgName: "irpcgen"}
	fmtImport     = importSpec{path: "fmt", pkgName: "fmt"} // todo: figure out the imports with importer, like we do with binmarshaller?
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

// makes sure varname is unique among existing vars. extends it with enough "_" if necessary.
// returns the new unique var name
func generateUniqueVarname(varName string, existingVars []funcParam) string {
loop:
	for _, vf := range existingVars {
		if vf.identifier == varName {
			varName += "_"
			goto loop
		}
	}
	return varName
}

// encoding/decoding of slices requires unique iterator names
// we generate them from variable name, but we need to remove some characters
// replaces '[' ']' '.' with underscore
func generateIteratorName(existingVars []string) string {
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

func generateKeyValueIteratorNames(existingVars []string) (kIt, vIt string) {
	// start with k, v ; if those are taken, continue to k2, v2; k3, v3 etc

	if !slices.Contains(existingVars, "k") && !slices.Contains(existingVars, "v") {
		return "k", "v"
	}

	// we start adding numbers starting with 2
	for i := 2; ; i++ {
		k := fmt.Sprintf("k%d", i)
		v := fmt.Sprintf("v%d", i)
		if !slices.Contains(existingVars, k) && !slices.Contains(existingVars, v) {
			return k, v
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

// sanitizeInterfaceName builds a readable identifier for an inline interface.
// Example: interface{Age() int; Name() string} → "Iface_Age_Name"
func sanitizeInterfaceName(iface *types.Interface) string {
	var parts []string

	// Collect method names
	for i := 0; i < iface.NumMethods(); i++ {
		m := iface.Method(i)
		parts = append(parts, m.Name())
	}

	// If no methods, just call it "Empty"
	if len(parts) == 0 {
		parts = append(parts, "Empty")
	}

	// Join with underscores
	name := "Iface_" + strings.Join(parts, "_")

	// Final cleanup: allow only letters, numbers, underscore
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	name = re.ReplaceAllString(name, "_")

	return name
}
