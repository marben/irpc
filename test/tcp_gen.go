// Code generated by irpc generator; DO NOT EDIT
package irpctestpkg

import (
	"encoding/binary"
	"fmt"
	"github.com/marben/irpc/pkg/irpc"
	"io"
	"math"
)

type tcpTestApiRpcService struct {
	impl tcpTestApi
}

func newTcpTestApiRpcService(impl tcpTestApi) *tcpTestApiRpcService {
	return &tcpTestApiRpcService{impl: impl}
}
func (tcpTestApiRpcService) Hash() []byte {
	return []byte("tcpTestApiRpcService")
}
func (s *tcpTestApiRpcService) GetFuncCall(funcId irpc.FuncId) (irpc.ArgDeserializer, error) {
	switch funcId {
	case 0:

		return func(r io.Reader) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_tcpTestApiDivReq
			if err := args.Deserialize(r); err != nil {
				return nil, err
			}
			return func() (irpc.Serializable, error) {
				// EXECUTE
				var resp _Irpc_tcpTestApiDivResp
				resp.Param0_, resp.Param1_ = s.impl.Div(args.Param0_a, args.Param0_b)
				return resp, nil
			}, nil
		}, nil
	default:
		return nil, fmt.Errorf("function '%d' doesn't exist on service '%s'", funcId, string(s.Hash()))
	}
}

type tcpTestApiRpcClient struct {
	endpoint *irpc.Endpoint
	id       irpc.RegisteredServiceId
}

func newTcpTestApiRpcClient(endpoint *irpc.Endpoint) (*tcpTestApiRpcClient, error) {
	id, err := endpoint.RegisterClient([]byte("tcpTestApiRpcService"))
	if err != nil {
		return nil, fmt.Errorf("register failed: %w", err)
	}
	return &tcpTestApiRpcClient{endpoint: endpoint, id: id}, nil
}
func (_c *tcpTestApiRpcClient) Div(a float64, b float64) (float64, error) {
	var req = _Irpc_tcpTestApiDivReq{
		Param0_a: a,
		Param0_b: b,
	}
	var resp _Irpc_tcpTestApiDivResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 0, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_, resp.Param1_
}

type _Irpc_tcpTestApiDivReq struct {
	Param0_a float64
	Param0_b float64
}

func (s _Irpc_tcpTestApiDivReq) Serialize(w io.Writer) error {
	{ // float64
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, math.Float64bits(s.Param0_a))
		if _, err := w.Write(b[:8]); err != nil {
			return fmt.Errorf("s.Param0_a float64 write: %w", err)
		}
	}
	{ // float64
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, math.Float64bits(s.Param0_b))
		if _, err := w.Write(b[:8]); err != nil {
			return fmt.Errorf("s.Param0_b float64 write: %w", err)
		}
	}
	return nil
}
func (s *_Irpc_tcpTestApiDivReq) Deserialize(r io.Reader) error {
	{ // float64
		b := make([]byte, 8)
		if _, err := io.ReadFull(r, b[:8]); err != nil {
			return fmt.Errorf("s.Param0_a float64 decode: %w", err)
		}
		s.Param0_a = math.Float64frombits(binary.LittleEndian.Uint64(b))
	}
	{ // float64
		b := make([]byte, 8)
		if _, err := io.ReadFull(r, b[:8]); err != nil {
			return fmt.Errorf("s.Param0_b float64 decode: %w", err)
		}
		s.Param0_b = math.Float64frombits(binary.LittleEndian.Uint64(b))
	}
	return nil
}

type _Irpc_tcpTestApiDivResp struct {
	Param0_ float64
	Param1_ error
}

func (s _Irpc_tcpTestApiDivResp) Serialize(w io.Writer) error {
	{ // float64
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, math.Float64bits(s.Param0_))
		if _, err := w.Write(b[:8]); err != nil {
			return fmt.Errorf("s.Param0_ float64 write: %w", err)
		}
	}
	{ // error
		var isNil bool
		if s.Param1_ == nil {
			isNil = true
		}
		b := make([]byte, 1)
		if isNil {
			b[0] = 1
		} else {
			b[0] = 0
		}
		if _, err := w.Write(b[:1]); err != nil {
			return fmt.Errorf("isNil bool write: %w", err)
		}

		if !isNil {
			{ // Error()
				_Error_0_ := s.Param1_.Error()
				var l int = len(_Error_0_)
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, uint64(l))
				if _, err := w.Write(b[:8]); err != nil {
					return fmt.Errorf("l int write: %w", err)
				}

				_, err := w.Write([]byte(_Error_0_))
				if err != nil {
					return fmt.Errorf("failed to write string to writer: %w", err)
				}
			}
		}
	}
	return nil
}
func (s *_Irpc_tcpTestApiDivResp) Deserialize(r io.Reader) error {
	{ // float64
		b := make([]byte, 8)
		if _, err := io.ReadFull(r, b[:8]); err != nil {
			return fmt.Errorf("s.Param0_ float64 decode: %w", err)
		}
		s.Param0_ = math.Float64frombits(binary.LittleEndian.Uint64(b))
	}
	{ // error
		var isNil bool
		b := make([]byte, 1)
		if _, err := io.ReadFull(r, b[:1]); err != nil {
			return fmt.Errorf("isNil bool decode: %w", err)
		}
		if b[0] == 0 {
			isNil = false
		} else {
			isNil = true
		}
		if isNil {
			s.Param1_ = nil
		} else {
			var impl _error_tcpTestApi_irpcInterfaceImpl
			{ // Error()
				{
					var l int
					b := make([]byte, 8)
					if _, err := io.ReadFull(r, b[:8]); err != nil {
						return fmt.Errorf("l int decode: %w", err)
					}
					l = int(binary.LittleEndian.Uint64(b))
					sbuf := make([]byte, l)
					_, err := io.ReadFull(r, sbuf)
					if err != nil {
						return fmt.Errorf("failed to read string data from reader: %w", err)
					}
					impl._Error_0_ = string(sbuf)
				}
			}
			s.Param1_ = impl
		}
	}
	return nil
}

type _error_tcpTestApi_irpcInterfaceImpl struct {
	_Error_0_ string
}

func (i _error_tcpTestApi_irpcInterfaceImpl) Error() string {
	return i._Error_0_
}
