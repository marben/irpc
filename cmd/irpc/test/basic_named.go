package irpctestpkg

import out2 "github.com/marben/irpc/cmd/irpc/test/out"

type FakeUint8 uint8

//go:generate go run ../
type basicNamedAPI interface {
	addFakeUint8(a, b out2.Uint8) FakeUint8
	addUint8(a, b uint8) uint8
	addByte(a, b byte) byte
}
