package irpctestpkg

//go:generate go run ../

import (
	"maps"
)

type namedIntFloatMap map[int]float64
type mapNamedInt int
type mapNamedFloat64 float64

type mapTest interface {
	mapSum(in map[int]float64) (keysSum int, valsSum float64)
	sumStructs(in map[intStruct]intStruct) (keysSum, valsSum int)
	sumSlices(in map[intStruct][]intStruct) (keysSum, valsSum int)
	namedMapInc(in namedIntFloatMap) namedIntFloatMap               // todo: write a test
	namedKeySum(in map[mapNamedInt]mapNamedFloat64) mapNamedFloat64 // todo: write a test
	emptyInterfaceMapReflect(in map[int]interface{}) map[int]interface{}
}

var _ mapTest = mapTestImpl{}

type mapTestImpl struct {
}

// emptyInterfaceMapSum implements mapTest.
func (mt mapTestImpl) emptyInterfaceMapReflect(in map[int]interface{}) map[int]interface{} {
	// log.Printf("implementation obtained map: %v", in)
	var rtnMap = make(map[int]interface{}, len(in))
	maps.Copy(rtnMap, in)
	return rtnMap
}

func (mt mapTestImpl) mapSum(in map[int]float64) (keysSum int, valsSum float64) {
	for k, v := range in {
		keysSum += k
		valsSum += v
	}
	return keysSum, valsSum
}

func (mt mapTestImpl) sumStructs(in map[intStruct]intStruct) (keysSum, valsSum int) {
	var keySum, valSum int
	for k, v := range in {
		keySum += k.i + k.j + k.k + k.l
		valSum += v.i + v.j + v.k + v.l
	}
	return keySum, valSum
}

func (mt mapTestImpl) sumSlices(in map[intStruct][]intStruct) (keysSum, valsSum int) {
	var keySum, valSum int
	for k, v := range in {
		keySum += k.i + k.j + k.k + k.l
		for _, v2 := range v {
			valSum += v2.i + v2.j + v2.k + v2.l
		}
	}
	return keySum, valSum
}

// namedKeySum implements mapTest.
func (mt mapTestImpl) namedKeySum(in map[mapNamedInt]mapNamedFloat64) mapNamedFloat64 {
	panic("unimplemented")
}

// namedMapInc implements mapTest.
func (mt mapTestImpl) namedMapInc(in namedIntFloatMap) namedIntFloatMap {
	panic("unimplemented")
}

type intStruct struct {
	i, j, k, l int
}
