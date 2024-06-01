package irpctestpkg

import (
	"errors"
)

//go:generate go run ../

type tcpTestApi interface {
	Div(a, b float64) (float64, error)
}

type tcpTestApiImpl struct {
}

var errDiv = errors.New("cannot divide by zero")

func (i tcpTestApiImpl) Div(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errDiv
	}

	return a / b, nil
}
