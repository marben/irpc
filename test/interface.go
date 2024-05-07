package irpctestpkg

import (
	"errors"
)

//go:generate go run ../
type interfaceTest interface {
	rtnErrorWithMessage(msg string) error
	rtnNilError() error
	rtnTwoErrors() (error, error)
	rtnStringAndError(msg string) (s string, err error)
	passCustomInterfaceAndReturnItModified(ci customInterface) (customInterface, error)
}

type interfaceTestImpl struct {
}

func (interfaceTestImpl) rtnErrorWithMessage(msg string) error {
	return errors.New(msg)
}

func (interfaceTestImpl) rtnNilError() error {
	return nil
}

func (interfaceTestImpl) rtnStringAndError(msg string) (s string, err error) {
	return msg, errors.New(msg)
}

// first error is nil
func (interfaceTestImpl) rtnTwoErrors() (error, error) {
	return nil, errors.New("err2")
}

func (interfaceTestImpl) passCustomInterfaceAndReturnItModified(ci customInterface) (customInterface, error) {
	if ci == nil {
		return nil, errors.New("nil pointer")
	}
	impl := customInterfaceImpl{
		i: ci.IntFunc() + 1,
		s: ci.StringFunc() + "_modified",
	}
	return impl, nil
}

// todo: currently we are also generating service and client for this interface. perhaps we don't want that?
type customInterface interface {
	IntFunc() int
	StringFunc() string
}

type customInterfaceImpl struct {
	i int
	s string
}

func (i customInterfaceImpl) IntFunc() int {
	return i.i
}

func (i customInterfaceImpl) StringFunc() string {
	return i.s
}
