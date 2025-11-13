package irpctestpkg

import (
	"strings"
	"unicode"
)

//go:generate go run ../

// basicAPI tests types.Basic types
// and this is a two line description
type basicAPI interface {
	addByte(a byte, b byte) byte
	addInt(a int, b int) int
	swapInt(a int, b int) (int, int) // multiple return values
	subUint(a, b uint) uint
	addInt8(a, b int8) int8
	addUint8(a, b uint8) uint8
	addInt16(a, b int16) int16
	addUint16(a, b uint16) uint16
	addInt32(a, b int32) int32
	addUint32(a, b uint32) uint32
	addInt64(a, b int64) int64
	addUint64(a, b uint64) uint64
	addFloat64(a, b float64) float64
	addFloat32(a, b float32) float32
	toUpper(c rune) rune
	toUpperString(s string) string
	negBool(ok bool) bool
}

var _ basicAPI = basicApiImpl{}

type basicApiImpl struct {
	skew int
}

// negBool implements basicAPI.
func (basicApiImpl) negBool(ok bool) bool {
	return !ok
}

func (i basicApiImpl) addByte(a byte, b byte) byte {
	return a + b + byte(i.skew)
}

func (i basicApiImpl) addInt(a int, b int) int {
	return a + b + i.skew
}

func (i basicApiImpl) swapInt(a int, b int) (int, int) {
	return b + i.skew, a + i.skew
}

func (i basicApiImpl) subUint(a, b uint) uint {
	return a - b + uint(i.skew)
}

func (i basicApiImpl) addUint8(a, b uint8) uint8 {
	return a + b + uint8(i.skew)
}

func (i basicApiImpl) addInt8(a, b int8) int8 {
	return a + b + int8(i.skew)
}

func (i basicApiImpl) addInt16(a, b int16) int16 {
	return a + b + int16(i.skew)
}

func (i basicApiImpl) addUint16(a, b uint16) uint16 {
	return a + b + uint16(i.skew)
}

func (i basicApiImpl) addInt32(a, b int32) int32 {
	return a + b + int32(i.skew)
}

func (i basicApiImpl) addUint32(a, b uint32) uint32 {
	return a + b + uint32(i.skew)
}

func (i basicApiImpl) addInt64(a, b int64) int64 {
	return a + b + int64(i.skew)
}

func (i basicApiImpl) addUint64(a, b uint64) uint64 {
	return a + b + uint64(i.skew)
}

func (i basicApiImpl) addFloat32(a, b float32) float32 {
	return a + b + float32(i.skew)
}

func (i basicApiImpl) addFloat64(a, b float64) float64 {
	return a + b + float64(i.skew)
}

func (i basicApiImpl) toUpper(c rune) rune {
	return unicode.ToUpper(c)
}

func (i basicApiImpl) toUpperString(s string) string {
	return strings.ToUpper(s)
}
