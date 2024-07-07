// Code generated by irpc generator; DO NOT EDIT
package irpctestpkg

import (
	"fmt"
	"github.com/marben/irpc/pkg/irpc"
)

type interfaceTestIRpcService struct {
	impl interfaceTest
}

func newInterfaceTestIRpcService(impl interfaceTest) *interfaceTestIRpcService {
	return &interfaceTestIRpcService{impl: impl}
}
func (interfaceTestIRpcService) Hash() []byte {
	return []byte("interfaceTestIRpcService")
}
func (s *interfaceTestIRpcService) GetFuncCall(funcId irpc.FuncId) (irpc.ArgDeserializer, error) {
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

type customInterfaceIRpcService struct {
	impl customInterface
}

func newCustomInterfaceIRpcService(impl customInterface) *customInterfaceIRpcService {
	return &customInterfaceIRpcService{impl: impl}
}
func (customInterfaceIRpcService) Hash() []byte {
	return []byte("customInterfaceIRpcService")
}
func (s *customInterfaceIRpcService) GetFuncCall(funcId irpc.FuncId) (irpc.ArgDeserializer, error) {
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

type interfaceTestIRpcClient struct {
	endpoint *irpc.Endpoint
	id       irpc.RegisteredServiceId
}

func newInterfaceTestIRpcClient(endpoint *irpc.Endpoint) (*interfaceTestIRpcClient, error) {
	id, err := endpoint.RegisterClient([]byte("interfaceTestIRpcService"))
	if err != nil {
		return nil, fmt.Errorf("register failed: %w", err)
	}
	return &interfaceTestIRpcClient{endpoint: endpoint, id: id}, nil
}
func (_c *interfaceTestIRpcClient) rtnErrorWithMessage(msg string) error {
	var req = _Irpc_interfaceTestrtnErrorWithMessageReq{
		Param0_msg: msg,
	}
	var resp _Irpc_interfaceTestrtnErrorWithMessageResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 0, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *interfaceTestIRpcClient) rtnNilError() error {
	var req = _Irpc_interfaceTestrtnNilErrorReq{}
	var resp _Irpc_interfaceTestrtnNilErrorResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 1, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *interfaceTestIRpcClient) rtnTwoErrors() (error, error) {
	var req = _Irpc_interfaceTestrtnTwoErrorsReq{}
	var resp _Irpc_interfaceTestrtnTwoErrorsResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 2, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_, resp.Param1_
}
func (_c *interfaceTestIRpcClient) rtnStringAndError(msg string) (s string, err error) {
	var req = _Irpc_interfaceTestrtnStringAndErrorReq{
		Param0_msg: msg,
	}
	var resp _Irpc_interfaceTestrtnStringAndErrorResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 3, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_s, resp.Param1_err
}
func (_c *interfaceTestIRpcClient) passCustomInterfaceAndReturnItModified(ci customInterface) (customInterface, error) {
	var req = _Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedReq{
		Param0_ci: ci,
	}
	var resp _Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 4, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_, resp.Param1_
}

type customInterfaceIRpcClient struct {
	endpoint *irpc.Endpoint
	id       irpc.RegisteredServiceId
}

func newCustomInterfaceIRpcClient(endpoint *irpc.Endpoint) (*customInterfaceIRpcClient, error) {
	id, err := endpoint.RegisterClient([]byte("customInterfaceIRpcService"))
	if err != nil {
		return nil, fmt.Errorf("register failed: %w", err)
	}
	return &customInterfaceIRpcClient{endpoint: endpoint, id: id}, nil
}
func (_c *customInterfaceIRpcClient) IntFunc() int {
	var req = _Irpc_customInterfaceIntFuncReq{}
	var resp _Irpc_customInterfaceIntFuncResp
	if err := _c.endpoint.CallRemoteFunc(_c.id, 0, req, &resp); err != nil {
		panic(err)
	}
	return resp.Param0_
}
func (_c *customInterfaceIRpcClient) StringFunc() string {
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

func (s _Irpc_interfaceTestrtnErrorWithMessageReq) Serialize(e *irpc.Encoder) error {
	if err := e.String(s.Param0_msg); err != nil {
		return fmt.Errorf("serialize s.Param0_msg of type 'string': %w", err)
	}
	return nil
}
func (s *_Irpc_interfaceTestrtnErrorWithMessageReq) Deserialize(d *irpc.Decoder) error {
	if err := d.String(&s.Param0_msg); err != nil {
		return fmt.Errorf("deserialize s.Param0_msg of type 'string': %w", err)
	}
	return nil
}

type _Irpc_interfaceTestrtnErrorWithMessageResp struct {
	Param0_ error
}

func (s _Irpc_interfaceTestrtnErrorWithMessageResp) Serialize(e *irpc.Encoder) error {
	{
		var isNil bool
		if s.Param0_ == nil {
			isNil = true
		}
		if err := e.Bool(isNil); err != nil {
			return fmt.Errorf("serialize isNil of type 'bool': %w", err)
		}

		if !isNil {
			{ // Error()
				_Error_0_ := s.Param0_.Error()
				if err := e.String(_Error_0_); err != nil {
					return fmt.Errorf("serialize _Error_0_ of type 'string': %w", err)
				}
			}
		}
	}
	return nil
}
func (s *_Irpc_interfaceTestrtnErrorWithMessageResp) Deserialize(d *irpc.Decoder) error {
	{
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

func (s _Irpc_interfaceTestrtnNilErrorReq) Serialize(e *irpc.Encoder) error {
	return nil
}
func (s *_Irpc_interfaceTestrtnNilErrorReq) Deserialize(d *irpc.Decoder) error {
	return nil
}

type _Irpc_interfaceTestrtnNilErrorResp struct {
	Param0_ error
}

func (s _Irpc_interfaceTestrtnNilErrorResp) Serialize(e *irpc.Encoder) error {
	{
		var isNil bool
		if s.Param0_ == nil {
			isNil = true
		}
		if err := e.Bool(isNil); err != nil {
			return fmt.Errorf("serialize isNil of type 'bool': %w", err)
		}

		if !isNil {
			{ // Error()
				_Error_0_ := s.Param0_.Error()
				if err := e.String(_Error_0_); err != nil {
					return fmt.Errorf("serialize _Error_0_ of type 'string': %w", err)
				}
			}
		}
	}
	return nil
}
func (s *_Irpc_interfaceTestrtnNilErrorResp) Deserialize(d *irpc.Decoder) error {
	{
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

func (s _Irpc_interfaceTestrtnTwoErrorsReq) Serialize(e *irpc.Encoder) error {
	return nil
}
func (s *_Irpc_interfaceTestrtnTwoErrorsReq) Deserialize(d *irpc.Decoder) error {
	return nil
}

type _Irpc_interfaceTestrtnTwoErrorsResp struct {
	Param0_ error
	Param1_ error
}

func (s _Irpc_interfaceTestrtnTwoErrorsResp) Serialize(e *irpc.Encoder) error {
	{
		var isNil bool
		if s.Param0_ == nil {
			isNil = true
		}
		if err := e.Bool(isNil); err != nil {
			return fmt.Errorf("serialize isNil of type 'bool': %w", err)
		}

		if !isNil {
			{ // Error()
				_Error_0_ := s.Param0_.Error()
				if err := e.String(_Error_0_); err != nil {
					return fmt.Errorf("serialize _Error_0_ of type 'string': %w", err)
				}
			}
		}
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
func (s *_Irpc_interfaceTestrtnTwoErrorsResp) Deserialize(d *irpc.Decoder) error {
	{
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
	{
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

func (s _Irpc_interfaceTestrtnStringAndErrorReq) Serialize(e *irpc.Encoder) error {
	if err := e.String(s.Param0_msg); err != nil {
		return fmt.Errorf("serialize s.Param0_msg of type 'string': %w", err)
	}
	return nil
}
func (s *_Irpc_interfaceTestrtnStringAndErrorReq) Deserialize(d *irpc.Decoder) error {
	if err := d.String(&s.Param0_msg); err != nil {
		return fmt.Errorf("deserialize s.Param0_msg of type 'string': %w", err)
	}
	return nil
}

type _Irpc_interfaceTestrtnStringAndErrorResp struct {
	Param0_s   string
	Param1_err error
}

func (s _Irpc_interfaceTestrtnStringAndErrorResp) Serialize(e *irpc.Encoder) error {
	if err := e.String(s.Param0_s); err != nil {
		return fmt.Errorf("serialize s.Param0_s of type 'string': %w", err)
	}
	{
		var isNil bool
		if s.Param1_err == nil {
			isNil = true
		}
		if err := e.Bool(isNil); err != nil {
			return fmt.Errorf("serialize isNil of type 'bool': %w", err)
		}

		if !isNil {
			{ // Error()
				_Error_0_ := s.Param1_err.Error()
				if err := e.String(_Error_0_); err != nil {
					return fmt.Errorf("serialize _Error_0_ of type 'string': %w", err)
				}
			}
		}
	}
	return nil
}
func (s *_Irpc_interfaceTestrtnStringAndErrorResp) Deserialize(d *irpc.Decoder) error {
	if err := d.String(&s.Param0_s); err != nil {
		return fmt.Errorf("deserialize s.Param0_s of type 'string': %w", err)
	}
	{
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

func (s _Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedReq) Serialize(e *irpc.Encoder) error {
	{
		var isNil bool
		if s.Param0_ci == nil {
			isNil = true
		}
		if err := e.Bool(isNil); err != nil {
			return fmt.Errorf("serialize isNil of type 'bool': %w", err)
		}

		if !isNil {
			{ // IntFunc()
				_IntFunc_0_ := s.Param0_ci.IntFunc()
				if err := e.Int(_IntFunc_0_); err != nil {
					return fmt.Errorf("serialize _IntFunc_0_ of type 'int': %w", err)
				}
			}
			{ // StringFunc()
				_StringFunc_0_ := s.Param0_ci.StringFunc()
				if err := e.String(_StringFunc_0_); err != nil {
					return fmt.Errorf("serialize _StringFunc_0_ of type 'string': %w", err)
				}
			}
		}
	}
	return nil
}
func (s *_Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedReq) Deserialize(d *irpc.Decoder) error {
	{
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

func (s _Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedResp) Serialize(e *irpc.Encoder) error {
	{
		var isNil bool
		if s.Param0_ == nil {
			isNil = true
		}
		if err := e.Bool(isNil); err != nil {
			return fmt.Errorf("serialize isNil of type 'bool': %w", err)
		}

		if !isNil {
			{ // IntFunc()
				_IntFunc_0_ := s.Param0_.IntFunc()
				if err := e.Int(_IntFunc_0_); err != nil {
					return fmt.Errorf("serialize _IntFunc_0_ of type 'int': %w", err)
				}
			}
			{ // StringFunc()
				_StringFunc_0_ := s.Param0_.StringFunc()
				if err := e.String(_StringFunc_0_); err != nil {
					return fmt.Errorf("serialize _StringFunc_0_ of type 'string': %w", err)
				}
			}
		}
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
func (s *_Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedResp) Deserialize(d *irpc.Decoder) error {
	{
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
	{
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

func (s _Irpc_customInterfaceIntFuncReq) Serialize(e *irpc.Encoder) error {
	return nil
}
func (s *_Irpc_customInterfaceIntFuncReq) Deserialize(d *irpc.Decoder) error {
	return nil
}

type _Irpc_customInterfaceIntFuncResp struct {
	Param0_ int
}

func (s _Irpc_customInterfaceIntFuncResp) Serialize(e *irpc.Encoder) error {
	if err := e.Int(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'int': %w", err)
	}
	return nil
}
func (s *_Irpc_customInterfaceIntFuncResp) Deserialize(d *irpc.Decoder) error {
	if err := d.Int(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'int': %w", err)
	}
	return nil
}

type _Irpc_customInterfaceStringFuncReq struct {
}

func (s _Irpc_customInterfaceStringFuncReq) Serialize(e *irpc.Encoder) error {
	return nil
}
func (s *_Irpc_customInterfaceStringFuncReq) Deserialize(d *irpc.Decoder) error {
	return nil
}

type _Irpc_customInterfaceStringFuncResp struct {
	Param0_ string
}

func (s _Irpc_customInterfaceStringFuncResp) Serialize(e *irpc.Encoder) error {
	if err := e.String(s.Param0_); err != nil {
		return fmt.Errorf("serialize s.Param0_ of type 'string': %w", err)
	}
	return nil
}
func (s *_Irpc_customInterfaceStringFuncResp) Deserialize(d *irpc.Decoder) error {
	if err := d.String(&s.Param0_); err != nil {
		return fmt.Errorf("deserialize s.Param0_ of type 'string': %w", err)
	}
	return nil
}
