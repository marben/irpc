// Code generated by irpc generator; DO NOT EDIT
package irpctestpkg

import (
	"context"
	"fmt"
	"github.com/marben/irpc/pkg/irpc"
)

type MathIRpcService struct {
	impl Math
}

func NewMathIRpcService(impl Math) *MathIRpcService {
	return &MathIRpcService{impl: impl}
}
func (MathIRpcService) Hash() []byte {
	return []byte("MathIRpcService")
}
func (s *MathIRpcService) GetFuncCall(funcId irpc.FuncId) (irpc.ArgDeserializer, error) {
	switch funcId {
	case 0: // Add
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_MathAddReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func(ctx context.Context) irpc.Serializable {
				// EXECUTE
				var resp _Irpc_MathAddResp
				resp.Param0_, resp.Param1_ = s.impl.Add(args.Param0_a, args.Param0_b)
				return resp
			}, nil
		}, nil
	default:
		return nil, fmt.Errorf("function '%d' doesn't exist on service '%s'", funcId, string(s.Hash()))
	}
}

type MathIRpcClient struct {
	endpoint *irpc.Endpoint
	id       irpc.RegisteredServiceId
}

func NewMathIRpcClient(endpoint *irpc.Endpoint) (*MathIRpcClient, error) {
	id, err := endpoint.RegisterClient(context.Background(), []byte("MathIRpcService"))
	if err != nil {
		return nil, fmt.Errorf("register failed: %w", err)
	}
	return &MathIRpcClient{endpoint: endpoint, id: id}, nil
}
func (_c *MathIRpcClient) Add(a int, b int) (int, error) {
	var req = _Irpc_MathAddReq{
		Param0_a: a,
		Param0_b: b,
	}
	var resp _Irpc_MathAddResp
	if err := _c.endpoint.CallRemoteFunc(context.Background(), _c.id, 0, req, &resp); err != nil {
		var zero _Irpc_MathAddResp
		return zero.Param0_, err
	}
	return resp.Param0_, resp.Param1_
}

type _Irpc_MathAddReq struct {
	Param0_a int
	Param0_b int
}

func (s _Irpc_MathAddReq) Serialize(e *irpc.Encoder) error {
	if err := e.VarInt(s.Param0_a); err != nil {
		return fmt.Errorf("serialize s.Param0_a of type 'int': %w", err)
	}
	if err := e.VarInt(s.Param0_b); err != nil {
		return fmt.Errorf("serialize s.Param0_b of type 'int': %w", err)
	}
	return nil
}
func (s *_Irpc_MathAddReq) Deserialize(d *irpc.Decoder) error {
	if err := d.VarInt(&s.Param0_a); err != nil {
		return fmt.Errorf("deserialize s.Param0_a of type 'int': %w", err)
	}
	if err := d.VarInt(&s.Param0_b); err != nil {
		return fmt.Errorf("deserialize s.Param0_b of type 'int': %w", err)
	}
	return nil
}

type _Irpc_MathAddResp struct {
	Param0_ int
	Param1_ error
}

func (s _Irpc_MathAddResp) Serialize(e *irpc.Encoder) error {
	if err := e.VarInt(s.Param0_); err != nil {
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
func (s *_Irpc_MathAddResp) Deserialize(d *irpc.Decoder) error {
	if err := d.VarInt(&s.Param0_); err != nil {
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
			var impl _error_Math_irpcInterfaceImpl
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

type _error_Math_irpcInterfaceImpl struct {
	_Error_0_ string
}

func (i _error_Math_irpcInterfaceImpl) Error() string {
	return i._Error_0_
}
