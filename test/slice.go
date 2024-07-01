package irpctestpkg

//go:generate go run ../

type sliceTest interface {
	SliceSum(slice []int) int
	VectMult(vect []int, s int) []int
	SliceOfFloat64Sum(slice []float64) float64
	SliceOfSlicesSum(slice [][]int) int
}

type sliceTestImpl struct {
	skew int
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
