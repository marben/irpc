package main

import (
	"fmt"
	"strings"
)

// todo: is this still necessary?
type encoder interface {
	encode(varId string, existingVars varNames, q *qualifier) string // inline variable encode
	decode(varId string, existingVars varNames, q *qualifier) string // inline variable decode
	codeblock(q *qualifier) string                                   // requested encoder's code block at top level
}

var (
	boolEncoder = newDirectCallEncoder("Bool", "Bool", "bool", nil)
	lenEncoder  = newDirectCallEncoder("Len", "Len", "int", nil)
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

type directCallEncoder struct {
	encFuncName string
	decFuncName string
	typeName    string // the actual type name as in int, []byte, etc
	ni          *namedInfo
}

func newDirectCallEncoder(encFunc, decFunc string, typeName string, ni *namedInfo) directCallEncoder {
	return directCallEncoder{
		encFuncName: encFunc,
		decFuncName: decFunc,
		typeName:    typeName,
		ni:          ni,
	}
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
