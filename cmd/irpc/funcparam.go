package main

import "fmt"

// represents a variable in param struct which in turn represents function parameter/return value
type funcParam struct {
	name            string // original name as defined in the interface. can be ""
	identifier      string // identifier we use for this field. it's either param.name or if there is none, we generate it
	structFieldName string
	typ             Type
}

// requestParamNames contains all parameter names, including ours
// if our parameter doesn't have a name, we will create one, making suere, we don't overlap with named parameters
func newRequestParam(p rpcParam, requestParamNames map[string]struct{}) (funcParam, error) {
	// figure out a unique id
	id := p.name
	if id == "" || id == "_" {
		id = fmt.Sprintf("p%d", p.pos)
		for {
			if _, exists := requestParamNames[id]; exists {
				id += "_"
			} else {
				break
			}
		}
	}
	requestParamNames[id] = struct{}{}

	return funcParam{
		name:            p.name,
		identifier:      id,
		structFieldName: fmt.Sprintf("Param%d_%s", p.pos, id),
		typ:             p.typ,
	}, nil
}

func newResultParam(p rpcParam) (funcParam, error) {
	sFieldName := fmt.Sprintf("Param%d", p.pos)
	if p.name != "" {
		sFieldName += "_" + p.name
	}

	return funcParam{
		name:            p.name,
		identifier:      p.name,
		structFieldName: sFieldName,
		typ:             p.typ,
	}, nil
}

// returns true if field is of type context.Context
func (vf funcParam) isContext() bool {
	if _, ok := vf.typ.(contextType); ok {
		return true
	}
	return false
}
