package main

import (
	"fmt"
	"strings"
)

type encoder interface {
	encode(varId string, existingVars varNames, q *qualifier) string // inline variable encode
	decode(varId string, existingVars varNames, q *qualifier) string // inline variable decode
	codeblock(q *qualifier) string                                   // requested encoder's code block at top level
}

var (
	boolEncoder          = newSymmetricDirectCallEncoder("Bool", "bool", nil)
	binMarshallerEncoder = directCallEncoder{
		encFuncName: "BinaryMarshaler",
		decFuncName: "BinaryUnmarshaler",
		typeName:    "encoding.BinaryUnmarshaler",
	}
	lenEncoder = newSymmetricDirectCallEncoder("Len", "int", nil)
)

type importSpec struct {
	alias   string // "myCtx" in `import myCtx "context"``
	path    string // fully qualifies the package
	pkgName string // the "context" in "context.Context"
}

// packageQualifier is import alias if defined, otherwise simply the package name
func (is importSpec) packageQualifier() string {
	if is.alias != "" {
		return is.alias
	}
	return is.pkgName
}

func newSymmetricDirectCallEncoder(encDecFunc string, typeName string, ni *namedInfo) directCallEncoder {
	return directCallEncoder{
		encFuncName: encDecFunc,
		decFuncName: encDecFunc,
		typeName:    typeName,
		ni:          ni,
	}
}

type directCallEncoder struct {
	encFuncName string
	decFuncName string
	typeName    string // the actual type name as in int, []byte, etc
	ni          *namedInfo
}

func (e directCallEncoder) needsCasting() bool {
	if e.ni != nil && e.ni.namedName != e.typeName {
		return true
	}
	return false
}

func (e directCallEncoder) encode(varId string, existingVars varNames, _ *qualifier) string {
	var varParam string
	if e.needsCasting() {
		varParam = fmt.Sprintf("%s(%s)", e.typeName, varId)
	} else {
		varParam = varId
	}

	return fmt.Sprintf(`if err := e.%s(%s); err != nil {
		return fmt.Errorf("serialize %s of type \"%s\": %%w", err)
	}
	`, e.encFuncName, varParam, varId, e.typeName)
}

func (e directCallEncoder) decode(varId string, existingVars varNames, _ *qualifier) string {
	var varParam string
	if e.needsCasting() {
		varParam = fmt.Sprintf("(*%s)(&%s)", e.typeName, varId)
	} else {
		varParam = "&" + varId
	}
	sb := &strings.Builder{}
	fmt.Fprintf(sb, `if err := d.%s(%s); err != nil{
		return fmt.Errorf("deserialize %s of type \"%s\": %%w",err)
	}
	`, e.decFuncName, varParam, varId, e.typeName)
	return sb.String()
}

func (e directCallEncoder) codeblock(q *qualifier) string {
	return ""
}
