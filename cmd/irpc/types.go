package main

import (
	"go/types"
)

type typeDesc struct {
	tt                types.Type
	qualifiedTypeName string
	enc               encoder // nil if doesn't exist. todo: get rid of nils, once we can move all encoders here
}
