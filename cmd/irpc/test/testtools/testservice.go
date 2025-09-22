package testtools

import (
	"context"
	"errors"
)

//go:generate go run ../../
type TestService interface {
	Div(a, b int) int
	DivErr(a, b int) (int, error)
	DivCtxErr(ctx context.Context, a, b int) (int, error)
}

func TestServiceId() []byte {
	impl := NewTestServiceImpl(0)
	service := NewTestServiceIRpcService(impl)
	return service.id
}

var _ TestService = NewTestServiceImpl(0)

var ErrDivByZero = errors.New("cannot divide by zero")

// functions can be redefined as needed
type TestServiceImpl struct {
	Skew          int
	DivFunc       func(a, b int) int
	DivErrFunc    func(a, b int) (int, error) // Errors on division by zero
	DivCtxErrFunc func(ctx context.Context, a int, b int) (int, error)
}

func NewTestServiceImpl(skew int) *TestServiceImpl {
	s := &TestServiceImpl{
		Skew: skew,
	}
	s.DivFunc = func(a, b int) int { return a/b + s.Skew }
	s.DivErrFunc = func(a, b int) (int, error) {
		if b == 0 {
			return 0, ErrDivByZero
		}
		return a/b + s.Skew, nil
	}
	s.DivCtxErrFunc = func(ctx context.Context, a, b int) (int, error) {
		// todo: check if context ended
		if b == 0 {
			return 0, ErrDivByZero
		}
		return a/b + s.Skew, nil
	}

	return s
}

// Div implements TestService.
func (t *TestServiceImpl) Div(a int, b int) int {
	return t.DivFunc(a, b)
}

// DivErr implements TestService.
func (t *TestServiceImpl) DivErr(a int, b int) (int, error) {
	return t.DivErrFunc(a, b)
}

// DivCtxErr implements TestService.
func (t *TestServiceImpl) DivCtxErr(ctx context.Context, a int, b int) (int, error) {
	return t.DivCtxErrFunc(ctx, a, b)
}
