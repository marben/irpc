package irpc

import (
	"context"
	"errors"
	"fmt"
)

// this file contains a testing service
// it is here, because importing one from testtols
// causes cyclic dependency and i cannot be bothered to implement
// a special (package irpc) case for generator atm

// the interface that is base for our rpc
type testServiceIface interface {
	Add(a, b int) int
}

// the implementation of our function
type testServiceImpl struct {
	skew       int // skew is added to result, to distinguish different versions of math
	addFunc    func(int, int) int
	divErrFunc func(int, int) (int, error)
}

func (mh testServiceImpl) Add(a, b int) int {
	if mh.addFunc == nil {
		return a + b + mh.skew
	}
	return mh.addFunc(a, b)
}

func (mh testServiceImpl) DivErr(a int, b int) (int, error) {
	if mh.divErrFunc == nil {
		if b == 0 {
			return 0, errors.New("cannot divide by zero")
		}
		return a/b + mh.skew, nil
	}
	return mh.divErrFunc(a, b)
}

var _ testServiceIface = testServiceImpl{}
var _ Service = &testIRpcService{}

type testIRpcService struct {
	impl testServiceIface
}

func newTestIRpcService(impl testServiceIface) *testIRpcService { return &testIRpcService{impl: impl} }

func (ms *testIRpcService) GetFuncCall(funcId FuncId) (ArgDeserializer, error) {
	switch funcId {
	case mathIrpcFuncAddId:
		return func(d *Decoder) (FuncExecutor, error) {
			// DESERIALIZE
			var args addParams
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func(ctx context.Context) Serializable {
				// EXECUTE
				var resp addRtnVals
				resp.Res = ms.impl.Add(args.A, args.B)
				return resp
			}, nil
		}, nil
	default:
		return nil, fmt.Errorf("function '%v' doesn't exist on service '%s'", funcId, ms.Hash())
	}
}

var mathIrpcServiceHash = []byte("MathServiceHash")

func (*testIRpcService) Hash() []byte {
	return mathIrpcServiceHash
}

var _ testServiceIface = &MathIRpcClient{}

const (
	mathIrpcFuncAddId FuncId = iota
)

type MathIRpcClient struct {
	ep *Endpoint
	id RegisteredServiceId
}

func NewMathIrpcClient(ep *Endpoint) (*MathIRpcClient, error) {
	id, err := ep.RegisterClient(context.Background(), mathIrpcServiceHash)
	if err != nil {
		return nil, fmt.Errorf("register failed: %w", err)
	}

	return &MathIRpcClient{ep, id}, nil
}

// todo: maybe we request error return from rpc functions?
func (mc *MathIRpcClient) Add(a int, b int) int {
	var params = addParams{A: a, B: b}
	var resp addRtnVals

	if err := mc.ep.CallRemoteFunc(context.Background(), mc.id, mathIrpcFuncAddId, params, &resp); err != nil {
		panic(fmt.Sprintf("callRemoteFunc failed: %v", err))
	}

	return resp.Res
}

type addParams struct {
	A int
	B int
}

func (p addParams) Serialize(e *Encoder) error {
	if err := e.Int(p.A); err != nil {
		return err
	}
	if err := e.Int(p.B); err != nil {
		return err
	}
	return nil
}

func (p *addParams) Deserialize(d *Decoder) error {
	if err := d.Int(&p.A); err != nil {
		return err
	}
	if err := d.Int(&p.B); err != nil {
		return err
	}
	return nil
}

type addRtnVals struct {
	Res int
}

func (v addRtnVals) Serialize(e *Encoder) error {
	return e.Int(v.Res)
}

func (v *addRtnVals) Deserialize(d *Decoder) error {
	return d.Int(&v.Res)
}
