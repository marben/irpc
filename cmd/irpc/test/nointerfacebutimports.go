package irpctestpkg

import (
	"time"

	"github.com/marben/irpc/cmd/irpc/test/out"
)

// this file tests, whether imports are properly ommited

//go:generate go run ../

func timeStringer(t time.Time) string {
	return t.String()
}

func outImport(u8 out.Uint8) uint8 {
	return uint8(u8)
}
