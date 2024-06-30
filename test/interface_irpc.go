// Code generated by irpc generator; DO NOT EDIT
package irpctestpkg

import (
	"encoding/binary"
	"fmt"
	"github.com/marben/irpc/pkg/irpc"
	"io"
)

type interfaceTestRpcService struct {
	impl interfaceTest
}

func newInterfaceTestRpcService(impl interfaceTest) *interfaceTestRpcService {
	return &interfaceTestRpcService{impl: impl}
}
func (interfaceTestRpcService) Hash() []byte {
	return []byte("interfaceTestRpcService")
}
func (s *interfaceTestRpcService) GetFuncCall(funcId irpc.FuncId) (irpc.ArgDeserializer, error) {
	switch funcId {
	case 0: // rtnErrorWithMessage
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_interfaceTestrtnErrorWithMessageReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_interfaceTestrtnErrorWithMessageResp
				resp.Param0_ = s.impl.rtnErrorWithMessage(args.Param0_msg)
				return resp
			}, nil
		}, nil
	case 1: // rtnNilError
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_interfaceTestrtnNilErrorReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_interfaceTestrtnNilErrorResp
				resp.Param0_ = s.impl.rtnNilError()
				return resp
			}, nil
		}, nil
	case 2: // rtnTwoErrors
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_interfaceTestrtnTwoErrorsReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_interfaceTestrtnTwoErrorsResp
				resp.Param0_, resp.Param1_ = s.impl.rtnTwoErrors()
				return resp
			}, nil
		}, nil
	case 3: // rtnStringAndError
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_interfaceTestrtnStringAndErrorReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_interfaceTestrtnStringAndErrorResp
				resp.Param0_s, resp.Param1_err = s.impl.rtnStringAndError(args.Param0_msg)
				return resp
			}, nil
		}, nil
	case 4: // passCustomInterfaceAndReturnItModified
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedResp
				resp.Param0_, resp.Param1_ = s.impl.passCustomInterfaceAndReturnItModified(args.Param0_ci)
				return resp
			}, nil
		}, nil
	default:
		return nil, fmt.Errorf("function '%d' doesn't exist on service '%s'", funcId, string(s.Hash()))
	}
}

type customInterfaceRpcService struct {
	impl customInterface
}

func newCustomInterfaceRpcService(impl customInterface) *customInterfaceRpcService {
	return &customInterfaceRpcService{impl: impl}
}
func (customInterfaceRpcService) Hash() []byte {
	return []byte("customInterfaceRpcService")
}
func (s *customInterfaceRpcService) GetFuncCall(funcId irpc.FuncId) (irpc.ArgDeserializer, error) {
	switch funcId {
	case 0: // IntFunc
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_customInterfaceIntFuncReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_customInterfaceIntFuncResp
				resp.Param0_ = s.impl.IntFunc()
				return resp
			}, nil
		}, nil
	case 1: // StringFunc
		return func(d *irpc.Decoder) (irpc.FuncExecutor, error) {
			// DESERIALIZE
			var args _Irpc_customInterfaceStringFuncReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func() irpc.Serializable {
				// EXECUTE
				var resp _Irpc_customInterfaceStringFuncResp
				resp.Param0_ = s.impl.StringFunc()
				return resp
			}, nil
		}, nil
	default:
		return nil, fmt.Errorf("function '%d' doesn't exist on service '%s'", funcId, string(s.Hash()))
	}
}

type interfaceTestRpcClient struct {
	endpoint *irpc.Endpoint
	id       irpc.RegisteredServiceId
}

func newInterfaceTestRpcClient(endpoint *irpc.Endpoint) (*interfaceTestRpcClient, error) {
	id, err := endpoint.RegisterClient([]byte("interfaceTestRpcService"))
	if err != nil {
		return nil, fmt.Errorf("register failed: %w", err)
	}
	return &interfaceTestRpcClient{endpoint: endpoint, id: id}, nil
}
func (_c *interfaceTestRpcClient) rtnErrorWithMessage(msg string) error {
	var req = _Irpc_interfaceTestrtnErrorWithMessageReq{
		Param0_msg: msg,
	}
	var resp _Irpc_interfaceTestrtnErrorWithMessageResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 0, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *interfaceTestRpcClient) rtnNilError() error {
	var req = _Irpc_interfaceTestrtnNilErrorReq{}
	var resp _Irpc_interfaceTestrtnNilErrorResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 1, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *interfaceTestRpcClient) rtnTwoErrors() (error, error) {
	var req = _Irpc_interfaceTestrtnTwoErrorsReq{}
	var resp _Irpc_interfaceTestrtnTwoErrorsResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 2, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_, resp.Param1_
}
func (_c *interfaceTestRpcClient) rtnStringAndError(msg string) (s string, err error) {
	var req = _Irpc_interfaceTestrtnStringAndErrorReq{
		Param0_msg: msg,
	}
	var resp _Irpc_interfaceTestrtnStringAndErrorResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 3, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_s, resp.Param1_err
}
func (_c *interfaceTestRpcClient) passCustomInterfaceAndReturnItModified(ci customInterface) (customInterface, error) {
	var req = _Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedReq{
		Param0_ci: ci,
	}
	var resp _Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 4, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_, resp.Param1_
}

type customInterfaceRpcClient struct {
	endpoint *irpc.Endpoint
	id       irpc.RegisteredServiceId
}

func newCustomInterfaceRpcClient(endpoint *irpc.Endpoint) (*customInterfaceRpcClient, error) {
	id, err := endpoint.RegisterClient([]byte("customInterfaceRpcService"))
	if err != nil {
		return nil, fmt.Errorf("register failed: %w", err)
	}
	return &customInterfaceRpcClient{endpoint: endpoint, id: id}, nil
}
func (_c *customInterfaceRpcClient) IntFunc() int {
	var req = _Irpc_customInterfaceIntFuncReq{}
	var resp _Irpc_customInterfaceIntFuncResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 0, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *customInterfaceRpcClient) StringFunc() string {
	var req = _Irpc_customInterfaceStringFuncReq{}
	var resp _Irpc_customInterfaceStringFuncResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 1, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}

type _Irpc_interfaceTestrtnErrorWithMessageReq struct {
	Param0_msg string
}

func (s _Irpc_interfaceTestrtnErrorWithMessageReq) Serialize(w io.Writer) error {
	b := make([]byte, 8)
	{ // string
		var l int = len(s.Param0_msg)
		binary.LittleEndian.PutUint64(b, uint64(l))
		if _, err := w.Write(b[:8]); err != nil {
			return fmt.Errorf("l int write: %w", err)
		}

		_, err := w.Write([]byte(s.Param0_msg))
		if err != nil {
			return fmt.Errorf("failed to write string to writer: %w", err)
		}
	}
	return nil
}
func (s *_Irpc_interfaceTestrtnErrorWithMessageReq) Deserialize(d *irpc.Decoder) error {
	{ // string
		if err := d.String(&s.Param0_msg); err != nil {
			return fmt.Errorf("deserialize s.Param0_msg of type 'string': %w", err)
		}
	}
	return nil
}

type _Irpc_interfaceTestrtnErrorWithMessageResp struct {
	Param0_ error
}

func (s _Irpc_interfaceTestrtnErrorWithMessageResp) Serialize(w io.Writer) error {
	b := make([]byte, 8)
	{ // error
		var isNil bool
		if s.Param0_ == nil {
			isNil = true
		}
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
				_Error_0_ := s.Param0_.Error()
				var l int = len(_Error_0_)
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
func (s *_Irpc_interfaceTestrtnErrorWithMessageResp) Deserialize(d *irpc.Decoder) error {
	{ // error
		var isNil bool
		if err := d.Bool(&isNil); err != nil {
			return fmt.Errorf("deserialize isNil of type 'bool': %w", err)
		}
		if isNil {
			s.Param0_ = nil
		} else {
			var impl _error_interfaceTest_irpcInterfaceImpl
			{ // Error()
				if err := d.String(&impl._Error_0_); err != nil {
					return fmt.Errorf("deserialize impl._Error_0_ of type 'string': %w", err)
				}
			}
			s.Param0_ = impl
		}
	}
	return nil
}

type _error_interfaceTest_irpcInterfaceImpl struct {
	_Error_0_ string
}

func (i _error_interfaceTest_irpcInterfaceImpl) Error() string {
	return i._Error_0_
}

type _Irpc_interfaceTestrtnNilErrorReq struct {
}

func (s _Irpc_interfaceTestrtnNilErrorReq) Serialize(w io.Writer) error {
	return nil
}
func (s *_Irpc_interfaceTestrtnNilErrorReq) Deserialize(d *irpc.Decoder) error {
	return nil
}

type _Irpc_interfaceTestrtnNilErrorResp struct {
	Param0_ error
}

func (s _Irpc_interfaceTestrtnNilErrorResp) Serialize(w io.Writer) error {
	b := make([]byte, 8)
	{ // error
		var isNil bool
		if s.Param0_ == nil {
			isNil = true
		}
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
				_Error_0_ := s.Param0_.Error()
				var l int = len(_Error_0_)
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
func (s *_Irpc_interfaceTestrtnNilErrorResp) Deserialize(d *irpc.Decoder) error {
	{ // error
		var isNil bool
		if err := d.Bool(&isNil); err != nil {
			return fmt.Errorf("deserialize isNil of type 'bool': %w", err)
		}
		if isNil {
			s.Param0_ = nil
		} else {
			var impl _error_interfaceTest_irpcInterfaceImpl
			{ // Error()
				if err := d.String(&impl._Error_0_); err != nil {
					return fmt.Errorf("deserialize impl._Error_0_ of type 'string': %w", err)
				}
			}
			s.Param0_ = impl
		}
	}
	return nil
}

type _Irpc_interfaceTestrtnTwoErrorsReq struct {
}

func (s _Irpc_interfaceTestrtnTwoErrorsReq) Serialize(w io.Writer) error {
	return nil
}
func (s *_Irpc_interfaceTestrtnTwoErrorsReq) Deserialize(d *irpc.Decoder) error {
	return nil
}

type _Irpc_interfaceTestrtnTwoErrorsResp struct {
	Param0_ error
	Param1_ error
}

func (s _Irpc_interfaceTestrtnTwoErrorsResp) Serialize(w io.Writer) error {
	b := make([]byte, 8)
	{ // error
		var isNil bool
		if s.Param0_ == nil {
			isNil = true
		}
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
				_Error_0_ := s.Param0_.Error()
				var l int = len(_Error_0_)
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
	{ // error
		var isNil bool
		if s.Param1_ == nil {
			isNil = true
		}
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
func (s *_Irpc_interfaceTestrtnTwoErrorsResp) Deserialize(d *irpc.Decoder) error {
	{ // error
		var isNil bool
		if err := d.Bool(&isNil); err != nil {
			return fmt.Errorf("deserialize isNil of type 'bool': %w", err)
		}
		if isNil {
			s.Param0_ = nil
		} else {
			var impl _error_interfaceTest_irpcInterfaceImpl
			{ // Error()
				if err := d.String(&impl._Error_0_); err != nil {
					return fmt.Errorf("deserialize impl._Error_0_ of type 'string': %w", err)
				}
			}
			s.Param0_ = impl
		}
	}
	{ // error
		var isNil bool
		if err := d.Bool(&isNil); err != nil {
			return fmt.Errorf("deserialize isNil of type 'bool': %w", err)
		}
		if isNil {
			s.Param1_ = nil
		} else {
			var impl _error_interfaceTest_irpcInterfaceImpl
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

type _Irpc_interfaceTestrtnStringAndErrorReq struct {
	Param0_msg string
}

func (s _Irpc_interfaceTestrtnStringAndErrorReq) Serialize(w io.Writer) error {
	b := make([]byte, 8)
	{ // string
		var l int = len(s.Param0_msg)
		binary.LittleEndian.PutUint64(b, uint64(l))
		if _, err := w.Write(b[:8]); err != nil {
			return fmt.Errorf("l int write: %w", err)
		}

		_, err := w.Write([]byte(s.Param0_msg))
		if err != nil {
			return fmt.Errorf("failed to write string to writer: %w", err)
		}
	}
	return nil
}
func (s *_Irpc_interfaceTestrtnStringAndErrorReq) Deserialize(d *irpc.Decoder) error {
	{ // string
		if err := d.String(&s.Param0_msg); err != nil {
			return fmt.Errorf("deserialize s.Param0_msg of type 'string': %w", err)
		}
	}
	return nil
}

type _Irpc_interfaceTestrtnStringAndErrorResp struct {
	Param0_s   string
	Param1_err error
}

func (s _Irpc_interfaceTestrtnStringAndErrorResp) Serialize(w io.Writer) error {
	b := make([]byte, 8)
	{ // string
		var l int = len(s.Param0_s)
		binary.LittleEndian.PutUint64(b, uint64(l))
		if _, err := w.Write(b[:8]); err != nil {
			return fmt.Errorf("l int write: %w", err)
		}

		_, err := w.Write([]byte(s.Param0_s))
		if err != nil {
			return fmt.Errorf("failed to write string to writer: %w", err)
		}
	}
	{ // error
		var isNil bool
		if s.Param1_err == nil {
			isNil = true
		}
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
				_Error_0_ := s.Param1_err.Error()
				var l int = len(_Error_0_)
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
func (s *_Irpc_interfaceTestrtnStringAndErrorResp) Deserialize(d *irpc.Decoder) error {
	{ // string
		if err := d.String(&s.Param0_s); err != nil {
			return fmt.Errorf("deserialize s.Param0_s of type 'string': %w", err)
		}
	}
	{ // error
		var isNil bool
		if err := d.Bool(&isNil); err != nil {
			return fmt.Errorf("deserialize isNil of type 'bool': %w", err)
		}
		if isNil {
			s.Param1_err = nil
		} else {
			var impl _error_interfaceTest_irpcInterfaceImpl
			{ // Error()
				if err := d.String(&impl._Error_0_); err != nil {
					return fmt.Errorf("deserialize impl._Error_0_ of type 'string': %w", err)
				}
			}
			s.Param1_err = impl
		}
	}
	return nil
}

type _Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedReq struct {
	Param0_ci customInterface
}

func (s _Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedReq) Serialize(w io.Writer) error {
	b := make([]byte, 8)
	{ // customInterface
		var isNil bool
		if s.Param0_ci == nil {
			isNil = true
		}
		if isNil {
			b[0] = 1
		} else {
			b[0] = 0
		}
		if _, err := w.Write(b[:1]); err != nil {
			return fmt.Errorf("isNil bool write: %w", err)
		}

		if !isNil {
			{ // IntFunc()
				_IntFunc_0_ := s.Param0_ci.IntFunc()
				binary.LittleEndian.PutUint64(b, uint64(_IntFunc_0_))
				if _, err := w.Write(b[:8]); err != nil {
					return fmt.Errorf("_IntFunc_0_ int write: %w", err)
				}
			}
			{ // StringFunc()
				_StringFunc_0_ := s.Param0_ci.StringFunc()
				var l int = len(_StringFunc_0_)
				binary.LittleEndian.PutUint64(b, uint64(l))
				if _, err := w.Write(b[:8]); err != nil {
					return fmt.Errorf("l int write: %w", err)
				}

				_, err := w.Write([]byte(_StringFunc_0_))
				if err != nil {
					return fmt.Errorf("failed to write string to writer: %w", err)
				}
			}
		}
	}
	return nil
}
func (s *_Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedReq) Deserialize(d *irpc.Decoder) error {
	{ // customInterface
		var isNil bool
		if err := d.Bool(&isNil); err != nil {
			return fmt.Errorf("deserialize isNil of type 'bool': %w", err)
		}
		if isNil {
			s.Param0_ci = nil
		} else {
			var impl _customInterface_interfaceTest_irpcInterfaceImpl
			{ // IntFunc()
				if err := d.Int(&impl._IntFunc_0_); err != nil {
					return fmt.Errorf("deserialize impl._IntFunc_0_ of type 'int': %w", err)
				}
			}
			{ // StringFunc()
				if err := d.String(&impl._StringFunc_0_); err != nil {
					return fmt.Errorf("deserialize impl._StringFunc_0_ of type 'string': %w", err)
				}
			}
			s.Param0_ci = impl
		}
	}
	return nil
}

type _customInterface_interfaceTest_irpcInterfaceImpl struct {
	_IntFunc_0_    int
	_StringFunc_0_ string
}

func (i _customInterface_interfaceTest_irpcInterfaceImpl) IntFunc() int {
	return i._IntFunc_0_
}
func (i _customInterface_interfaceTest_irpcInterfaceImpl) StringFunc() string {
	return i._StringFunc_0_
}

type _Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedResp struct {
	Param0_ customInterface
	Param1_ error
}

func (s _Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedResp) Serialize(w io.Writer) error {
	b := make([]byte, 8)
	{ // customInterface
		var isNil bool
		if s.Param0_ == nil {
			isNil = true
		}
		if isNil {
			b[0] = 1
		} else {
			b[0] = 0
		}
		if _, err := w.Write(b[:1]); err != nil {
			return fmt.Errorf("isNil bool write: %w", err)
		}

		if !isNil {
			{ // IntFunc()
				_IntFunc_0_ := s.Param0_.IntFunc()
				binary.LittleEndian.PutUint64(b, uint64(_IntFunc_0_))
				if _, err := w.Write(b[:8]); err != nil {
					return fmt.Errorf("_IntFunc_0_ int write: %w", err)
				}
			}
			{ // StringFunc()
				_StringFunc_0_ := s.Param0_.StringFunc()
				var l int = len(_StringFunc_0_)
				binary.LittleEndian.PutUint64(b, uint64(l))
				if _, err := w.Write(b[:8]); err != nil {
					return fmt.Errorf("l int write: %w", err)
				}

				_, err := w.Write([]byte(_StringFunc_0_))
				if err != nil {
					return fmt.Errorf("failed to write string to writer: %w", err)
				}
			}
		}
	}
	{ // error
		var isNil bool
		if s.Param1_ == nil {
			isNil = true
		}
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
func (s *_Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedResp) Deserialize(d *irpc.Decoder) error {
	{ // customInterface
		var isNil bool
		if err := d.Bool(&isNil); err != nil {
			return fmt.Errorf("deserialize isNil of type 'bool': %w", err)
		}
		if isNil {
			s.Param0_ = nil
		} else {
			var impl _customInterface_interfaceTest_irpcInterfaceImpl
			{ // IntFunc()
				if err := d.Int(&impl._IntFunc_0_); err != nil {
					return fmt.Errorf("deserialize impl._IntFunc_0_ of type 'int': %w", err)
				}
			}
			{ // StringFunc()
				if err := d.String(&impl._StringFunc_0_); err != nil {
					return fmt.Errorf("deserialize impl._StringFunc_0_ of type 'string': %w", err)
				}
			}
			s.Param0_ = impl
		}
	}
	{ // error
		var isNil bool
		if err := d.Bool(&isNil); err != nil {
			return fmt.Errorf("deserialize isNil of type 'bool': %w", err)
		}
		if isNil {
			s.Param1_ = nil
		} else {
			var impl _error_interfaceTest_irpcInterfaceImpl
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

type _Irpc_customInterfaceIntFuncReq struct {
}

func (s _Irpc_customInterfaceIntFuncReq) Serialize(w io.Writer) error {
	return nil
}
func (s *_Irpc_customInterfaceIntFuncReq) Deserialize(d *irpc.Decoder) error {
	return nil
}

type _Irpc_customInterfaceIntFuncResp struct {
	Param0_ int
}

func (s _Irpc_customInterfaceIntFuncResp) Serialize(w io.Writer) error {
	b := make([]byte, 8)
	{ // int
		binary.LittleEndian.PutUint64(b, uint64(s.Param0_))
		if _, err := w.Write(b[:8]); err != nil {
			return fmt.Errorf("s.Param0_ int write: %w", err)
		}
	}
	return nil
}
func (s *_Irpc_customInterfaceIntFuncResp) Deserialize(d *irpc.Decoder) error {
	{ // int
		if err := d.Int(&s.Param0_); err != nil {
			return fmt.Errorf("deserialize s.Param0_ of type 'int': %w", err)
		}
	}
	return nil
}

type _Irpc_customInterfaceStringFuncReq struct {
}

func (s _Irpc_customInterfaceStringFuncReq) Serialize(w io.Writer) error {
	return nil
}
func (s *_Irpc_customInterfaceStringFuncReq) Deserialize(d *irpc.Decoder) error {
	return nil
}

type _Irpc_customInterfaceStringFuncResp struct {
	Param0_ string
}

func (s _Irpc_customInterfaceStringFuncResp) Serialize(w io.Writer) error {
	b := make([]byte, 8)
	{ // string
		var l int = len(s.Param0_)
		binary.LittleEndian.PutUint64(b, uint64(l))
		if _, err := w.Write(b[:8]); err != nil {
			return fmt.Errorf("l int write: %w", err)
		}

		_, err := w.Write([]byte(s.Param0_))
		if err != nil {
			return fmt.Errorf("failed to write string to writer: %w", err)
		}
	}
	return nil
}
func (s *_Irpc_customInterfaceStringFuncResp) Deserialize(d *irpc.Decoder) error {
	{ // string
		if err := d.String(&s.Param0_); err != nil {
			return fmt.Errorf("deserialize s.Param0_ of type 'string': %w", err)
		}
	}
	return nil
}
