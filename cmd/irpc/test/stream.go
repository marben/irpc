package irpctestpkg

import (
	gocontext "context"
	"errors"
	"strings"
)

//go:generate go run ../

type streamPoint struct {
	X, Y int
}

type streamTestApi interface {
	// a normal method, next to methods with streams
	Add(a, b int) (int, error)

	// unnamed parameters, channel parameter only
	Sum(gocontext.Context, <-chan int) (int, error)

	// scalar parameter, channel parameter, scalar result and two returned channels
	Split(ctx gocontext.Context, sep string, lines <-chan string) (sepLen int, words <-chan string, lens <-chan int, err error)

	// returned channel of structs only, the implementation may fail
	Points(ctx gocontext.Context, n int) (<-chan streamPoint, error)
}

type streamTestImpl struct{}

func (streamTestImpl) Add(a, b int) (int, error) {
	return a + b, nil
}

func (streamTestImpl) Sum(ctx gocontext.Context, in <-chan int) (int, error) {
	sum := 0
	for {
		select {
		case v, ok := <-in:
			if !ok {
				return sum, nil
			}
			sum += v
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}
}

func (streamTestImpl) Split(ctx gocontext.Context, sep string, lines <-chan string) (int, <-chan string, <-chan int, error) {
	words := make(chan string)
	lens := make(chan int)
	go func() {
		defer close(words)
		defer close(lens)
		for line := range lines {
			for _, w := range strings.Split(line, sep) {
				select {
				case words <- w:
				case <-ctx.Done():
					return
				}
			}
			select {
			case lens <- len(line):
			case <-ctx.Done():
				return
			}
		}
	}()
	return len(sep), words, lens, nil
}

var errNegativePoints = errors.New("n must be >= 0")

func (streamTestImpl) Points(ctx gocontext.Context, n int) (<-chan streamPoint, error) {
	if n < 0 {
		return nil, errNegativePoints
	}
	out := make(chan streamPoint)
	go func() {
		defer close(out)
		for i := range n {
			select {
			case out <- streamPoint{X: i, Y: -i}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, nil
}
