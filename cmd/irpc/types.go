package main

import "go/types"

type typeDesc struct {
	TT        types.Type
	Qualifier string
	TypeName  string
}

func newTypeDesc(t types.Type, qualifier string) (typeDesc, error) {
	// switch t := t.Underlying().(type) {
	// case *types.Basic:
	// 	switch t.Kind() {
	// 	case types.Uint8:
	// 		return qualifiedUint8Encoder{q: qualifier, tt: t}, nil
	// 	}
	// }
	qf := func(pkg *types.Package) string {
		return qualifier
	}

	return typeDesc{TT: t, Qualifier: qualifier, TypeName: types.TypeString(t, qf)}, nil
}
