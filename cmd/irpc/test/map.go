package irpctestpkg

//go:generate go run ../

type mapTest interface {
	mapSum(in map[int]float64) (keysSum int, valsSum float64)
	sumStructs(in map[intStruct]intStruct) (keysSum, valsSum int)
	sumSlices(in map[intStruct][]intStruct) (keysSum, valsSum int)
}

type mapTestImpl struct {
}

func (mt mapTestImpl) mapSum(in map[int]float64) (keysSum int, valsSum float64) {
	for k, v := range in {
		keysSum += k
		valsSum += v
	}
	return keysSum, valsSum
}

func (mt *mapTestImpl) sumStructs(in map[intStruct]intStruct) (keysSum, valsSum int) {
	var keySum, valSum int
	for k, v := range in {
		keySum += k.i + k.j + k.k + k.l
		valSum += v.i + v.j + v.k + v.l
	}
	return keySum, valSum
}

func (mt *mapTestImpl) sumSlices(in map[intStruct][]intStruct) (keysSum, valsSum int) {
	var keySum, valSum int
	for k, v := range in {
		keySum += k.i + k.j + k.k + k.l
		for _, v2 := range v {
			valSum += v2.i + v2.j + v2.k + v2.l
		}
	}
	return keySum, valSum
}

type intStruct struct {
	i, j, k, l int
}
