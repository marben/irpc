package irpctestpkg

import (
	"errors"
	"fmt"
)

//go:generate go run ../

var divByZeroErr = errors.New("you cannot divide by zero")

type endpointApi interface {
	Div(a, b float64) (float64, error)
}

type endpointApiImpl struct {
}

func (api *endpointApiImpl) Div(a, b float64) (float64, error) {
	return 0, fmt.Errorf("not implemented")
}
