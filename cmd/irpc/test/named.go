package irpctestpkg

import (
	"slices"
)

type weekDay uint8

type weekDay2 weekDay

const (
	Monday weekDay = iota
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
	Sunday
)

const (
	Monday2 weekDay2 = iota
	Tuesday2
	Wednesday2
	Thursday2
	Friday2
	Saturday2
	Sunday2
)

type namedWeekDaysSliceType []weekDay2
type namedByteSliceType []byte
type namedMap map[int]float64

//go:generate go run ../
type namedTest interface {
	isWeekend(wd weekDay) bool
	isWeekend2(wd weekDay2) bool

	containsSaturday(wds []weekDay) bool // should use byteSliceEncoder
	containsSaturday2(wds namedWeekDaysSliceType) bool

	namedBytesSum(nb namedByteSliceType) int
	namedMapSum(namedMap) float64
}

var _ namedTest = namedTestImpl{}

type namedTestImpl struct {
}

// namedMapSum implements namedTest.
func (e namedTestImpl) namedMapSum(nm namedMap) float64 {
	var sum float64
	for i, f := range nm {
		sum += float64(i)
		sum += f
	}
	return sum
}

// namedBytesSum implements enumTest.
func (e namedTestImpl) namedBytesSum(nb namedByteSliceType) int {
	var sum int
	for _, b := range nb {
		sum += int(b)
	}

	return sum
}

func (e namedTestImpl) isWeekend(wd weekDay) bool {
	return wd == Saturday || wd == Sunday
}

func (e namedTestImpl) isWeekend2(wd weekDay2) bool {
	return wd == Saturday2 || wd == Sunday2
}

func (e namedTestImpl) containsSaturday(wds []weekDay) bool {
	return slices.Contains(wds, Saturday)
}

func (e namedTestImpl) containsSaturday2(wds namedWeekDaysSliceType) bool {
	return slices.Contains(wds, Saturday2)
}
