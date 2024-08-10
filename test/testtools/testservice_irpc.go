// Code generated by irpc generator; DO NOT EDIT
package testtools

import (
	"fmt"
	"github.com/marben/irpc/pkg/irpc"
)

type TestServiceIRpcService struct {
	impl TestService
}

func NewTestServiceIRpcService(impl TestService) *TestServiceIRpcService {
	return &TestServiceIRpcService{impl: impl}
}
func (TestServiceIRpcService) Hash() []byte {
	return []byte("TestServiceIRpcService")
}
func (s *TestServiceIRpcService) GetFuncCall(funcId irpc.FuncId) (irpc.ArgDeserializer, error) {
	switch funcId {
	case 0: // Div
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_TestServiceDivReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_TestServiceDivResp
				resp.Param0_ = s.impl.Div(args.Param0_a, args.Param0_b)
				return resp
			}, nil
		}, nil
	case 1: // DivErr
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_TestServiceDivErrReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_TestServiceDivErrResp
				resp.Param0_, resp.Param1_ = s.impl.DivErr(args.Param0_a, args.Param0_b)
				return resp
			}, nil
		}, nil
	default:
		return nil, fmt.Errorf("function '%d' doesn't exist on service '%s'", funcId, string(s.Hash()))
	}
}

type TestServiceIRpcClient struct {
	endpoint *irpc.Endpoint
	id       irpc.RegisteredServiceId
}

func NewTestServiceIRpcClient(endpoint *irpc.Endpoint) (*TestServiceIRpcClient, error) {
	id, err := endpoint.RegisterClient([]byte("TestServiceIRpcService"))
	if err != nil {
		return nil, fmt.Errorf("register failed: %w", err)
	}
	return &TestServiceIRpcClient{endpoint: endpoint, id: id}, nil
}
func (_c *TestServiceIRpcClient) Div(a int, b int) int {
	var req = _Irpc_TestServiceDivReq{
		Param0_a: a,
		Param0_b: b,
	}
	var resp _Irpc_TestServiceDivResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 0, req, &resp); err != nil {
		panic(err) // to avoid panic, make your func return error and regenerate the code
	}
	return resp.Param0_
}
func (_c *TestServiceIRpcClient) DivErr(a int, b int) (int, error) {
	var req = _Irpc_TestServiceDivErrReq{
		Param0_a: a,
		Param0_b: b,
	}
	var resp _Irpc_TestServiceDivErrResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 1, req, &resp); err != nil {
		var zero _Irpc_TestServiceDivErrResp
		return zero.Param0_, err
	}
	return resp.Param0_, resp.Param1_
}

type _Irpc_TestServiceDivReq struct {
	Param0_a int
	Param0_b int
}

func (s _Irpc_TestServiceDivReq) Serialize(e *irpc.Encoder) error {
	if err := e.Int(s.Param0_a); err != nil {
		return fmt.Errorf("serialize s.Param0_a of type 'int': %w", err)
	}
	if err := e.Int(s.Param0_b); err != nil {
		return fmt.Errorf("serialize s.Param0_b of type 'int': %w", err)
	}
	return nil
}
func (s *_Irpc_TestServiceDivReq) Deserialize(d *irpc.Decoder) error {
	if err := d.Int(&s.Param0_a); err != nil {
		return fmt.Errorf("deserialize s.Param0_a of type 'int': %w", err)
	}
	if err := d.Int(&s.Param0_b); err != nil {
		return fmt.Errorf("deserialize s.Param0_b of type 'int': %w", err)
	}
	return nil
}

type _Irpc_TestServiceDivResp struct {
	Param0_ int
}

func (s _Irpc_TestServiceDivResp) Serialize(e *irpc.Encoder) error {
	if err := e.Int(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'int': %w", err)
	}
	return nil
}
func (s *_Irpc_TestServiceDivResp) Deserialize(d *irpc.Decoder) error {
	if err := d.Int(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'int': %w", err)
	}
	return nil
}

type _Irpc_TestServiceDivErrReq struct {
	Param0_a int
	Param0_b int
}

func (s _Irpc_TestServiceDivErrReq) Serialize(e *irpc.Encoder) error {
	if err := e.Int(s.Param0_a); err != nil {
		return fmt.Errorf("serialize s.Param0_a of type 'int': %w", err)
	}
	if err := e.Int(s.Param0_b); err != nil {
		return fmt.Errorf("serialize s.Param0_b of type 'int': %w", err)
	}
	return nil
}
func (s *_Irpc_TestServiceDivErrReq) Deserialize(d *irpc.Decoder) error {
	if err := d.Int(&s.Param0_a); err != nil {
		return fmt.Errorf("deserialize s.Param0_a of type 'int': %w", err)
	}
	if err := d.Int(&s.Param0_b); err != nil {
		return fmt.Errorf("deserialize s.Param0_b of type 'int': %w", err)
	}
	return nil
}

type _Irpc_TestServiceDivErrResp struct {
	Param0_ int
	Param1_ error
}

func (s _Irpc_TestServiceDivErrResp) Serialize(e *irpc.Encoder) error {
	if err := e.Int(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'int': %w", err)
	}
	{
		var isNil bool
		if s.Param1_ == nil {
			isNil = true
		}
		if err := e.Bool(isNil); err != nil {
			return fmt.Errorf("serialize isNil of type 'bool': %w", err)
		}

		if !isNil {
			{ // Error()
				_Error_0_ := s.Param1_.Error()
				if err := e.String(_Error_0_); err != nil {
					return fmt.Errorf("serialize _Error_0_ of type 'string': %w", err)
				}
			}
		}
	}
	return nil
}
func (s *_Irpc_TestServiceDivErrResp) Deserialize(d *irpc.Decoder) error {
	if err := d.Int(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'int': %w", err)
	}
	{
		var isNil bool
		if err := d.Bool(&isNil); err != nil {
			return fmt.Errorf("deserialize isNil of type 'bool': %w", err)
		}

		if isNil {
			s.Param1_ = nil
		} else {
			var impl _error_TestService_irpcInterfaceImpl
			{ // Error()
				if err := d.String(&impl._Error_0_); err != nil {
					return fmt.Errorf("deserialize impl._Error_0_ of type 'string': %w", err)
				}
			}
			s.Param1_ = impl
		}
	}
	return nil
}

type _error_TestService_irpcInterfaceImpl struct {
	_Error_0_ string
}

func (i _error_TestService_irpcInterfaceImpl) Error() string {
	return i._Error_0_
}