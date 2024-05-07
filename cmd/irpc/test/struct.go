package irpctestpkg

//go:generate go run ../

type vect3 struct {
	a, b, c int
}

type vect3x3 struct {
	v1, v2, v3 vect3
}

type sliceStruct struct {
	s1 []int
	s2 []int
}

type structAPI interface {
	VectSum(v vect3) int
	Vect3x3Sum(v vect3x3) vect3
	SumSliceStruct(s sliceStruct) int
}

var _ structAPI = structImpl{}

type structImpl struct {
	skew int
}

// Vect3x3Sum implements structAPI.
func (i structImpl) Vect3x3Sum(v vect3x3) vect3 {
	return vect3{
		a: v.v1.a + v.v2.a + v.v3.a + i.skew,
		b: v.v1.b + v.v2.b + v.v3.b + i.skew,
		c: v.v1.c + v.v2.c + v.v3.c + i.skew,
	}
}

func (i structImpl) VectSum(v vect3) int {
	return v.a + v.b + v.c + i.skew
}

// sumSliceStruct implements structAPI.
func (i structImpl) SumSliceStruct(s sliceStruct) int {
	var sum int
	for _, v := range append(s.s1, s.s2...) {
		sum += v
	}
	return sum + i.skew
}
