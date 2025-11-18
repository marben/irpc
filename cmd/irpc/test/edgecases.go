package irpctestpkg

type emptyAPI interface{}

//go:generate go run ../
type edgeCases interface {
	noReturn(i int)
	noParams() int
	nothingAtAll()
	unnamedIntParam(int, int)
	mixedParamIds(_ int, p0 uint8, _ struct{ a int })
	underscoreParamNames(_ int, p0 uint8, _ float64) (_ float64)
	underscoreRtnName(p0 int) (_ int, _ uint8)
}

type anotherInterface interface {
	anotherAdd(a, b int) int
}

var _ edgeCases = edgeCasesImpl{}

type edgeCasesImpl struct {
	noReturnFunc        func(int)
	noParamsFunc        func() int
	nothingAtAllFunc    func()
	unnamedIntParamFunc func(int, int)
	mixedParamIdsFunc   func(int)
}

// underscoreRtnName implements edgeCases.
func (e edgeCasesImpl) underscoreRtnName(p0 int) (_ int, _ uint8) {
	panic("unimplemented")
}

// underscoreParamNames implements edgeCases.
func (e edgeCasesImpl) underscoreParamNames(_ int, par uint8, _ float64) (_ float64) {
	panic("unimplemented")
}

// mixedParamIds implements edgeCases.
func (e edgeCasesImpl) mixedParamIds(a int, b uint8, c struct{ a int }) {
	if e.mixedParamIdsFunc != nil {
		e.mixedParamIdsFunc(a + int(b) + c.a)
	}
}

// unnamedIntParam implements edgeCases.
func (e edgeCasesImpl) unnamedIntParam(i, j int) {
	if e.unnamedIntParamFunc != nil {
		e.unnamedIntParamFunc(i, j)
	}
}

func (e edgeCasesImpl) noReturn(i int) {
	if e.noReturnFunc != nil {
		e.noReturnFunc(i)
	}
}

func (e edgeCasesImpl) noParams() int {
	if e.noParamsFunc != nil {
		return e.noParamsFunc()
	}
	return 7
}

func (e edgeCasesImpl) nothingAtAll() {
	if e.nothingAtAllFunc != nil {
		e.nothingAtAllFunc()
	}
}
