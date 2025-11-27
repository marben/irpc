package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"regexp"
	"strings"
)

var _ Type = interfaceType{}

type interfaceType struct {
	fncs         []ifaceFunc
	implTypeName string
	ni           *namedInfo
	apiName      string
	t            *types.Interface
	boolT        Type
}

func (tr *typeResolver) newInterfaceType(apiName string, ni *namedInfo, ifaceT *types.Interface, astExpr ast.Expr) (interfaceType, error) {
	var ifaceAst *ast.InterfaceType
	if astExpr != nil {
		var ok bool
		ifaceAst, ok = astExpr.(*ast.InterfaceType)
		if !ok {
			return interfaceType{}, fmt.Errorf("provided ast is not *ast.InterfaceType, but %T", astExpr)
		}
	}

	fncs := []ifaceFunc{}
	for i := 0; i < ifaceT.NumMethods(); i++ {
		method := ifaceT.Method(i)
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

		results := make([]ifaceImplVar, 0, sig.Results().Len())
		for i := 0; i < sig.Results().Len(); i++ {
			v := sig.Results().At(i)
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
				f:                   newField(v.Name(), typ),
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
		n = sanitizeInterfaceName(ifaceT)
	}
	implTypeName := "_" + n + "_" + apiName + "_impl"

	return interfaceType{
		implTypeName: implTypeName,
		fncs:         fncs,
		ni:           ni,
		apiName:      apiName,
		t:            ifaceT,
		boolT:        tr.boolType,
	}, nil
}

// name implements Type.
func (i interfaceType) name(q *qualifier) string {
	if i.ni != nil {
		return i.ni.qualifiedName(q)
	}

	sb := strings.Builder{}
	sb.WriteString("interface{")
	for _, f := range i.fncs {
		sb.WriteString(f.signature(q))
		sb.WriteString(";")
	}
	sb.WriteString("}")
	return sb.String()
}

// genEncFunc implements Type.
func (i interfaceType) genEncFunc(q *qualifier) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "func (enc *irpcgen.Encoder, v %s) error {\n", i.name(q))
	fmt.Fprintf(sb, `isNil := v == nil
			if err := irpcgen.EncBool(enc, isNil); err != nil {
				return fmt.Errorf("serialize isNil == %%t: %%w", isNil, err)
			}
			if isNil {
				return nil
			}
		`)
	for _, fn := range i.fncs {
		for i, r := range fn.results {
			sb.WriteString(r.implStructParamName)
			if i != len(fn.results)-1 {
				sb.WriteString(",")
			}
		}
		fmt.Fprintf(sb, ":= v.%s()\n", fn.name)
		for _, r := range fn.results {
			fmt.Fprintf(sb, `if err := %s(enc, %s); err != nil {
				return fmt.Errorf("serialize \"v.%s()\" of type %s: %%w", err)
			}
				`, r.f.t.genEncFunc(q), r.implStructParamName, fn.name, r.f.t.name(q.copy()))
		}
	}
	sb.WriteString("")
	sb.WriteString("return nil\n")
	sb.WriteString("}")

	return sb.String()
}

// genDecFunc implements Type.
func (i interfaceType) genDecFunc(q *qualifier) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "func (dec *irpcgen.Decoder, s *%s) error {\n", i.name(q))
	fmt.Fprintf(sb, `var isNil bool
		if err := irpcgen.DecBool(dec, &isNil); err != nil {
			return fmt.Errorf("deserialize isNil: %%w:", err)
		}
			if isNil {
				return nil
			}
		`)
	fmt.Fprintf(sb, "var impl %s\n", i.implTypeName)
	for _, fn := range i.fncs {
		for _, r := range fn.results {
			fmt.Fprintf(sb, `if err := %s(dec, &impl.%s); err != nil{
				return fmt.Errorf("deserialize \"%s\" %s: %%w", err)
			}
			`, r.f.t.genDecFunc(q), r.implStructParamName, r.implStructParamName, r.f.t.name(q.copy()))
		}
	}
	sb.WriteString("*s = impl\n")
	sb.WriteString("return nil\n")
	sb.WriteString("}")
	return sb.String()
}

// we need concrete implementation of our interface
func (i interfaceType) codeblocks(q *qualifier) []string {
	sb := &strings.Builder{}

	// struct that will hold all our interface's return values
	fmt.Fprintf(sb, "type %s struct {\n", i.implTypeName)
	for _, f := range i.fncs {
		for _, v := range f.results {
			fmt.Fprintf(sb, "%s %s\n", v.implStructParamName, v.f.t.name(q))
		}
	}
	sb.WriteString("}\n")

	// we implement each interface function and return value stored withing the struct
	for _, f := range i.fncs {
		fmt.Fprintf(sb, "func (i %s)%s()(%s){\n", i.implTypeName, f.name, f.listNamesWithTypes(q))
		fmt.Fprintf(sb, "return %s\n", f.listImplNamesPrefixed("i."))
		sb.WriteString("}\n")
	}

	return []string{sb.String()}
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
		fmt.Fprintf(b, "%s %s", v.f.name, v.f.t.name(q)) // todo: need qualified name?
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
	f                   field  // the field as present in the function
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
