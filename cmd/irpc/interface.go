package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"log"
	"regexp"
	"strings"
)

var _ Type = interfaceType{}

type interfaceType struct {
	// name         string
	fncs         []ifaceFunc
	implTypeName string
	// importSpec   importSpec
	ni      *namedInfo
	apiName string
	t       *types.Interface
}

func (tr *typeResolver) newInterfaceType(apiName string, ni *namedInfo, t *types.Interface, astExpr ast.Expr) (interfaceType, error) {
	// name, importSpec := tr.typeNameAndImport(t, astExpr)

	var ifaceAst *ast.InterfaceType
	if astExpr != nil {
		var ok bool
		ifaceAst, ok = astExpr.(*ast.InterfaceType)
		if !ok {
			return interfaceType{}, fmt.Errorf("provided ast is not *ast.InterfaceType, but %T", astExpr)
		}
	}

	log.Printf("iface ast: %v", ifaceAst)

	fncs := []ifaceFunc{}
	for i := 0; i < t.NumMethods(); i++ {
		method := t.Method(i)
		sig := method.Type().(*types.Signature)

		// we only handle interfaces that return values.
		// we call them and store them
		if sig.Params().Len() != 0 {
			return interfaceType{}, fmt.Errorf("unexpectedly params are not nil. they are: %+v", sig.Params())
		}

		var funcAst *ast.FuncType
		if ifaceAst != nil {
			fieldType := astTypeFieldFromFieldList(ifaceAst.Methods, i)
			var ok bool
			if funcAst, ok = fieldType.(*ast.FuncType); !ok {
				return interfaceType{}, fmt.Errorf("fieldType %q is of type %T", fieldType, fieldType)
			}
		}

		log.Printf("funcAst: %#v", funcAst)

		results := make([]ifaceImplVar, 0, sig.Results().Len())
		for i := 0; i < sig.Results().Len(); i++ {
			v := sig.Results().At(i)
			// fieldAst := interfaceMethodFieldOrNil(ifaceAst, i)
			// log.Printf("v.Type(): %#v %q", v.Type(), v.Type())
			// log.Printf("fieldAst: %#v", fieldAst)
			var typeAst ast.Expr
			if funcAst != nil {
				typeAst = astTypeFieldFromFieldList(funcAst.Results, i)
			}
			typ, err := tr.newType(apiName, v.Type(), typeAst)
			if err != nil {
				return interfaceType{}, fmt.Errorf("newType for var %q: %w", v, err)
			}

			results = append(results, ifaceImplVar{
				implStructParamName: fmt.Sprintf("_%s_%d_%s", method.Name(), i, v.Name()),
				f:                   field{name: v.Name(), t: typ},
			})
		}

		fncs = append(fncs, ifaceFunc{
			name:    method.Name(),
			results: results,
		})
	}

	// we need to make the type unique within package, because we want each file to be self contained
	// it would be possible to use filename instead of apiName, but that would confuse file renaming. this will do for now
	var n string
	if ni != nil {
		n = ni.namedName
	} else {
		n = sanitizeInterfaceName(t)
	}
	implTypeName := "_" + n + "_" + apiName + "_impl"
	log.Printf("implTypeName: %q", implTypeName)

	return interfaceType{
		implTypeName: implTypeName,
		fncs:         fncs,
		ni:           ni,
		apiName:      apiName,
		t:            t,
	}, nil
}

func (i interfaceType) implTypeNam() string {
	return i.implTypeName
}

// Name implements Type.
func (i interfaceType) Name(q *qualifier) string {
	if i.ni != nil {
		return i.ni.qualifiedName(q)
	}

	sb := strings.Builder{}
	sb.WriteString("interface {")
	for _, f := range i.fncs {
		sb.WriteString(f.signature(q))
		sb.WriteString(";")
	}
	sb.WriteString("}")
	return sb.String()
}

// we need concrete implementation of our interface
func (i interfaceType) codeblock(q *qualifier) string {
	sb := &strings.Builder{}

	// struct that will hold all our interface's return values
	fmt.Fprintf(sb, "type %s struct {\n", i.implTypeNam())
	for _, f := range i.fncs {
		for _, v := range f.results {
			fmt.Fprintf(sb, "%s %s\n", v.implStructParamName, v.f.t.Name(q))
		}
	}
	sb.WriteString("}\n")

	// we implement each interface function and return value stored withing the struct
	for _, f := range i.fncs {
		fmt.Fprintf(sb, "func (i %s)%s()(%s){\n", i.implTypeNam(), f.name, f.listNamesWithTypes(q))
		fmt.Fprintf(sb, "return %s\n", f.listImplNamesPrefixed("i."))
		sb.WriteString("}\n")
	}

	return sb.String()
}

// decode implements Type.
func (i interfaceType) decode(varId string, existingVars varNames, q *qualifier) string {
	log.Println("adding varname: ", varId)
	existingVars.addVarName(varId)
	sb := &strings.Builder{}
	sb.WriteString("{\n") // separate block
	fmt.Fprintf(sb, `var isNil bool
	%s
	if isNil {
		%s = nil
	} else {
	`, boolEncoder.decode("isNil", existingVars, q), varId)

	implVarName := existingVars.generateUniqueVarName("impl")
	log.Println("obrained varname: ", implVarName)
	log.Println("while the list is: ", existingVars)

	fmt.Fprintf(sb, "var %s %s\n", implVarName, i.implTypeNam())
	for _, f := range i.fncs {
		fmt.Fprintf(sb, "{ // %s()\n", f.name)
		for _, v := range f.results {
			sb.WriteString(v.f.t.decode(implVarName+"."+v.implStructParamName, existingVars, q))
		}
		sb.WriteString("}\n")
	}
	fmt.Fprintf(sb, "%s = %s\n", varId, implVarName)
	sb.WriteString("}\n") // else {
	sb.WriteString("}\n") // separate block

	return sb.String()
}

// encode implements Type.
func (i interfaceType) encode(varId string, existingVars varNames, q *qualifier) string {
	sb := &strings.Builder{}
	sb.WriteString("{\n") // separate block
	fmt.Fprintf(sb, `var isNil bool
	if %s == nil {
		isNil = true
	}
	%s
	`, varId, boolEncoder.encode("isNil", existingVars, q))
	sb.WriteString("if !isNil{\n")
	for _, f := range i.fncs {
		fmt.Fprintf(sb, "{ // %s()\n", f.name)
		for i, v := range f.results {
			sb.WriteString(v.implStructParamName)
			if i != len(f.results)-1 {
				sb.WriteString(",")
			}
		}
		fmt.Fprintf(sb, ":= %s.%s()\n", varId, f.name)
		for _, v := range f.results {
			sb.WriteString(v.f.t.encode(v.implStructParamName, existingVars, q))
		}
		sb.WriteString("}\n")
	}
	sb.WriteString("}\n") // if !isNil
	sb.WriteString("}\n") // separate block

	return sb.String()
}

type ifaceFunc struct {
	name    string
	results []ifaceImplVar
}

func (ifnc ifaceFunc) signature(q *qualifier) string {
	return ifnc.name + "() " + "(" + ifnc.listNamesWithTypes(q) + ")"
}

// comma separated list of variable names and types. ex: "a int, b float64"
func (ifnc ifaceFunc) listNamesWithTypes(q *qualifier) string {
	// todo: somehow share with paramStructGenerator's funcCallParams ?
	b := &strings.Builder{}
	for i, v := range ifnc.results {
		fmt.Fprintf(b, "%s %s", v.f.name, v.f.t.Name(q)) // todo: need qualified name?
		if i != len(ifnc.results)-1 {
			b.WriteString(",")
		}
	}
	return b.String()
}

func (ifnc ifaceFunc) listImplNamesPrefixed(prefix string) string {
	// todo: share with paramstructgenerator
	buf := bytes.NewBuffer(nil)
	for i, p := range ifnc.results {
		fmt.Fprintf(buf, "%s%s", prefix, p.implStructParamName)
		if i != len(ifnc.results)-1 {
			fmt.Fprintf(buf, ",")
		}
	}
	return buf.String()
}

type ifaceImplVar struct {
	implStructParamName string // name as used within interface's implementation struct // todo: get rid of?
	// t                   Type   // todo: type is inside field
	f field // the field as present in the function
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
	name := "iface_" + strings.Join(parts, "_")

	// Final cleanup: allow only letters, numbers, underscore
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	name = re.ReplaceAllString(name, "_")

	return name
}

func interfaceMethodFieldOrNil(ast *ast.InterfaceType, i int) ast.Expr {
	if ast == nil {
		return nil
	}

	return astTypeFieldFromFieldList(ast.Methods, i)
}
