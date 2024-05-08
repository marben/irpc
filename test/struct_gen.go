// Code generated by irpc generator; DO NOT EDIT
package irpctestpkg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/marben/irpc/pkg/irpc"
	"io"
)

type structAPIRpcService struct {
	impl structAPI
}

func newStructAPIRpcService(impl structAPI) *structAPIRpcService {
	return &structAPIRpcService{impl: impl}
}
func (structAPIRpcService) Id() string {
	return "structAPIRpcService"
}
func (s *structAPIRpcService) CallFunc(funcName string, args []byte) ([]byte, error) {
	switch funcName {
	case "VectSum":
		return s.callVectSum(args)
	case "Vect3x3Sum":
		return s.callVect3x3Sum(args)
	case "SumSliceStruct":
		return s.callSumSliceStruct(args)
	default:
		return nil, fmt.Errorf("function '%s' doesn't exist on service '%s'", funcName, s.Id())
	}
}
func (s *structAPIRpcService) callVectSum(params []byte) ([]byte, error) {
	r := bytes.NewBuffer(params)
	var req _Irpc_structAPIVectSumReq
	if err := req.Deserialize(r); err != nil {
		return nil, fmt.Errorf("failed to deserialize VectSum: %w", err)
	}
	var resp _Irpc_structAPIVectSumResp
	resp.Param0_ = s.impl.VectSum(req.Param0_v)
	b := bytes.NewBuffer(nil)
	err := resp.Serialize(b)
	if err != nil {
		return nil, fmt.Errorf("response serialization failed: %w", err)
	}
	return b.Bytes(), nil
}
func (s *structAPIRpcService) callVect3x3Sum(params []byte) ([]byte, error) {
	r := bytes.NewBuffer(params)
	var req _Irpc_structAPIVect3x3SumReq
	if err := req.Deserialize(r); err != nil {
		return nil, fmt.Errorf("failed to deserialize Vect3x3Sum: %w", err)
	}
	var resp _Irpc_structAPIVect3x3SumResp
	resp.Param0_ = s.impl.Vect3x3Sum(req.Param0_v)
	b := bytes.NewBuffer(nil)
	err := resp.Serialize(b)
	if err != nil {
		return nil, fmt.Errorf("response serialization failed: %w", err)
	}
	return b.Bytes(), nil
}
func (s *structAPIRpcService) callSumSliceStruct(params []byte) ([]byte, error) {
	r := bytes.NewBuffer(params)
	var req _Irpc_structAPISumSliceStructReq
	if err := req.Deserialize(r); err != nil {
		return nil, fmt.Errorf("failed to deserialize SumSliceStruct: %w", err)
	}
	var resp _Irpc_structAPISumSliceStructResp
	resp.Param0_ = s.impl.SumSliceStruct(req.Param0_s)
	b := bytes.NewBuffer(nil)
	err := resp.Serialize(b)
	if err != nil {
		return nil, fmt.Errorf("response serialization failed: %w", err)
	}
	return b.Bytes(), nil
}

type structAPIRpcClient struct {
	endpoint *irpc.Endpoint
}

func newStructAPIRpcClient(endpoint *irpc.Endpoint) *structAPIRpcClient {
	return &structAPIRpcClient{endpoint: endpoint}
}
func (_c *structAPIRpcClient) VectSum(v vect3) int {
	var req = _Irpc_structAPIVectSumReq{
		Param0_v: v,
	}
	var resp _Irpc_structAPIVectSumResp
	if err := _c.endpoint.CallRemoteFunc("structAPIRpcService", "VectSum", req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *structAPIRpcClient) Vect3x3Sum(v vect3x3) vect3 {
	var req = _Irpc_structAPIVect3x3SumReq{
		Param0_v: v,
	}
	var resp _Irpc_structAPIVect3x3SumResp
	if err := _c.endpoint.CallRemoteFunc("structAPIRpcService", "Vect3x3Sum", req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *structAPIRpcClient) SumSliceStruct(s sliceStruct) int {
	var req = _Irpc_structAPISumSliceStructReq{
		Param0_s: s,
	}
	var resp _Irpc_structAPISumSliceStructResp
	if err := _c.endpoint.CallRemoteFunc("structAPIRpcService", "SumSliceStruct", req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}

type _Irpc_structAPIVectSumReq struct {
	Param0_v vect3
}

func (s _Irpc_structAPIVectSumReq) Serialize(w io.Writer) error {
	{ // vect3
		{
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(s.Param0_v.a))
			if _, err := w.Write(b[:8]); err != nil {
				return fmt.Errorf("s.Param0_v.a int write: %w", err)
			}
		}
		{
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(s.Param0_v.b))
			if _, err := w.Write(b[:8]); err != nil {
				return fmt.Errorf("s.Param0_v.b int write: %w", err)
			}
		}
		{
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(s.Param0_v.c))
			if _, err := w.Write(b[:8]); err != nil {
				return fmt.Errorf("s.Param0_v.c int write: %w", err)
			}
		}
	}
	return nil
}
func (s *_Irpc_structAPIVectSumReq) Deserialize(r io.Reader) error {
	{ // vect3
		{
			b := make([]byte, 8)
			if _, err := io.ReadFull(r, b[:8]); err != nil {
				return fmt.Errorf("s.Param0_v.a int decode: %w", err)
			}
			s.Param0_v.a = int(binary.LittleEndian.Uint64(b))
		}
		{
			b := make([]byte, 8)
			if _, err := io.ReadFull(r, b[:8]); err != nil {
				return fmt.Errorf("s.Param0_v.b int decode: %w", err)
			}
			s.Param0_v.b = int(binary.LittleEndian.Uint64(b))
		}
		{
			b := make([]byte, 8)
			if _, err := io.ReadFull(r, b[:8]); err != nil {
				return fmt.Errorf("s.Param0_v.c int decode: %w", err)
			}
			s.Param0_v.c = int(binary.LittleEndian.Uint64(b))
		}
	}
	return nil
}

type _Irpc_structAPIVectSumResp struct {
	Param0_ int
}

func (s _Irpc_structAPIVectSumResp) Serialize(w io.Writer) error {
	{ // int
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(s.Param0_))
		if _, err := w.Write(b[:8]); err != nil {
			return fmt.Errorf("s.Param0_ int write: %w", err)
		}
	}
	return nil
}
func (s *_Irpc_structAPIVectSumResp) Deserialize(r io.Reader) error {
	{ // int
		b := make([]byte, 8)
		if _, err := io.ReadFull(r, b[:8]); err != nil {
			return fmt.Errorf("s.Param0_ int decode: %w", err)
		}
		s.Param0_ = int(binary.LittleEndian.Uint64(b))
	}
	return nil
}

type _Irpc_structAPIVect3x3SumReq struct {
	Param0_v vect3x3
}

func (s _Irpc_structAPIVect3x3SumReq) Serialize(w io.Writer) error {
	{ // vect3x3
		{
			{
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, uint64(s.Param0_v.v1.a))
				if _, err := w.Write(b[:8]); err != nil {
					return fmt.Errorf("s.Param0_v.v1.a int write: %w", err)
				}
			}
			{
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, uint64(s.Param0_v.v1.b))
				if _, err := w.Write(b[:8]); err != nil {
					return fmt.Errorf("s.Param0_v.v1.b int write: %w", err)
				}
			}
			{
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, uint64(s.Param0_v.v1.c))
				if _, err := w.Write(b[:8]); err != nil {
					return fmt.Errorf("s.Param0_v.v1.c int write: %w", err)
				}
			}
		}
		{
			{
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, uint64(s.Param0_v.v2.a))
				if _, err := w.Write(b[:8]); err != nil {
					return fmt.Errorf("s.Param0_v.v2.a int write: %w", err)
				}
			}
			{
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, uint64(s.Param0_v.v2.b))
				if _, err := w.Write(b[:8]); err != nil {
					return fmt.Errorf("s.Param0_v.v2.b int write: %w", err)
				}
			}
			{
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, uint64(s.Param0_v.v2.c))
				if _, err := w.Write(b[:8]); err != nil {
					return fmt.Errorf("s.Param0_v.v2.c int write: %w", err)
				}
			}
		}
		{
			{
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, uint64(s.Param0_v.v3.a))
				if _, err := w.Write(b[:8]); err != nil {
					return fmt.Errorf("s.Param0_v.v3.a int write: %w", err)
				}
			}
			{
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, uint64(s.Param0_v.v3.b))
				if _, err := w.Write(b[:8]); err != nil {
					return fmt.Errorf("s.Param0_v.v3.b int write: %w", err)
				}
			}
			{
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, uint64(s.Param0_v.v3.c))
				if _, err := w.Write(b[:8]); err != nil {
					return fmt.Errorf("s.Param0_v.v3.c int write: %w", err)
				}
			}
		}
	}
	return nil
}
func (s *_Irpc_structAPIVect3x3SumReq) Deserialize(r io.Reader) error {
	{ // vect3x3
		{
			{
				b := make([]byte, 8)
				if _, err := io.ReadFull(r, b[:8]); err != nil {
					return fmt.Errorf("s.Param0_v.v1.a int decode: %w", err)
				}
				s.Param0_v.v1.a = int(binary.LittleEndian.Uint64(b))
			}
			{
				b := make([]byte, 8)
				if _, err := io.ReadFull(r, b[:8]); err != nil {
					return fmt.Errorf("s.Param0_v.v1.b int decode: %w", err)
				}
				s.Param0_v.v1.b = int(binary.LittleEndian.Uint64(b))
			}
			{
				b := make([]byte, 8)
				if _, err := io.ReadFull(r, b[:8]); err != nil {
					return fmt.Errorf("s.Param0_v.v1.c int decode: %w", err)
				}
				s.Param0_v.v1.c = int(binary.LittleEndian.Uint64(b))
			}
		}
		{
			{
				b := make([]byte, 8)
				if _, err := io.ReadFull(r, b[:8]); err != nil {
					return fmt.Errorf("s.Param0_v.v2.a int decode: %w", err)
				}
				s.Param0_v.v2.a = int(binary.LittleEndian.Uint64(b))
			}
			{
				b := make([]byte, 8)
				if _, err := io.ReadFull(r, b[:8]); err != nil {
					return fmt.Errorf("s.Param0_v.v2.b int decode: %w", err)
				}
				s.Param0_v.v2.b = int(binary.LittleEndian.Uint64(b))
			}
			{
				b := make([]byte, 8)
				if _, err := io.ReadFull(r, b[:8]); err != nil {
					return fmt.Errorf("s.Param0_v.v2.c int decode: %w", err)
				}
				s.Param0_v.v2.c = int(binary.LittleEndian.Uint64(b))
			}
		}
		{
			{
				b := make([]byte, 8)
				if _, err := io.ReadFull(r, b[:8]); err != nil {
					return fmt.Errorf("s.Param0_v.v3.a int decode: %w", err)
				}
				s.Param0_v.v3.a = int(binary.LittleEndian.Uint64(b))
			}
			{
				b := make([]byte, 8)
				if _, err := io.ReadFull(r, b[:8]); err != nil {
					return fmt.Errorf("s.Param0_v.v3.b int decode: %w", err)
				}
				s.Param0_v.v3.b = int(binary.LittleEndian.Uint64(b))
			}
			{
				b := make([]byte, 8)
				if _, err := io.ReadFull(r, b[:8]); err != nil {
					return fmt.Errorf("s.Param0_v.v3.c int decode: %w", err)
				}
				s.Param0_v.v3.c = int(binary.LittleEndian.Uint64(b))
			}
		}
	}
	return nil
}

type _Irpc_structAPIVect3x3SumResp struct {
	Param0_ vect3
}

func (s _Irpc_structAPIVect3x3SumResp) Serialize(w io.Writer) error {
	{ // vect3
		{
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(s.Param0_.a))
			if _, err := w.Write(b[:8]); err != nil {
				return fmt.Errorf("s.Param0_.a int write: %w", err)
			}
		}
		{
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(s.Param0_.b))
			if _, err := w.Write(b[:8]); err != nil {
				return fmt.Errorf("s.Param0_.b int write: %w", err)
			}
		}
		{
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(s.Param0_.c))
			if _, err := w.Write(b[:8]); err != nil {
				return fmt.Errorf("s.Param0_.c int write: %w", err)
			}
		}
	}
	return nil
}
func (s *_Irpc_structAPIVect3x3SumResp) Deserialize(r io.Reader) error {
	{ // vect3
		{
			b := make([]byte, 8)
			if _, err := io.ReadFull(r, b[:8]); err != nil {
				return fmt.Errorf("s.Param0_.a int decode: %w", err)
			}
			s.Param0_.a = int(binary.LittleEndian.Uint64(b))
		}
		{
			b := make([]byte, 8)
			if _, err := io.ReadFull(r, b[:8]); err != nil {
				return fmt.Errorf("s.Param0_.b int decode: %w", err)
			}
			s.Param0_.b = int(binary.LittleEndian.Uint64(b))
		}
		{
			b := make([]byte, 8)
			if _, err := io.ReadFull(r, b[:8]); err != nil {
				return fmt.Errorf("s.Param0_.c int decode: %w", err)
			}
			s.Param0_.c = int(binary.LittleEndian.Uint64(b))
		}
	}
	return nil
}

type _Irpc_structAPISumSliceStructReq struct {
	Param0_s sliceStruct
}

func (s _Irpc_structAPISumSliceStructReq) Serialize(w io.Writer) error {
	{ // sliceStruct
		{
			var l int = len(s.Param0_s.s1)
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(l))
			if _, err := w.Write(b[:8]); err != nil {
				return fmt.Errorf("l int write: %w", err)
			}

			for i := 0; i < l; i++ {
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, uint64(s.Param0_s.s1[i]))
				if _, err := w.Write(b[:8]); err != nil {
					return fmt.Errorf("s.Param0_s.s1[i] int write: %w", err)
				}

			}
		}
		{
			var l int = len(s.Param0_s.s2)
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(l))
			if _, err := w.Write(b[:8]); err != nil {
				return fmt.Errorf("l int write: %w", err)
			}

			for i := 0; i < l; i++ {
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, uint64(s.Param0_s.s2[i]))
				if _, err := w.Write(b[:8]); err != nil {
					return fmt.Errorf("s.Param0_s.s2[i] int write: %w", err)
				}

			}
		}
	}
	return nil
}
func (s *_Irpc_structAPISumSliceStructReq) Deserialize(r io.Reader) error {
	{ // sliceStruct
		{
			var l int
			b := make([]byte, 8)
			if _, err := io.ReadFull(r, b[:8]); err != nil {
				return fmt.Errorf("l int decode: %w", err)
			}
			l = int(binary.LittleEndian.Uint64(b))
			s.Param0_s.s1 = make([]int, l)
			for i := 0; i < l; i++ {
				b := make([]byte, 8)
				if _, err := io.ReadFull(r, b[:8]); err != nil {
					return fmt.Errorf("s.Param0_s.s1[i] int decode: %w", err)
				}
				s.Param0_s.s1[i] = int(binary.LittleEndian.Uint64(b))
			}
		}
		{
			var l int
			b := make([]byte, 8)
			if _, err := io.ReadFull(r, b[:8]); err != nil {
				return fmt.Errorf("l int decode: %w", err)
			}
			l = int(binary.LittleEndian.Uint64(b))
			s.Param0_s.s2 = make([]int, l)
			for i := 0; i < l; i++ {
				b := make([]byte, 8)
				if _, err := io.ReadFull(r, b[:8]); err != nil {
					return fmt.Errorf("s.Param0_s.s2[i] int decode: %w", err)
				}
				s.Param0_s.s2[i] = int(binary.LittleEndian.Uint64(b))
			}
		}
	}
	return nil
}

type _Irpc_structAPISumSliceStructResp struct {
	Param0_ int
}

func (s _Irpc_structAPISumSliceStructResp) Serialize(w io.Writer) error {
	{ // int
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(s.Param0_))
		if _, err := w.Write(b[:8]); err != nil {
			return fmt.Errorf("s.Param0_ int write: %w", err)
		}
	}
	return nil
}
func (s *_Irpc_structAPISumSliceStructResp) Deserialize(r io.Reader) error {
	{ // int
		b := make([]byte, 8)
		if _, err := io.ReadFull(r, b[:8]); err != nil {
			return fmt.Errorf("s.Param0_ int decode: %w", err)
		}
		s.Param0_ = int(binary.LittleEndian.Uint64(b))
	}
	return nil
}
