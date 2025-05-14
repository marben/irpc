package irpctestpkg

//go:generate go run ../
type Math interface {
	Add(a, b int) (int, error)
}

var _ Math = MathImpl{0}

type MathImpl struct {
	Skew int
}

// Add implements Math.
func (m MathImpl) Add(a int, b int) (int, error) {
	return a + b + m.Skew, nil
}
