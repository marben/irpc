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
	uint64Encoder        = newSymmetricDirectCallEncoder("UvarInt64", "uint64")
	boolEncoder          = newSymmetricDirectCallEncoder("Bool", "bool")
	binMarshallerEncoder = directCallEncoder{
		encFuncName:        "BinaryMarshaler",
		decFuncName:        "BinaryUnmarshaler",
		underlyingTypeName: "encoding.BinaryUnmarshaler", // todo: needed?
		needsCasting:       false,
	}
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

func newSymmetricDirectCallEncoder(encDecFunc string, underlyingTypeName string) directCallEncoder {
	return directCallEncoder{
		encFuncName:        encDecFunc,
		decFuncName:        encDecFunc,
		underlyingTypeName: underlyingTypeName,
	}
}

type directCallEncoder struct {
	encFuncName        string
	decFuncName        string
	underlyingTypeName string
	needsCasting       bool // if named, otherwise ""
}

func (e directCallEncoder) encode(varId string, existingVars varNames, q *qualifier) string {
	var varParam string
	if e.needsCasting {
		varParam = fmt.Sprintf("%s(%s)", e.underlyingTypeName, varId)
	} else {
		varParam = varId
	}

	return fmt.Sprintf(`if err := e.%s(%s); err != nil {
		return fmt.Errorf("serialize %s of type '%s': %%w", err)
	}
	`, e.encFuncName, varParam, varId, e.underlyingTypeName)
}

func (e directCallEncoder) decode(varId string, existingVars varNames, _ *qualifier) string {
	var varParam string
	if e.needsCasting {
		varParam = fmt.Sprintf("(*%s)(&%s)", e.underlyingTypeName, varId)
	} else {
		varParam = "&" + varId
	}
	sb := &strings.Builder{}
	fmt.Fprintf(sb, `if err := d.%s(%s); err != nil{
		return fmt.Errorf("deserialize %s of type '%s': %%w",err)
	}
	`, e.decFuncName, varParam, varId, e.underlyingTypeName)
	return sb.String()
}

func (e directCallEncoder) codeblock(q *qualifier) string {
	return ""
}
