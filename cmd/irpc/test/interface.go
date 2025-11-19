package irpctestpkg

import (
	"errors"

	"github.com/marben/irpc/cmd/irpc/test/out"
	out2 "github.com/marben/irpc/cmd/irpc/test/out"
)

//go:generate go run ../
type interfaceTest interface {
	rtnErrorWithMessage(msg string) error
	rtnNilError() error
	rtnTwoErrors() (error, error)
	rtnStringAndError(msg string) (s string, err error)
	passCustomInterfaceAndReturnItModified(ci customInterface) (customInterface, error)
	passJustCustomInterfaceWithoutError(ci customInterface) customInterface // todo: write a test

	// anonymous interface: // todo: write a test
	passAnonInterface(input interface {
		Name() string
		Age() int
	}) string

	passAnonInterfaceWithNamedParams(input interface {
		a() (out.Uint8, int)
		b() out2.Uint8
		c() (out2.Uint8, error)
	}) string
}

var _ interfaceTest = interfaceTestImpl{}

type interfaceTestImpl struct {
}

// passAnonInterfaceWithNamedParams implements interfaceTest.
func (i interfaceTestImpl) passAnonInterfaceWithNamedParams(input interface {
	a() (out.Uint8, int)
	b() out2.Uint8
	c() (out2.Uint8, error)
}) string {
	panic("unimplemented")
}

// passAnonInterface implements interfaceTest.
func (i interfaceTestImpl) passAnonInterface(input interface {
	Age() int
	Name() string
}) string {
	panic("unimplemented")
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

// passJustCustomInterfaceWithoutError implements interfaceTest.
func (i interfaceTestImpl) passJustCustomInterfaceWithoutError(ci customInterface) customInterface {
	return ci
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
