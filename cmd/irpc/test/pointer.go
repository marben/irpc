package irpctestpkg

import (
	"image"
	"image/color"
)

//go:generate go run ../
type pointerTest interface {
	getImage() *image.RGBA
	getNilImage() *image.RGBA
	isNil(ptr *int) bool
}

var _ pointerTest = &pointerTestImpl{}

func newPointerTestImpl() *pointerTestImpl {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	img.SetRGBA(50, 50, color.RGBA{R: 10, G: 20, B: 30})
	return &pointerTestImpl{img: img}
}

type pointerTestImpl struct {
	img *image.RGBA
}

// getImage implements pointerTest.
func (p *pointerTestImpl) getImage() *image.RGBA {
	return p.img
}

// getNilImage implements pointerTest.
func (p *pointerTestImpl) getNilImage() *image.RGBA {
	return nil
}

// isNil implements pointerTest.
func (p *pointerTestImpl) isNil(ptr *int) bool {
	return ptr == nil
}
