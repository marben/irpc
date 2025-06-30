package irpctestpkg

type emptyAPI interface{}

//go:generate go run ../
type edgeCases interface {
	noReturn(i int)
	naParams() int
	nothingAtAll()
}

var _ edgeCases = edgeCasesImpl{}

type edgeCasesImpl struct {
	noReturnFunc     func(int)
	noParamsFunc     func() int
	nothingAtAllFunc func()
}

func (e edgeCasesImpl) noReturn(i int) {
	if e.noReturnFunc != nil {
		e.noReturnFunc(i)
	}
}

func (e edgeCasesImpl) naParams() int {
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
