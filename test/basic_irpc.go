// Code generated by irpc generator; DO NOT EDIT
package irpctestpkg

import (
	"fmt"
	"github.com/marben/irpc/pkg/irpc"
)

type basicAPIRpcService struct {
	impl basicAPI
}

func newBasicAPIRpcService(impl basicAPI) *basicAPIRpcService {
	return &basicAPIRpcService{impl: impl}
}
func (basicAPIRpcService) Hash() []byte {
	return []byte("basicAPIRpcService")
}
func (s *basicAPIRpcService) GetFuncCall(funcId irpc.FuncId) (irpc.ArgDeserializer, error) {
	switch funcId {
	case 0: // addByte
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_basicAPIaddByteReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_basicAPIaddByteResp
				resp.Param0_ = s.impl.addByte(args.Param0_a, args.Param1_b)
				return resp
			}, nil
		}, nil
	case 1: // addInt
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_basicAPIaddIntReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_basicAPIaddIntResp
				resp.Param0_ = s.impl.addInt(args.Param0_a, args.Param1_b)
				return resp
			}, nil
		}, nil
	case 2: // swapInt
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_basicAPIswapIntReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_basicAPIswapIntResp
				resp.Param0_, resp.Param1_ = s.impl.swapInt(args.Param0_a, args.Param1_b)
				return resp
			}, nil
		}, nil
	case 3: // subUint
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_basicAPIsubUintReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_basicAPIsubUintResp
				resp.Param0_ = s.impl.subUint(args.Param0_a, args.Param0_b)
				return resp
			}, nil
		}, nil
	case 4: // addInt8
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_basicAPIaddInt8Req
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_basicAPIaddInt8Resp
				resp.Param0_ = s.impl.addInt8(args.Param0_a, args.Param0_b)
				return resp
			}, nil
		}, nil
	case 5: // addUint8
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_basicAPIaddUint8Req
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_basicAPIaddUint8Resp
				resp.Param0_ = s.impl.addUint8(args.Param0_a, args.Param0_b)
				return resp
			}, nil
		}, nil
	case 6: // addInt16
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_basicAPIaddInt16Req
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_basicAPIaddInt16Resp
				resp.Param0_ = s.impl.addInt16(args.Param0_a, args.Param0_b)
				return resp
			}, nil
		}, nil
	case 7: // addUint16
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_basicAPIaddUint16Req
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_basicAPIaddUint16Resp
				resp.Param0_ = s.impl.addUint16(args.Param0_a, args.Param0_b)
				return resp
			}, nil
		}, nil
	case 8: // addInt32
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_basicAPIaddInt32Req
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_basicAPIaddInt32Resp
				resp.Param0_ = s.impl.addInt32(args.Param0_a, args.Param0_b)
				return resp
			}, nil
		}, nil
	case 9: // addUint32
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_basicAPIaddUint32Req
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_basicAPIaddUint32Resp
				resp.Param0_ = s.impl.addUint32(args.Param0_a, args.Param0_b)
				return resp
			}, nil
		}, nil
	case 10: // addInt64
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_basicAPIaddInt64Req
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_basicAPIaddInt64Resp
				resp.Param0_ = s.impl.addInt64(args.Param0_a, args.Param0_b)
				return resp
			}, nil
		}, nil
	case 11: // addUint64
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_basicAPIaddUint64Req
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_basicAPIaddUint64Resp
				resp.Param0_ = s.impl.addUint64(args.Param0_a, args.Param0_b)
				return resp
			}, nil
		}, nil
	case 12: // addFloat64
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_basicAPIaddFloat64Req
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_basicAPIaddFloat64Resp
				resp.Param0_ = s.impl.addFloat64(args.Param0_a, args.Param0_b)
				return resp
			}, nil
		}, nil
	case 13: // addFloat32
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_basicAPIaddFloat32Req
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_basicAPIaddFloat32Resp
				resp.Param0_ = s.impl.addFloat32(args.Param0_a, args.Param0_b)
				return resp
			}, nil
		}, nil
	case 14: // toUpper
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_basicAPItoUpperReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_basicAPItoUpperResp
				resp.Param0_ = s.impl.toUpper(args.Param0_c)
				return resp
			}, nil
		}, nil
	case 15: // toUpperString
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_basicAPItoUpperStringReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_basicAPItoUpperStringResp
				resp.Param0_ = s.impl.toUpperString(args.Param0_s)
				return resp
			}, nil
		}, nil
	case 16: // negBool
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_basicAPInegBoolReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_basicAPInegBoolResp
				resp.Param0_ = s.impl.negBool(args.Param0_ok)
				return resp
			}, nil
		}, nil
	default:
		return nil, fmt.Errorf("function '%d' doesn't exist on service '%s'", funcId, string(s.Hash()))
	}
}

type emptyAPIRpcService struct {
	impl emptyAPI
}

func newEmptyAPIRpcService(impl emptyAPI) *emptyAPIRpcService {
	return &emptyAPIRpcService{impl: impl}
}
func (emptyAPIRpcService) Hash() []byte {
	return []byte("emptyAPIRpcService")
}
func (s *emptyAPIRpcService) GetFuncCall(funcId irpc.FuncId) (irpc.ArgDeserializer, error) {
	switch funcId {
	default:
		return nil, fmt.Errorf("function '%d' doesn't exist on service '%s'", funcId, string(s.Hash()))
	}
}

type basicAPIRpcClient struct {
	endpoint *irpc.Endpoint
	id       irpc.RegisteredServiceId
}

func newBasicAPIRpcClient(endpoint *irpc.Endpoint) (*basicAPIRpcClient, error) {
	id, err := endpoint.RegisterClient([]byte("basicAPIRpcService"))
	if err != nil {
		return nil, fmt.Errorf("register failed: %w", err)
	}
	return &basicAPIRpcClient{endpoint: endpoint, id: id}, nil
}
func (_c *basicAPIRpcClient) addByte(a byte, b byte) byte {
	var req = _Irpc_basicAPIaddByteReq{
		Param0_a: a,
		Param1_b: b,
	}
	var resp _Irpc_basicAPIaddByteResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 0, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *basicAPIRpcClient) addInt(a int, b int) int {
	var req = _Irpc_basicAPIaddIntReq{
		Param0_a: a,
		Param1_b: b,
	}
	var resp _Irpc_basicAPIaddIntResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 1, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *basicAPIRpcClient) swapInt(a int, b int) (int, int) {
	var req = _Irpc_basicAPIswapIntReq{
		Param0_a: a,
		Param1_b: b,
	}
	var resp _Irpc_basicAPIswapIntResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 2, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_, resp.Param1_
}
func (_c *basicAPIRpcClient) subUint(a uint, b uint) uint {
	var req = _Irpc_basicAPIsubUintReq{
		Param0_a: a,
		Param0_b: b,
	}
	var resp _Irpc_basicAPIsubUintResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 3, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *basicAPIRpcClient) addInt8(a int8, b int8) int8 {
	var req = _Irpc_basicAPIaddInt8Req{
		Param0_a: a,
		Param0_b: b,
	}
	var resp _Irpc_basicAPIaddInt8Resp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 4, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *basicAPIRpcClient) addUint8(a uint8, b uint8) uint8 {
	var req = _Irpc_basicAPIaddUint8Req{
		Param0_a: a,
		Param0_b: b,
	}
	var resp _Irpc_basicAPIaddUint8Resp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 5, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *basicAPIRpcClient) addInt16(a int16, b int16) int16 {
	var req = _Irpc_basicAPIaddInt16Req{
		Param0_a: a,
		Param0_b: b,
	}
	var resp _Irpc_basicAPIaddInt16Resp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 6, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *basicAPIRpcClient) addUint16(a uint16, b uint16) uint16 {
	var req = _Irpc_basicAPIaddUint16Req{
		Param0_a: a,
		Param0_b: b,
	}
	var resp _Irpc_basicAPIaddUint16Resp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 7, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *basicAPIRpcClient) addInt32(a int32, b int32) int32 {
	var req = _Irpc_basicAPIaddInt32Req{
		Param0_a: a,
		Param0_b: b,
	}
	var resp _Irpc_basicAPIaddInt32Resp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 8, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *basicAPIRpcClient) addUint32(a uint32, b uint32) uint32 {
	var req = _Irpc_basicAPIaddUint32Req{
		Param0_a: a,
		Param0_b: b,
	}
	var resp _Irpc_basicAPIaddUint32Resp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 9, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *basicAPIRpcClient) addInt64(a int64, b int64) int64 {
	var req = _Irpc_basicAPIaddInt64Req{
		Param0_a: a,
		Param0_b: b,
	}
	var resp _Irpc_basicAPIaddInt64Resp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 10, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *basicAPIRpcClient) addUint64(a uint64, b uint64) uint64 {
	var req = _Irpc_basicAPIaddUint64Req{
		Param0_a: a,
		Param0_b: b,
	}
	var resp _Irpc_basicAPIaddUint64Resp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 11, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *basicAPIRpcClient) addFloat64(a float64, b float64) float64 {
	var req = _Irpc_basicAPIaddFloat64Req{
		Param0_a: a,
		Param0_b: b,
	}
	var resp _Irpc_basicAPIaddFloat64Resp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 12, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *basicAPIRpcClient) addFloat32(a float32, b float32) float32 {
	var req = _Irpc_basicAPIaddFloat32Req{
		Param0_a: a,
		Param0_b: b,
	}
	var resp _Irpc_basicAPIaddFloat32Resp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 13, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *basicAPIRpcClient) toUpper(c rune) rune {
	var req = _Irpc_basicAPItoUpperReq{
		Param0_c: c,
	}
	var resp _Irpc_basicAPItoUpperResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 14, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *basicAPIRpcClient) toUpperString(s string) string {
	var req = _Irpc_basicAPItoUpperStringReq{
		Param0_s: s,
	}
	var resp _Irpc_basicAPItoUpperStringResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 15, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *basicAPIRpcClient) negBool(ok bool) bool {
	var req = _Irpc_basicAPInegBoolReq{
		Param0_ok: ok,
	}
	var resp _Irpc_basicAPInegBoolResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 16, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}

type emptyAPIRpcClient struct {
	endpoint *irpc.Endpoint
	id       irpc.RegisteredServiceId
}

func newEmptyAPIRpcClient(endpoint *irpc.Endpoint) (*emptyAPIRpcClient, error) {
	id, err := endpoint.RegisterClient([]byte("emptyAPIRpcService"))
	if err != nil {
		return nil, fmt.Errorf("register failed: %w", err)
	}
	return &emptyAPIRpcClient{endpoint: endpoint, id: id}, nil
}

type _Irpc_basicAPIaddByteReq struct {
	Param0_a byte
	Param1_b byte
}

func (s _Irpc_basicAPIaddByteReq) Serialize(e *irpc.Encoder) error {
	if err := e.Uint8(s.Param0_a); err != nil {
		return fmt.Errorf("serialize s.Param0_a of type 'uint8': %w", err)
	}
	if err := e.Uint8(s.Param1_b); err != nil {
		return fmt.Errorf("serialize s.Param1_b of type 'uint8': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddByteReq) Deserialize(d *irpc.Decoder) error {
	if err := d.Uint8(&s.Param0_a); err != nil {
		return fmt.Errorf("deserialize s.Param0_a of type 'uint8': %w", err)
	}
	if err := d.Uint8(&s.Param1_b); err != nil {
		return fmt.Errorf("deserialize s.Param1_b of type 'uint8': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddByteResp struct {
	Param0_ byte
}

func (s _Irpc_basicAPIaddByteResp) Serialize(e *irpc.Encoder) error {
	if err := e.Uint8(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'uint8': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddByteResp) Deserialize(d *irpc.Decoder) error {
	if err := d.Uint8(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'uint8': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddIntReq struct {
	Param0_a int
	Param1_b int
}

func (s _Irpc_basicAPIaddIntReq) Serialize(e *irpc.Encoder) error {
	if err := e.Int(s.Param0_a); err != nil {
		return fmt.Errorf("serialize s.Param0_a of type 'int': %w", err)
	}
	if err := e.Int(s.Param1_b); err != nil {
		return fmt.Errorf("serialize s.Param1_b of type 'int': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddIntReq) Deserialize(d *irpc.Decoder) error {
	if err := d.Int(&s.Param0_a); err != nil {
		return fmt.Errorf("deserialize s.Param0_a of type 'int': %w", err)
	}
	if err := d.Int(&s.Param1_b); err != nil {
		return fmt.Errorf("deserialize s.Param1_b of type 'int': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddIntResp struct {
	Param0_ int
}

func (s _Irpc_basicAPIaddIntResp) Serialize(e *irpc.Encoder) error {
	if err := e.Int(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'int': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddIntResp) Deserialize(d *irpc.Decoder) error {
	if err := d.Int(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'int': %w", err)
	}
	return nil
}

type _Irpc_basicAPIswapIntReq struct {
	Param0_a int
	Param1_b int
}

func (s _Irpc_basicAPIswapIntReq) Serialize(e *irpc.Encoder) error {
	if err := e.Int(s.Param0_a); err != nil {
		return fmt.Errorf("serialize s.Param0_a of type 'int': %w", err)
	}
	if err := e.Int(s.Param1_b); err != nil {
		return fmt.Errorf("serialize s.Param1_b of type 'int': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIswapIntReq) Deserialize(d *irpc.Decoder) error {
	if err := d.Int(&s.Param0_a); err != nil {
		return fmt.Errorf("deserialize s.Param0_a of type 'int': %w", err)
	}
	if err := d.Int(&s.Param1_b); err != nil {
		return fmt.Errorf("deserialize s.Param1_b of type 'int': %w", err)
	}
	return nil
}

type _Irpc_basicAPIswapIntResp struct {
	Param0_ int
	Param1_ int
}

func (s _Irpc_basicAPIswapIntResp) Serialize(e *irpc.Encoder) error {
	if err := e.Int(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'int': %w", err)
	}
	if err := e.Int(s.Param1_); err != nil {
		return fmt.Errorf("serialize s.Param1_ of type 'int': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIswapIntResp) Deserialize(d *irpc.Decoder) error {
	if err := d.Int(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'int': %w", err)
	}
	if err := d.Int(&s.Param1_); err != nil {
		return fmt.Errorf("deserialize s.Param1_ of type 'int': %w", err)
	}
	return nil
}

type _Irpc_basicAPIsubUintReq struct {
	Param0_a uint
	Param0_b uint
}

func (s _Irpc_basicAPIsubUintReq) Serialize(e *irpc.Encoder) error {
	if err := e.Uint(s.Param0_a); err != nil {
		return fmt.Errorf("serialize s.Param0_a of type 'uint': %w", err)
	}
	if err := e.Uint(s.Param0_b); err != nil {
		return fmt.Errorf("serialize s.Param0_b of type 'uint': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIsubUintReq) Deserialize(d *irpc.Decoder) error {
	if err := d.Uint(&s.Param0_a); err != nil {
		return fmt.Errorf("deserialize s.Param0_a of type 'uint': %w", err)
	}
	if err := d.Uint(&s.Param0_b); err != nil {
		return fmt.Errorf("deserialize s.Param0_b of type 'uint': %w", err)
	}
	return nil
}

type _Irpc_basicAPIsubUintResp struct {
	Param0_ uint
}

func (s _Irpc_basicAPIsubUintResp) Serialize(e *irpc.Encoder) error {
	if err := e.Uint(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'uint': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIsubUintResp) Deserialize(d *irpc.Decoder) error {
	if err := d.Uint(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'uint': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddInt8Req struct {
	Param0_a int8
	Param0_b int8
}

func (s _Irpc_basicAPIaddInt8Req) Serialize(e *irpc.Encoder) error {
	if err := e.Int8(s.Param0_a); err != nil {
		return fmt.Errorf("serialize s.Param0_a of type 'int8': %w", err)
	}
	if err := e.Int8(s.Param0_b); err != nil {
		return fmt.Errorf("serialize s.Param0_b of type 'int8': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddInt8Req) Deserialize(d *irpc.Decoder) error {
	if err := d.Int8(&s.Param0_a); err != nil {
		return fmt.Errorf("deserialize s.Param0_a of type 'int8': %w", err)
	}
	if err := d.Int8(&s.Param0_b); err != nil {
		return fmt.Errorf("deserialize s.Param0_b of type 'int8': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddInt8Resp struct {
	Param0_ int8
}

func (s _Irpc_basicAPIaddInt8Resp) Serialize(e *irpc.Encoder) error {
	if err := e.Int8(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'int8': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddInt8Resp) Deserialize(d *irpc.Decoder) error {
	if err := d.Int8(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'int8': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddUint8Req struct {
	Param0_a uint8
	Param0_b uint8
}

func (s _Irpc_basicAPIaddUint8Req) Serialize(e *irpc.Encoder) error {
	if err := e.Uint8(s.Param0_a); err != nil {
		return fmt.Errorf("serialize s.Param0_a of type 'uint8': %w", err)
	}
	if err := e.Uint8(s.Param0_b); err != nil {
		return fmt.Errorf("serialize s.Param0_b of type 'uint8': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddUint8Req) Deserialize(d *irpc.Decoder) error {
	if err := d.Uint8(&s.Param0_a); err != nil {
		return fmt.Errorf("deserialize s.Param0_a of type 'uint8': %w", err)
	}
	if err := d.Uint8(&s.Param0_b); err != nil {
		return fmt.Errorf("deserialize s.Param0_b of type 'uint8': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddUint8Resp struct {
	Param0_ uint8
}

func (s _Irpc_basicAPIaddUint8Resp) Serialize(e *irpc.Encoder) error {
	if err := e.Uint8(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'uint8': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddUint8Resp) Deserialize(d *irpc.Decoder) error {
	if err := d.Uint8(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'uint8': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddInt16Req struct {
	Param0_a int16
	Param0_b int16
}

func (s _Irpc_basicAPIaddInt16Req) Serialize(e *irpc.Encoder) error {
	if err := e.Int16(s.Param0_a); err != nil {
		return fmt.Errorf("serialize s.Param0_a of type 'int16': %w", err)
	}
	if err := e.Int16(s.Param0_b); err != nil {
		return fmt.Errorf("serialize s.Param0_b of type 'int16': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddInt16Req) Deserialize(d *irpc.Decoder) error {
	if err := d.Int16(&s.Param0_a); err != nil {
		return fmt.Errorf("deserialize s.Param0_a of type 'int16': %w", err)
	}
	if err := d.Int16(&s.Param0_b); err != nil {
		return fmt.Errorf("deserialize s.Param0_b of type 'int16': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddInt16Resp struct {
	Param0_ int16
}

func (s _Irpc_basicAPIaddInt16Resp) Serialize(e *irpc.Encoder) error {
	if err := e.Int16(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'int16': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddInt16Resp) Deserialize(d *irpc.Decoder) error {
	if err := d.Int16(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'int16': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddUint16Req struct {
	Param0_a uint16
	Param0_b uint16
}

func (s _Irpc_basicAPIaddUint16Req) Serialize(e *irpc.Encoder) error {
	if err := e.Uint16(s.Param0_a); err != nil {
		return fmt.Errorf("serialize s.Param0_a of type 'uint16': %w", err)
	}
	if err := e.Uint16(s.Param0_b); err != nil {
		return fmt.Errorf("serialize s.Param0_b of type 'uint16': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddUint16Req) Deserialize(d *irpc.Decoder) error {
	if err := d.Uint16(&s.Param0_a); err != nil {
		return fmt.Errorf("deserialize s.Param0_a of type 'uint16': %w", err)
	}
	if err := d.Uint16(&s.Param0_b); err != nil {
		return fmt.Errorf("deserialize s.Param0_b of type 'uint16': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddUint16Resp struct {
	Param0_ uint16
}

func (s _Irpc_basicAPIaddUint16Resp) Serialize(e *irpc.Encoder) error {
	if err := e.Uint16(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'uint16': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddUint16Resp) Deserialize(d *irpc.Decoder) error {
	if err := d.Uint16(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'uint16': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddInt32Req struct {
	Param0_a int32
	Param0_b int32
}

func (s _Irpc_basicAPIaddInt32Req) Serialize(e *irpc.Encoder) error {
	if err := e.Int32(s.Param0_a); err != nil {
		return fmt.Errorf("serialize s.Param0_a of type 'int32': %w", err)
	}
	if err := e.Int32(s.Param0_b); err != nil {
		return fmt.Errorf("serialize s.Param0_b of type 'int32': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddInt32Req) Deserialize(d *irpc.Decoder) error {
	if err := d.Int32(&s.Param0_a); err != nil {
		return fmt.Errorf("deserialize s.Param0_a of type 'int32': %w", err)
	}
	if err := d.Int32(&s.Param0_b); err != nil {
		return fmt.Errorf("deserialize s.Param0_b of type 'int32': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddInt32Resp struct {
	Param0_ int32
}

func (s _Irpc_basicAPIaddInt32Resp) Serialize(e *irpc.Encoder) error {
	if err := e.Int32(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'int32': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddInt32Resp) Deserialize(d *irpc.Decoder) error {
	if err := d.Int32(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'int32': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddUint32Req struct {
	Param0_a uint32
	Param0_b uint32
}

func (s _Irpc_basicAPIaddUint32Req) Serialize(e *irpc.Encoder) error {
	if err := e.Uint32(s.Param0_a); err != nil {
		return fmt.Errorf("serialize s.Param0_a of type 'uint32': %w", err)
	}
	if err := e.Uint32(s.Param0_b); err != nil {
		return fmt.Errorf("serialize s.Param0_b of type 'uint32': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddUint32Req) Deserialize(d *irpc.Decoder) error {
	if err := d.Uint32(&s.Param0_a); err != nil {
		return fmt.Errorf("deserialize s.Param0_a of type 'uint32': %w", err)
	}
	if err := d.Uint32(&s.Param0_b); err != nil {
		return fmt.Errorf("deserialize s.Param0_b of type 'uint32': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddUint32Resp struct {
	Param0_ uint32
}

func (s _Irpc_basicAPIaddUint32Resp) Serialize(e *irpc.Encoder) error {
	if err := e.Uint32(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'uint32': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddUint32Resp) Deserialize(d *irpc.Decoder) error {
	if err := d.Uint32(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'uint32': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddInt64Req struct {
	Param0_a int64
	Param0_b int64
}

func (s _Irpc_basicAPIaddInt64Req) Serialize(e *irpc.Encoder) error {
	if err := e.Int64(s.Param0_a); err != nil {
		return fmt.Errorf("serialize s.Param0_a of type 'int64': %w", err)
	}
	if err := e.Int64(s.Param0_b); err != nil {
		return fmt.Errorf("serialize s.Param0_b of type 'int64': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddInt64Req) Deserialize(d *irpc.Decoder) error {
	if err := d.Int64(&s.Param0_a); err != nil {
		return fmt.Errorf("deserialize s.Param0_a of type 'int64': %w", err)
	}
	if err := d.Int64(&s.Param0_b); err != nil {
		return fmt.Errorf("deserialize s.Param0_b of type 'int64': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddInt64Resp struct {
	Param0_ int64
}

func (s _Irpc_basicAPIaddInt64Resp) Serialize(e *irpc.Encoder) error {
	if err := e.Int64(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'int64': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddInt64Resp) Deserialize(d *irpc.Decoder) error {
	if err := d.Int64(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'int64': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddUint64Req struct {
	Param0_a uint64
	Param0_b uint64
}

func (s _Irpc_basicAPIaddUint64Req) Serialize(e *irpc.Encoder) error {
	if err := e.Uint64(s.Param0_a); err != nil {
		return fmt.Errorf("serialize s.Param0_a of type 'uint64': %w", err)
	}
	if err := e.Uint64(s.Param0_b); err != nil {
		return fmt.Errorf("serialize s.Param0_b of type 'uint64': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddUint64Req) Deserialize(d *irpc.Decoder) error {
	if err := d.Uint64(&s.Param0_a); err != nil {
		return fmt.Errorf("deserialize s.Param0_a of type 'uint64': %w", err)
	}
	if err := d.Uint64(&s.Param0_b); err != nil {
		return fmt.Errorf("deserialize s.Param0_b of type 'uint64': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddUint64Resp struct {
	Param0_ uint64
}

func (s _Irpc_basicAPIaddUint64Resp) Serialize(e *irpc.Encoder) error {
	if err := e.Uint64(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'uint64': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddUint64Resp) Deserialize(d *irpc.Decoder) error {
	if err := d.Uint64(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'uint64': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddFloat64Req struct {
	Param0_a float64
	Param0_b float64
}

func (s _Irpc_basicAPIaddFloat64Req) Serialize(e *irpc.Encoder) error {
	if err := e.Float64(s.Param0_a); err != nil {
		return fmt.Errorf("serialize s.Param0_a of type 'float64': %w", err)
	}
	if err := e.Float64(s.Param0_b); err != nil {
		return fmt.Errorf("serialize s.Param0_b of type 'float64': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddFloat64Req) Deserialize(d *irpc.Decoder) error {
	if err := d.Float64(&s.Param0_a); err != nil {
		return fmt.Errorf("deserialize s.Param0_a of type 'float64': %w", err)
	}
	if err := d.Float64(&s.Param0_b); err != nil {
		return fmt.Errorf("deserialize s.Param0_b of type 'float64': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddFloat64Resp struct {
	Param0_ float64
}

func (s _Irpc_basicAPIaddFloat64Resp) Serialize(e *irpc.Encoder) error {
	if err := e.Float64(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'float64': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddFloat64Resp) Deserialize(d *irpc.Decoder) error {
	if err := d.Float64(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'float64': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddFloat32Req struct {
	Param0_a float32
	Param0_b float32
}

func (s _Irpc_basicAPIaddFloat32Req) Serialize(e *irpc.Encoder) error {
	if err := e.Float32(s.Param0_a); err != nil {
		return fmt.Errorf("serialize s.Param0_a of type 'float32': %w", err)
	}
	if err := e.Float32(s.Param0_b); err != nil {
		return fmt.Errorf("serialize s.Param0_b of type 'float32': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddFloat32Req) Deserialize(d *irpc.Decoder) error {
	if err := d.Float32(&s.Param0_a); err != nil {
		return fmt.Errorf("deserialize s.Param0_a of type 'float32': %w", err)
	}
	if err := d.Float32(&s.Param0_b); err != nil {
		return fmt.Errorf("deserialize s.Param0_b of type 'float32': %w", err)
	}
	return nil
}

type _Irpc_basicAPIaddFloat32Resp struct {
	Param0_ float32
}

func (s _Irpc_basicAPIaddFloat32Resp) Serialize(e *irpc.Encoder) error {
	if err := e.Float32(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'float32': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPIaddFloat32Resp) Deserialize(d *irpc.Decoder) error {
	if err := d.Float32(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'float32': %w", err)
	}
	return nil
}

type _Irpc_basicAPItoUpperReq struct {
	Param0_c rune
}

func (s _Irpc_basicAPItoUpperReq) Serialize(e *irpc.Encoder) error {
	if err := e.Int32(s.Param0_c); err != nil {
		return fmt.Errorf("serialize s.Param0_c of type 'int32': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPItoUpperReq) Deserialize(d *irpc.Decoder) error {
	if err := d.Int32(&s.Param0_c); err != nil {
		return fmt.Errorf("deserialize s.Param0_c of type 'int32': %w", err)
	}
	return nil
}

type _Irpc_basicAPItoUpperResp struct {
	Param0_ rune
}

func (s _Irpc_basicAPItoUpperResp) Serialize(e *irpc.Encoder) error {
	if err := e.Int32(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'int32': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPItoUpperResp) Deserialize(d *irpc.Decoder) error {
	if err := d.Int32(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'int32': %w", err)
	}
	return nil
}

type _Irpc_basicAPItoUpperStringReq struct {
	Param0_s string
}

func (s _Irpc_basicAPItoUpperStringReq) Serialize(e *irpc.Encoder) error {
	if err := e.String(s.Param0_s); err != nil {
		return fmt.Errorf("serialize s.Param0_s of type 'string': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPItoUpperStringReq) Deserialize(d *irpc.Decoder) error {
	if err := d.String(&s.Param0_s); err != nil {
		return fmt.Errorf("deserialize s.Param0_s of type 'string': %w", err)
	}
	return nil
}

type _Irpc_basicAPItoUpperStringResp struct {
	Param0_ string
}

func (s _Irpc_basicAPItoUpperStringResp) Serialize(e *irpc.Encoder) error {
	if err := e.String(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'string': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPItoUpperStringResp) Deserialize(d *irpc.Decoder) error {
	if err := d.String(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'string': %w", err)
	}
	return nil
}

type _Irpc_basicAPInegBoolReq struct {
	Param0_ok bool
}

func (s _Irpc_basicAPInegBoolReq) Serialize(e *irpc.Encoder) error {
	if err := e.Bool(s.Param0_ok); err != nil {
		return fmt.Errorf("serialize s.Param0_ok of type 'bool': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPInegBoolReq) Deserialize(d *irpc.Decoder) error {
	if err := d.Bool(&s.Param0_ok); err != nil {
		return fmt.Errorf("deserialize s.Param0_ok of type 'bool': %w", err)
	}
	return nil
}

type _Irpc_basicAPInegBoolResp struct {
	Param0_ bool
}

func (s _Irpc_basicAPInegBoolResp) Serialize(e *irpc.Encoder) error {
	if err := e.Bool(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'bool': %w", err)
	}
	return nil
}
func (s *_Irpc_basicAPInegBoolResp) Deserialize(d *irpc.Decoder) error {
	if err := d.Bool(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'bool': %w", err)
	}
	return nil
}
