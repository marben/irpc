// Code generated by irpc generator; DO NOT EDIT
package irpctestpkg

import (
	"context"
	"fmt"
	"github.com/marben/irpc/cmd/irpc/test/out"
	"github.com/marben/irpc/irpcgen"
)

type outsideTestIRpcService struct {
	impl outsideTest
	id   []byte
}

func newOutsideTestIRpcService(impl outsideTest) *outsideTestIRpcService {
	return &outsideTestIRpcService{
		impl: impl,
		id:   []byte{226, 145, 184, 233, 94, 168, 110, 136, 234, 134, 72, 225, 142, 9, 110, 191, 161, 229, 206, 231, 180, 253, 30, 48, 28, 78, 14, 81, 76, 163, 5, 89},
	}
}
func (s *outsideTestIRpcService) Id() []byte {
	return s.id
}
func (s *outsideTestIRpcService) GetFuncCall(funcId irpcgen.FuncId) (irpcgen.ArgDeserializer, error) {
	switch funcId {
	case 0: // addUint8
		return func(d *irpcgen.Decoder) (irpcgen.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_outsideTestaddUint8Req
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func(ctx context.Context) irpcgen.Serializable {
				// EXECUTE
				var resp _Irpc_outsideTestaddUint8Resp
				resp.Param0 = s.impl.addUint8(args.Param0_a, args.Param0_b)
				return resp
			}, nil
		}, nil
	default:
		return nil, fmt.Errorf("function '%d' doesn't exist on service '%s'", funcId, s.Id())
	}
}

type outsideTestIRpcClient struct {
	endpoint irpcgen.Endpoint
	id       []byte
}

func newOutsideTestIRpcClient(endpoint irpcgen.Endpoint) (*outsideTestIRpcClient, error) {
	id := []byte{226, 145, 184, 233, 94, 168, 110, 136, 234, 134, 72, 225, 142, 9, 110, 191, 161, 229, 206, 231, 180, 253, 30, 48, 28, 78, 14, 81, 76, 163, 5, 89}
	if err := endpoint.RegisterClient(id); err != nil {
		return nil, fmt.Errorf("register failed: %w", err)
	}
	return &outsideTestIRpcClient{endpoint: endpoint, id: id}, nil
}
func (_c *outsideTestIRpcClient) addUint8(a out.Uint8, b out.Uint8) out.Uint8 {
	var req = _Irpc_outsideTestaddUint8Req{
		Param0_a: a,
		Param0_b: b,
	}
	var resp _Irpc_outsideTestaddUint8Resp
	if err := _c.endpoint.CallRemoteFunc(context.Background(), _c.id, 0, req, &resp); err != nil {
		panic(err) // to avoid panic, make your func return error and regenerate the code
	}
	return resp.Param0
}

type _Irpc_outsideTestaddUint8Req struct {
	Param0_a out.Uint8
	Param0_b out.Uint8
}

func (s _Irpc_outsideTestaddUint8Req) Serialize(e *irpcgen.Encoder) error {
	if err := e.Uint8(uint8(s.Param0_a)); err != nil {
		return fmt.Errorf("serialize s.Param0_a of type 'uint8': %w", err)
	}
	if err := e.Uint8(uint8(s.Param0_b)); err != nil {
		return fmt.Errorf("serialize s.Param0_b of type 'uint8': %w", err)
	}
	return nil
}
func (s *_Irpc_outsideTestaddUint8Req) Deserialize(d *irpcgen.Decoder) error {
	if err := d.Uint8((*uint8)(&s.Param0_a)); err != nil {
		return fmt.Errorf("deserialize s.Param0_a of type 'uint8': %w", err)
	}
	if err := d.Uint8((*uint8)(&s.Param0_b)); err != nil {
		return fmt.Errorf("deserialize s.Param0_b of type 'uint8': %w", err)
	}
	return nil
}

type _Irpc_outsideTestaddUint8Resp struct {
	Param0 out.Uint8
}

func (s _Irpc_outsideTestaddUint8Resp) Serialize(e *irpcgen.Encoder) error {
	if err := e.Uint8(uint8(s.Param0)); err != nil {
		return fmt.Errorf("serialize s.Param0 of type 'uint8': %w", err)
	}
	return nil
}
func (s *_Irpc_outsideTestaddUint8Resp) Deserialize(d *irpcgen.Decoder) error {
	if err := d.Uint8((*uint8)(&s.Param0)); err != nil {
		return fmt.Errorf("deserialize s.Param0 of type 'uint8': %w", err)
	}
	return nil
}
