// Code generated by irpc generator; DO NOT EDIT
package irpctestpkg

import (
	"context"
	"fmt"
	"github.com/marben/irpc/pkg/irpc"
)

type sliceTestIRpcService struct {
	impl sliceTest
}

func newSliceTestIRpcService(impl sliceTest) *sliceTestIRpcService {
	return &sliceTestIRpcService{impl: impl}
}
func (sliceTestIRpcService) Hash() []byte {
	return []byte("sliceTestIRpcService")
}
func (s *sliceTestIRpcService) GetFuncCall(funcId irpc.FuncId) (irpc.ArgDeserializer, error) {
	switch funcId {
	case 0: // SliceSum
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_sliceTestSliceSumReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func(ctx context.Context) irpc.Serializable {
				// EXECUTE
				var resp _Irpc_sliceTestSliceSumResp
				resp.Param0_ = s.impl.SliceSum(args.Param0_slice)
				return resp
			}, nil
		}, nil
	case 1: // VectMult
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_sliceTestVectMultReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func(ctx context.Context) irpc.Serializable {
				// EXECUTE
				var resp _Irpc_sliceTestVectMultResp
				resp.Param0_ = s.impl.VectMult(args.Param0_vect, args.Param1_s)
				return resp
			}, nil
		}, nil
	case 2: // SliceOfFloat64Sum
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_sliceTestSliceOfFloat64SumReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func(ctx context.Context) irpc.Serializable {
				// EXECUTE
				var resp _Irpc_sliceTestSliceOfFloat64SumResp
				resp.Param0_ = s.impl.SliceOfFloat64Sum(args.Param0_slice)
				return resp
			}, nil
		}, nil
	case 3: // SliceOfSlicesSum
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_sliceTestSliceOfSlicesSumReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func(ctx context.Context) irpc.Serializable {
				// EXECUTE
				var resp _Irpc_sliceTestSliceOfSlicesSumResp
				resp.Param0_ = s.impl.SliceOfSlicesSum(args.Param0_slice)
				return resp
			}, nil
		}, nil
	case 4: // SliceOfBytesSum
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_sliceTestSliceOfBytesSumReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func(ctx context.Context) irpc.Serializable {
				// EXECUTE
				var resp _Irpc_sliceTestSliceOfBytesSumResp
				resp.Param0_ = s.impl.SliceOfBytesSum(args.Param0_slice)
				return resp
			}, nil
		}, nil
	default:
		return nil, fmt.Errorf("function '%d' doesn't exist on service '%s'", funcId, string(s.Hash()))
	}
}

type sliceTestIRpcClient struct {
	endpoint *irpc.Endpoint
	id       irpc.RegisteredServiceId
}

func newSliceTestIRpcClient(endpoint *irpc.Endpoint) (*sliceTestIRpcClient, error) {
	id, err := endpoint.RegisterClient(context.Background(), []byte("sliceTestIRpcService"))
	if err != nil {
		return nil, fmt.Errorf("register failed: %w", err)
	}
	return &sliceTestIRpcClient{endpoint: endpoint, id: id}, nil
}
func (_c *sliceTestIRpcClient) SliceSum(slice []int) int {
	var req = _Irpc_sliceTestSliceSumReq{
		Param0_slice: slice,
	}
	var resp _Irpc_sliceTestSliceSumResp
	if err := _c.endpoint.CallRemoteFunc(context.Background(), _c.id, 0, req, &resp); err != nil {
		panic(err) // to avoid panic, make your func return error and regenerate the code
	}
	return resp.Param0_
}
func (_c *sliceTestIRpcClient) VectMult(vect []int, s int) []int {
	var req = _Irpc_sliceTestVectMultReq{
		Param0_vect: vect,
		Param1_s:    s,
	}
	var resp _Irpc_sliceTestVectMultResp
	if err := _c.endpoint.CallRemoteFunc(context.Background(), _c.id, 1, req, &resp); err != nil {
		panic(err) // to avoid panic, make your func return error and regenerate the code
	}
	return resp.Param0_
}
func (_c *sliceTestIRpcClient) SliceOfFloat64Sum(slice []float64) float64 {
	var req = _Irpc_sliceTestSliceOfFloat64SumReq{
		Param0_slice: slice,
	}
	var resp _Irpc_sliceTestSliceOfFloat64SumResp
	if err := _c.endpoint.CallRemoteFunc(context.Background(), _c.id, 2, req, &resp); err != nil {
		panic(err) // to avoid panic, make your func return error and regenerate the code
	}
	return resp.Param0_
}
func (_c *sliceTestIRpcClient) SliceOfSlicesSum(slice [][]int) int {
	var req = _Irpc_sliceTestSliceOfSlicesSumReq{
		Param0_slice: slice,
	}
	var resp _Irpc_sliceTestSliceOfSlicesSumResp
	if err := _c.endpoint.CallRemoteFunc(context.Background(), _c.id, 3, req, &resp); err != nil {
		panic(err) // to avoid panic, make your func return error and regenerate the code
	}
	return resp.Param0_
}
func (_c *sliceTestIRpcClient) SliceOfBytesSum(slice []byte) int {
	var req = _Irpc_sliceTestSliceOfBytesSumReq{
		Param0_slice: slice,
	}
	var resp _Irpc_sliceTestSliceOfBytesSumResp
	if err := _c.endpoint.CallRemoteFunc(context.Background(), _c.id, 4, req, &resp); err != nil {
		panic(err) // to avoid panic, make your func return error and regenerate the code
	}
	return resp.Param0_
}

type _Irpc_sliceTestSliceSumReq struct {
	Param0_slice []int
}

func (s _Irpc_sliceTestSliceSumReq) Serialize(e *irpc.Encoder) error {
	{ // s.Param0_slice
		var l int = len(s.Param0_slice)
		if err := e.Int(l); err != nil {
			return fmt.Errorf("serialize l of type 'int': %w", err)
		}

		for i := 0; i < l; i++ {
			if err := e.Int(s.Param0_slice[i]); err != nil {
				return fmt.Errorf("serialize s.Param0_slice[i] of type 'int': %w", err)
			}

		}
	}
	return nil
}
func (s *_Irpc_sliceTestSliceSumReq) Deserialize(d *irpc.Decoder) error {
	{ // s.Param0_slice
		var l int
		if err := d.Int(&l); err != nil {
			return fmt.Errorf("deserialize l of type 'int': %w", err)
		}

		s.Param0_slice = make([]int, l)
		for i := 0; i < l; i++ {
			if err := d.Int(&s.Param0_slice[i]); err != nil {
				return fmt.Errorf("deserialize s.Param0_slice[i] of type 'int': %w", err)
			}

		}
	}
	return nil
}

type _Irpc_sliceTestSliceSumResp struct {
	Param0_ int
}

func (s _Irpc_sliceTestSliceSumResp) Serialize(e *irpc.Encoder) error {
	if err := e.Int(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'int': %w", err)
	}
	return nil
}
func (s *_Irpc_sliceTestSliceSumResp) Deserialize(d *irpc.Decoder) error {
	if err := d.Int(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'int': %w", err)
	}
	return nil
}

type _Irpc_sliceTestVectMultReq struct {
	Param0_vect []int
	Param1_s    int
}

func (s _Irpc_sliceTestVectMultReq) Serialize(e *irpc.Encoder) error {
	{ // s.Param0_vect
		var l int = len(s.Param0_vect)
		if err := e.Int(l); err != nil {
			return fmt.Errorf("serialize l of type 'int': %w", err)
		}

		for i := 0; i < l; i++ {
			if err := e.Int(s.Param0_vect[i]); err != nil {
				return fmt.Errorf("serialize s.Param0_vect[i] of type 'int': %w", err)
			}

		}
	}
	if err := e.Int(s.Param1_s); err != nil {
		return fmt.Errorf("serialize s.Param1_s of type 'int': %w", err)
	}
	return nil
}
func (s *_Irpc_sliceTestVectMultReq) Deserialize(d *irpc.Decoder) error {
	{ // s.Param0_vect
		var l int
		if err := d.Int(&l); err != nil {
			return fmt.Errorf("deserialize l of type 'int': %w", err)
		}

		s.Param0_vect = make([]int, l)
		for i := 0; i < l; i++ {
			if err := d.Int(&s.Param0_vect[i]); err != nil {
				return fmt.Errorf("deserialize s.Param0_vect[i] of type 'int': %w", err)
			}

		}
	}
	if err := d.Int(&s.Param1_s); err != nil {
		return fmt.Errorf("deserialize s.Param1_s of type 'int': %w", err)
	}
	return nil
}

type _Irpc_sliceTestVectMultResp struct {
	Param0_ []int
}

func (s _Irpc_sliceTestVectMultResp) Serialize(e *irpc.Encoder) error {
	{ // s.Param0_
		var l int = len(s.Param0_)
		if err := e.Int(l); err != nil {
			return fmt.Errorf("serialize l of type 'int': %w", err)
		}

		for i := 0; i < l; i++ {
			if err := e.Int(s.Param0_[i]); err != nil {
				return fmt.Errorf("serialize s.Param0_[i] of type 'int': %w", err)
			}

		}
	}
	return nil
}
func (s *_Irpc_sliceTestVectMultResp) Deserialize(d *irpc.Decoder) error {
	{ // s.Param0_
		var l int
		if err := d.Int(&l); err != nil {
			return fmt.Errorf("deserialize l of type 'int': %w", err)
		}

		s.Param0_ = make([]int, l)
		for i := 0; i < l; i++ {
			if err := d.Int(&s.Param0_[i]); err != nil {
				return fmt.Errorf("deserialize s.Param0_[i] of type 'int': %w", err)
			}

		}
	}
	return nil
}

type _Irpc_sliceTestSliceOfFloat64SumReq struct {
	Param0_slice []float64
}

func (s _Irpc_sliceTestSliceOfFloat64SumReq) Serialize(e *irpc.Encoder) error {
	{ // s.Param0_slice
		var l int = len(s.Param0_slice)
		if err := e.Int(l); err != nil {
			return fmt.Errorf("serialize l of type 'int': %w", err)
		}

		for i := 0; i < l; i++ {
			if err := e.Float64(s.Param0_slice[i]); err != nil {
				return fmt.Errorf("serialize s.Param0_slice[i] of type 'float64': %w", err)
			}

		}
	}
	return nil
}
func (s *_Irpc_sliceTestSliceOfFloat64SumReq) Deserialize(d *irpc.Decoder) error {
	{ // s.Param0_slice
		var l int
		if err := d.Int(&l); err != nil {
			return fmt.Errorf("deserialize l of type 'int': %w", err)
		}

		s.Param0_slice = make([]float64, l)
		for i := 0; i < l; i++ {
			if err := d.Float64(&s.Param0_slice[i]); err != nil {
				return fmt.Errorf("deserialize s.Param0_slice[i] of type 'float64': %w", err)
			}

		}
	}
	return nil
}

type _Irpc_sliceTestSliceOfFloat64SumResp struct {
	Param0_ float64
}

func (s _Irpc_sliceTestSliceOfFloat64SumResp) Serialize(e *irpc.Encoder) error {
	if err := e.Float64(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'float64': %w", err)
	}
	return nil
}
func (s *_Irpc_sliceTestSliceOfFloat64SumResp) Deserialize(d *irpc.Decoder) error {
	if err := d.Float64(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'float64': %w", err)
	}
	return nil
}

type _Irpc_sliceTestSliceOfSlicesSumReq struct {
	Param0_slice [][]int
}

func (s _Irpc_sliceTestSliceOfSlicesSumReq) Serialize(e *irpc.Encoder) error {
	{ // s.Param0_slice
		var l int = len(s.Param0_slice)
		if err := e.Int(l); err != nil {
			return fmt.Errorf("serialize l of type 'int': %w", err)
		}

		for i := 0; i < l; i++ {
			{ // s.Param0_slice[i]
				var l int = len(s.Param0_slice[i])
				if err := e.Int(l); err != nil {
					return fmt.Errorf("serialize l of type 'int': %w", err)
				}

				for j := 0; j < l; j++ {
					if err := e.Int(s.Param0_slice[i][j]); err != nil {
						return fmt.Errorf("serialize s.Param0_slice[i][j] of type 'int': %w", err)
					}

				}
			}

		}
	}
	return nil
}
func (s *_Irpc_sliceTestSliceOfSlicesSumReq) Deserialize(d *irpc.Decoder) error {
	{ // s.Param0_slice
		var l int
		if err := d.Int(&l); err != nil {
			return fmt.Errorf("deserialize l of type 'int': %w", err)
		}

		s.Param0_slice = make([][]int, l)
		for i := 0; i < l; i++ {
			{ // s.Param0_slice[i]
				var l int
				if err := d.Int(&l); err != nil {
					return fmt.Errorf("deserialize l of type 'int': %w", err)
				}

				s.Param0_slice[i] = make([]int, l)
				for j := 0; j < l; j++ {
					if err := d.Int(&s.Param0_slice[i][j]); err != nil {
						return fmt.Errorf("deserialize s.Param0_slice[i][j] of type 'int': %w", err)
					}

				}
			}

		}
	}
	return nil
}

type _Irpc_sliceTestSliceOfSlicesSumResp struct {
	Param0_ int
}

func (s _Irpc_sliceTestSliceOfSlicesSumResp) Serialize(e *irpc.Encoder) error {
	if err := e.Int(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'int': %w", err)
	}
	return nil
}
func (s *_Irpc_sliceTestSliceOfSlicesSumResp) Deserialize(d *irpc.Decoder) error {
	if err := d.Int(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'int': %w", err)
	}
	return nil
}

type _Irpc_sliceTestSliceOfBytesSumReq struct {
	Param0_slice []byte
}

func (s _Irpc_sliceTestSliceOfBytesSumReq) Serialize(e *irpc.Encoder) error {
	if err := e.ByteSlice(s.Param0_slice); err != nil {
		return fmt.Errorf("serialize s.Param0_slice of type '[]byte': %w", err)
	}
	return nil
}
func (s *_Irpc_sliceTestSliceOfBytesSumReq) Deserialize(d *irpc.Decoder) error {
	if err := d.ByteSlice(&s.Param0_slice); err != nil {
		return fmt.Errorf("deserialize s.Param0_slice of type '[]byte': %w", err)
	}
	return nil
}

type _Irpc_sliceTestSliceOfBytesSumResp struct {
	Param0_ int
}

func (s _Irpc_sliceTestSliceOfBytesSumResp) Serialize(e *irpc.Encoder) error {
	if err := e.Int(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'int': %w", err)
	}
	return nil
}
func (s *_Irpc_sliceTestSliceOfBytesSumResp) Deserialize(d *irpc.Decoder) error {
	if err := d.Int(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'int': %w", err)
	}
	return nil
}
