package irpctestpkg

import (
	"slices"
	"time"
)

//go:generate go run ../

type namedByteSlice []byte
type namedBoolSlice []bool

type sliceTest interface {
	SliceSum(slice []int) int
	VectMult(vect []int, s int) []int
	SliceOfFloat64Sum(slice []float64) float64
	SliceOfSlicesSum(slice [][]int) int
	SliceOfBytesSum(slice []byte) int
	namedByteSlice(slice namedByteSlice) int
	sliceOfBools(slice []bool) namedBoolSlice
	sliceOfUint8(slice []uint8) []byte
	sliceOfMaps(slice []map[int]string)
	sliceOfStructs(slice []struct {
		A int
		B string
	}) (sumA int)
	sliceOfErrors(slice []error)
	isNilSlice(s []string) bool
	isNilBoolSlice(bs []bool) bool
	isNilByteSlice(bs []byte) bool
	sliceOfTimesReverse(in []time.Time) []time.Time
}

var _ sliceTest = sliceTestImpl{}

type sliceTestImpl struct {
	skew int
}

// sliceOfTimesReverse implements sliceTest.
func (st sliceTestImpl) sliceOfTimesReverse(in []time.Time) []time.Time {
	slices.Reverse(in)
	return in
}

// isNilBoolSlice implements sliceTest.
func (st sliceTestImpl) isNilBoolSlice(bs []bool) bool {
	return bs == nil
}

// isNilByteSlice implements sliceTest.
func (st sliceTestImpl) isNilByteSlice(bs []byte) bool {
	return bs == nil
}

// isNilSlice implements sliceTest.
func (st sliceTestImpl) isNilSlice(s []string) bool {
	return s == nil
}

// sliceOfStructs implements sliceTest.
func (st sliceTestImpl) sliceOfStructs(slice []struct {
	A int
	B string
}) (sumA int) {
	for _, s := range slice {
		sumA += s.A
	}
	return sumA
}

// sliceOfMaps implements sliceTest.
func (st sliceTestImpl) sliceOfMaps(slice []map[int]string) {
	panic("unimplemented")
}

// sliceOfErrors implements sliceTest.
func (st sliceTestImpl) sliceOfErrors(slice []error) {
	panic("unimplemented")
}

// sliceOfUint8 implements sliceTest.
func (st sliceTestImpl) sliceOfUint8(slice []uint8) []byte {
	slices.Reverse(slice)
	return slice
}

// sliceOfBools implements sliceTest.
func (st sliceTestImpl) sliceOfBools(slice []bool) namedBoolSlice {
	panic("unimplemented")
}

func (st sliceTestImpl) namedByteSlice(slice namedByteSlice) int {
	panic("unimplemented")
}

func (st sliceTestImpl) SliceSum(slice []int) int {
	var s int
	for _, v := range slice {
		s += v
	}
	return s + st.skew
}

func (st sliceTestImpl) VectMult(vect []int, s int) []int {
	rtn := make([]int, len(vect))
	for i, v := range vect {
		rtn[i] = v*s + st.skew
	}
	return rtn
}

func (st sliceTestImpl) SliceOfFloat64Sum(slice []float64) float64 {
	var sum float64
	for _, v := range slice {
		sum += v
	}
	return sum + float64(st.skew)
}

func (st sliceTestImpl) SliceOfSlicesSum(slice [][]int) int {
	var sum int
	for _, s := range slice {
		for _, v := range s {
			sum += v
		}
	}
	return sum
}
func (st sliceTestImpl) SliceOfBytesSum(slice []byte) int {
	var res int
	for _, v := range slice {
		res += int(v)
	}
	return res
}
