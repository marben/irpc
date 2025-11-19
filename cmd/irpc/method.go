package main

import (
	"fmt"
	"go/ast"
	"strings"
)

type methodGenerator struct {
	name      string
	index     int
	req, resp paramStructGenerator
	ctxVar    string // context used for method call (either there is context param, or we use context.Background() )
	goDoc     string
}

func newMethodGenerator(tr typeResolver, apiName string, methodField *ast.Field, index int) (methodGenerator, error) {
	if len(methodField.Names) == 0 {
		return methodGenerator{}, fmt.Errorf("method of interface %q has no name", apiName)
	}
	methodName := methodField.Names[0].Name

	astFuncType, ok := methodField.Type.(*ast.FuncType)
	if !ok {
		return methodGenerator{}, fmt.Errorf("*ast.Field %v is not *ast.FuncType", methodField)
	}

	params, err := tr.loadRpcParamList(apiName, astFuncType.Params.List)
	if err != nil {
		return methodGenerator{}, fmt.Errorf("params list load for %s: %w", methodName, err)
	}

	var results []rpcParam
	if astFuncType.Results != nil {
		results, err = tr.loadRpcParamList(apiName, astFuncType.Results.List)
		if err != nil {
			return methodGenerator{}, fmt.Errorf("results list load for %s: %w", methodName, err)
		}
	}

	req, resp, err := newReqRespStructsGenerator(apiName, methodName, params, results)
	if err != nil {
		return methodGenerator{}, fmt.Errorf("newReqRespStructsGenerator(): %w", err)
	}

	// context
	// we currently only support one or no context var
	// multiple ctx vars could be combined, but it doesn't make much sense and i cannot be bothered atm
	ctxParams := []genParam{}
	for _, p := range req.params {
		if p.isContext() {
			ctxParams = append(ctxParams, p)
		}
	}
	var ctxVarName string
	switch len(ctxParams) {
	case 0:
		ctxVarName = "context.Background()"
	case 1:
		ctxVarName = ctxParams[0].identifier
	default:
		return methodGenerator{}, fmt.Errorf("%s - %s : cannot have more than one context parameter", apiName, methodName)
	}

	return methodGenerator{
		name:   methodName,
		index:  index,
		req:    req,
		resp:   resp,
		ctxVar: ctxVarName,
		goDoc:  godocFromAstCommentGroup(methodField.Doc),
	}, nil
}

// creates method call list with each var prefixed with 'prefix'
// replaces any parameter of type context.Context with 'ctxVarName'
func (mg methodGenerator) requestParamsListPrefixed(prefix, ctxVarName string) string {
	sb := &strings.Builder{}
	for i, p := range mg.req.params {
		if p.isContext() {
			sb.WriteString(ctxVarName)
		} else {
			fmt.Fprintf(sb, "%s%s", prefix, p.structFieldName)
		}
		if i != len(mg.req.params) {
			sb.WriteString(",")
		}
	}
	return sb.String()
}

func (mg methodGenerator) executorFuncCode(q *qualifier) string {
	q.addUsedImport(contextImport)
	if mg.resp.isEmpty() {
		return fmt.Sprintf(`func(ctx context.Context) irpcgen.Serializable {
				// EXECUTE
				s.impl.%[2]s(%[3]s)
				return irpcgen.EmptySerializable{}
			}`, mg.resp.structName, mg.name, mg.requestParamsListPrefixed("args.", "ctx"))
	}

	return fmt.Sprintf(`func(ctx context.Context) irpcgen.Serializable {
				// EXECUTE
				var resp %[1]s
				%[2]s = s.impl.%[3]s(%[4]s)
				return resp
			}`, mg.resp.structName, mg.resp.paramListPrefixed("resp."), mg.name, mg.requestParamsListPrefixed("args.", "ctx"))
}
