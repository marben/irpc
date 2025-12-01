package irpctestpkg

import (
	"time"

	"github.com/marben/irpc/cmd/irpc/test/out"
)

type namedSliceOfInts []int
type namedSliceOfTime []time.Time

//go:generate go run ../
type sliceNamedApi interface {
	sumNamedInts(vec namedSliceOfInts) int
	sumOutsideNamedInts(vec out.AliasedByteSlice) int
	sumSliceOfNamedInts(vec []out.Uint8) out.Uint8
	reverseNamedSliceOfTime(in namedSliceOfTime) namedSliceOfTime
}
