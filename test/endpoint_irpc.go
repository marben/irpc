// Code generated by irpc generator; DO NOT EDIT
package irpctestpkg

import (
	"fmt"
	"github.com/marben/irpc/pkg/irpc"
)

type endpointApiIRpcService struct {
	impl endpointApi
}

func newEndpointApiIRpcService(impl endpointApi) *endpointApiIRpcService {
	return &endpointApiIRpcService{impl: impl}
}
func (endpointApiIRpcService) Hash() []byte {
	return []byte("endpointApiIRpcService")
}
func (s *endpointApiIRpcService) GetFuncCall(funcId irpc.FuncId) (irpc.ArgDeserializer, error) {
	switch funcId {
	case 0: // Div
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_endpointApiDivReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_endpointApiDivResp
				resp.Param0_, resp.Param1_ = s.impl.Div(args.Param0_a, args.Param0_b)
				return resp
			}, nil
		}, nil
	default:
		return nil, fmt.Errorf("function '%d' doesn't exist on service '%s'", funcId, string(s.Hash()))
	}
}

type endpointApiIRpcClient struct {
	endpoint *irpc.Endpoint
	id       irpc.RegisteredServiceId
}

func newEndpointApiIRpcClient(endpoint *irpc.Endpoint) (*endpointApiIRpcClient, error) {
	id, err := endpoint.RegisterClient([]byte("endpointApiIRpcService"))
	if err != nil {
		return nil, fmt.Errorf("register failed: %w", err)
	}
	return &endpointApiIRpcClient{endpoint: endpoint, id: id}, nil
}
func (_c *endpointApiIRpcClient) Div(a float64, b float64) (float64, error) {
	var req = _Irpc_endpointApiDivReq{
		Param0_a: a,
		Param0_b: b,
	}
	var resp _Irpc_endpointApiDivResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 0, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_, resp.Param1_
}

type _Irpc_endpointApiDivReq struct {
	Param0_a float64
	Param0_b float64
}

func (s _Irpc_endpointApiDivReq) Serialize(e *irpc.Encoder) error {
	if err := e.Float64(s.Param0_a); err != nil {
		return fmt.Errorf("serialize s.Param0_a of type 'float64': %w", err)
	}
	if err := e.Float64(s.Param0_b); err != nil {
		return fmt.Errorf("serialize s.Param0_b of type 'float64': %w", err)
	}
	return nil
}
func (s *_Irpc_endpointApiDivReq) Deserialize(d *irpc.Decoder) error {
	if err := d.Float64(&s.Param0_a); err != nil {
		return fmt.Errorf("deserialize s.Param0_a of type 'float64': %w", err)
	}
	if err := d.Float64(&s.Param0_b); err != nil {
		return fmt.Errorf("deserialize s.Param0_b of type 'float64': %w", err)
	}
	return nil
}

type _Irpc_endpointApiDivResp struct {
	Param0_ float64
	Param1_ error
}

func (s _Irpc_endpointApiDivResp) Serialize(e *irpc.Encoder) error {
	if err := e.Float64(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'float64': %w", err)
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
func (s *_Irpc_endpointApiDivResp) Deserialize(d *irpc.Decoder) error {
	if err := d.Float64(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'float64': %w", err)
	}
	{
		var isNil bool
		if err := d.Bool(&isNil); err != nil {
			return fmt.Errorf("deserialize isNil of type 'bool': %w", err)
		}

		if isNil {
			s.Param1_ = nil
		} else {
			var impl _error_endpointApi_irpcInterfaceImpl
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

type _error_endpointApi_irpcInterfaceImpl struct {
	_Error_0_ string
}

func (i _error_endpointApi_irpcInterfaceImpl) Error() string {
	return i._Error_0_
}
