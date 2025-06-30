package irpctestpkg

type emptyAPI interface{}

//go:generate go run ../
type edgeCases interface {
	noReturn(i int)
	noParams() int
	nothingAtAll()
	// unnamedIntParam(int) // todo: uncomment and test!
}

var _ edgeCases = edgeCasesImpl{}

type edgeCasesImpl struct {
	noReturnFunc        func(int)
	noParamsFunc        func() int
	nothingAtAllFunc    func()
	unnamedIntParamFunc func(int)
}

// unnameIntParam implements edgeCases.
func (e edgeCasesImpl) unnamedIntParam(i int) {
	if e.unnamedIntParamFunc != nil {
		e.unnamedIntParamFunc(i)
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
