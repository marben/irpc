package irpctestpkg

import "github.com/marben/irpc/cmd/irpc/test/out"

type namedSliceOfInts []int

//go:generate go run ../
type sliceNamedApi interface {
	sumNamedInts(vec namedSliceOfInts) int
	sumOutsideNamedInts(vec out.AliasedByteSlice) int
}
