package irpctestpkg

import "github.com/marben/irpc/cmd/irpc/test/out"

// import ous "github.com/marben/irpc/cmd/irpc/test/out"

//go:generate go run ../
type outsideTest interface {
	addUint8(a, b out.Uint8) out.Uint8
}

var _ outsideTest = outsideTestImpl{}

type outsideTestImpl struct {
}

// addUint8 implements outsideTest.
func (o outsideTestImpl) addUint8(a out.Uint8, b out.Uint8) out.Uint8 {
	return a + b
}
