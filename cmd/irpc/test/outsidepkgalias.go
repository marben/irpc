package irpctestpkg

import (
	"github.com/marben/irpc/cmd/irpc/test/out"
	out1 "github.com/marben/irpc/cmd/irpc/test/out"
	out2 "github.com/marben/irpc/cmd/irpc/test/out"
)

//go:generate go run ../
type outsidepkgalias interface {
	add(a out1.Uint8, b out1.Uint8) int
	add2(a out1.Uint8, b out.Uint8) int
	add3(a int, b out.Uint8) out2.Uint8
	sum(inSlice out1.AliasedByteSlice) int
}

type outsidepkgaliasImpl struct{}

// sum implements outsidepkgalias.
func (o *outsidepkgaliasImpl) sum(inSlice out2.AliasedByteSlice) int {
	var sum int
	for _, v := range inSlice {
		sum += int(v)
	}
	return sum
}

// add3 implements outsidepkgalias.
func (o *outsidepkgaliasImpl) add3(a int, b out2.Uint8) out2.Uint8 {
	return out2.Uint8(a + int(b))
}

// add implements outsidepkgalias.
func (o *outsidepkgaliasImpl) add(a out1.Uint8, b out1.Uint8) int {
	return int(a + b)
}

// add2 implements outsidepkgalias.
func (o *outsidepkgaliasImpl) add2(a out1.Uint8, b out1.Uint8) int {
	return int(a + b)
}

var _ outsidepkgalias = &outsidepkgaliasImpl{}
