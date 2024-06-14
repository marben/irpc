// Code generated by irpc generator; DO NOT EDIT
package irpctestpkg

import (
	"bytes"
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
func (s *interfaceTestRpcService) CallFunc(funcId irpc.FuncId, args []byte) ([]byte, error) {
	switch funcId {
	case 0:
		return s.callrtnErrorWithMessage(args)
	case 1:
		return s.callrtnNilError(args)
	case 2:
		return s.callrtnTwoErrors(args)
	case 3:
		return s.callrtnStringAndError(args)
	case 4:
		return s.callpassCustomInterfaceAndReturnItModified(args)
	default:
		return nil, fmt.Errorf("function '%d' doesn't exist on service '%s'", funcId, string(s.Hash()))
	}
}
func (s *interfaceTestRpcService) callrtnErrorWithMessage(params []byte) ([]byte, error) {
	r := bytes.NewBuffer(params)
	var req _Irpc_interfaceTestrtnErrorWithMessageReq
	if err := req.Deserialize(r); err != nil {
		return nil, fmt.Errorf("failed to deserialize rtnErrorWithMessage: %w", err)
	}
	var resp _Irpc_interfaceTestrtnErrorWithMessageResp
	resp.Param0_ = s.impl.rtnErrorWithMessage(req.Param0_msg)
	b := bytes.NewBuffer(nil)
	err := resp.Serialize(b)
	if err != nil {
		return nil, fmt.Errorf("response serialization failed: %w", err)
	}
	return b.Bytes(), nil
}
func (s *interfaceTestRpcService) callrtnNilError(params []byte) ([]byte, error) {
	r := bytes.NewBuffer(params)
	var req _Irpc_interfaceTestrtnNilErrorReq
	if err := req.Deserialize(r); err != nil {
		return nil, fmt.Errorf("failed to deserialize rtnNilError: %w", err)
	}
	var resp _Irpc_interfaceTestrtnNilErrorResp
	resp.Param0_ = s.impl.rtnNilError()
	b := bytes.NewBuffer(nil)
	err := resp.Serialize(b)
	if err != nil {
		return nil, fmt.Errorf("response serialization failed: %w", err)
	}
	return b.Bytes(), nil
}
func (s *interfaceTestRpcService) callrtnTwoErrors(params []byte) ([]byte, error) {
	r := bytes.NewBuffer(params)
	var req _Irpc_interfaceTestrtnTwoErrorsReq
	if err := req.Deserialize(r); err != nil {
		return nil, fmt.Errorf("failed to deserialize rtnTwoErrors: %w", err)
	}
	var resp _Irpc_interfaceTestrtnTwoErrorsResp
	resp.Param0_, resp.Param1_ = s.impl.rtnTwoErrors()
	b := bytes.NewBuffer(nil)
	err := resp.Serialize(b)
	if err != nil {
		return nil, fmt.Errorf("response serialization failed: %w", err)
	}
	return b.Bytes(), nil
}
func (s *interfaceTestRpcService) callrtnStringAndError(params []byte) ([]byte, error) {
	r := bytes.NewBuffer(params)
	var req _Irpc_interfaceTestrtnStringAndErrorReq
	if err := req.Deserialize(r); err != nil {
		return nil, fmt.Errorf("failed to deserialize rtnStringAndError: %w", err)
	}
	var resp _Irpc_interfaceTestrtnStringAndErrorResp
	resp.Param0_s, resp.Param1_err = s.impl.rtnStringAndError(req.Param0_msg)
	b := bytes.NewBuffer(nil)
	err := resp.Serialize(b)
	if err != nil {
		return nil, fmt.Errorf("response serialization failed: %w", err)
	}
	return b.Bytes(), nil
}
func (s *interfaceTestRpcService) callpassCustomInterfaceAndReturnItModified(params []byte) ([]byte, error) {
	r := bytes.NewBuffer(params)
	var req _Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedReq
	if err := req.Deserialize(r); err != nil {
		return nil, fmt.Errorf("failed to deserialize passCustomInterfaceAndReturnItModified: %w", err)
	}
	var resp _Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedResp
	resp.Param0_, resp.Param1_ = s.impl.passCustomInterfaceAndReturnItModified(req.Param0_ci)
	b := bytes.NewBuffer(nil)
	err := resp.Serialize(b)
	if err != nil {
		return nil, fmt.Errorf("response serialization failed: %w", err)
	}
	return b.Bytes(), nil
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
func (s *customInterfaceRpcService) CallFunc(funcId irpc.FuncId, args []byte) ([]byte, error) {
	switch funcId {
	case 0:
		return s.callIntFunc(args)
	case 1:
		return s.callStringFunc(args)
	default:
		return nil, fmt.Errorf("function '%d' doesn't exist on service '%s'", funcId, string(s.Hash()))
	}
}
func (s *customInterfaceRpcService) callIntFunc(params []byte) ([]byte, error) {
	r := bytes.NewBuffer(params)
	var req _Irpc_customInterfaceIntFuncReq
	if err := req.Deserialize(r); err != nil {
		return nil, fmt.Errorf("failed to deserialize IntFunc: %w", err)
	}
	var resp _Irpc_customInterfaceIntFuncResp
	resp.Param0_ = s.impl.IntFunc()
	b := bytes.NewBuffer(nil)
	err := resp.Serialize(b)
	if err != nil {
		return nil, fmt.Errorf("response serialization failed: %w", err)
	}
	return b.Bytes(), nil
}
func (s *customInterfaceRpcService) callStringFunc(params []byte) ([]byte, error) {
	r := bytes.NewBuffer(params)
	var req _Irpc_customInterfaceStringFuncReq
	if err := req.Deserialize(r); err != nil {
		return nil, fmt.Errorf("failed to deserialize StringFunc: %w", err)
	}
	var resp _Irpc_customInterfaceStringFuncResp
	resp.Param0_ = s.impl.StringFunc()
	b := bytes.NewBuffer(nil)
	err := resp.Serialize(b)
	if err != nil {
		return nil, fmt.Errorf("response serialization failed: %w", err)
	}
	return b.Bytes(), nil
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
	{ // string
		var l int = len(s.Param0_msg)
		b := make([]byte, 8)
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
func (s *_Irpc_interfaceTestrtnErrorWithMessageReq) Deserialize(r io.Reader) error {
	{ // string
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
			s.Param0_msg = string(sbuf)
		}
	}
	return nil
}

type _Irpc_interfaceTestrtnErrorWithMessageResp struct {
	Param0_ error
}

func (s _Irpc_interfaceTestrtnErrorWithMessageResp) Serialize(w io.Writer) error {
	{ // error
		var isNil bool
		if s.Param0_ == nil {
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
				_Error_0_ := s.Param0_.Error()
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
func (s *_Irpc_interfaceTestrtnErrorWithMessageResp) Deserialize(r io.Reader) error {
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
			s.Param0_ = nil
		} else {
			var impl _error_interfaceTest_irpcInterfaceImpl
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
func (s *_Irpc_interfaceTestrtnNilErrorReq) Deserialize(r io.Reader) error {
	return nil
}

type _Irpc_interfaceTestrtnNilErrorResp struct {
	Param0_ error
}

func (s _Irpc_interfaceTestrtnNilErrorResp) Serialize(w io.Writer) error {
	{ // error
		var isNil bool
		if s.Param0_ == nil {
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
				_Error_0_ := s.Param0_.Error()
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
func (s *_Irpc_interfaceTestrtnNilErrorResp) Deserialize(r io.Reader) error {
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
			s.Param0_ = nil
		} else {
			var impl _error_interfaceTest_irpcInterfaceImpl
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
func (s *_Irpc_interfaceTestrtnTwoErrorsReq) Deserialize(r io.Reader) error {
	return nil
}

type _Irpc_interfaceTestrtnTwoErrorsResp struct {
	Param0_ error
	Param1_ error
}

func (s _Irpc_interfaceTestrtnTwoErrorsResp) Serialize(w io.Writer) error {
	{ // error
		var isNil bool
		if s.Param0_ == nil {
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
				_Error_0_ := s.Param0_.Error()
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
func (s *_Irpc_interfaceTestrtnTwoErrorsResp) Deserialize(r io.Reader) error {
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
			s.Param0_ = nil
		} else {
			var impl _error_interfaceTest_irpcInterfaceImpl
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
			s.Param0_ = impl
		}
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
			var impl _error_interfaceTest_irpcInterfaceImpl
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

type _Irpc_interfaceTestrtnStringAndErrorReq struct {
	Param0_msg string
}

func (s _Irpc_interfaceTestrtnStringAndErrorReq) Serialize(w io.Writer) error {
	{ // string
		var l int = len(s.Param0_msg)
		b := make([]byte, 8)
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
func (s *_Irpc_interfaceTestrtnStringAndErrorReq) Deserialize(r io.Reader) error {
	{ // string
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
			s.Param0_msg = string(sbuf)
		}
	}
	return nil
}

type _Irpc_interfaceTestrtnStringAndErrorResp struct {
	Param0_s   string
	Param1_err error
}

func (s _Irpc_interfaceTestrtnStringAndErrorResp) Serialize(w io.Writer) error {
	{ // string
		var l int = len(s.Param0_s)
		b := make([]byte, 8)
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
				_Error_0_ := s.Param1_err.Error()
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
func (s *_Irpc_interfaceTestrtnStringAndErrorResp) Deserialize(r io.Reader) error {
	{ // string
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
			s.Param0_s = string(sbuf)
		}
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
			s.Param1_err = nil
		} else {
			var impl _error_interfaceTest_irpcInterfaceImpl
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
			s.Param1_err = impl
		}
	}
	return nil
}

type _Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedReq struct {
	Param0_ci customInterface
}

func (s _Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedReq) Serialize(w io.Writer) error {
	{ // customInterface
		var isNil bool
		if s.Param0_ci == nil {
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
			{ // IntFunc()
				_IntFunc_0_ := s.Param0_ci.IntFunc()
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, uint64(_IntFunc_0_))
				if _, err := w.Write(b[:8]); err != nil {
					return fmt.Errorf("_IntFunc_0_ int write: %w", err)
				}
			}
			{ // StringFunc()
				_StringFunc_0_ := s.Param0_ci.StringFunc()
				var l int = len(_StringFunc_0_)
				b := make([]byte, 8)
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
func (s *_Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedReq) Deserialize(r io.Reader) error {
	{ // customInterface
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
			s.Param0_ci = nil
		} else {
			var impl _customInterface_interfaceTest_irpcInterfaceImpl
			{ // IntFunc()
				b := make([]byte, 8)
				if _, err := io.ReadFull(r, b[:8]); err != nil {
					return fmt.Errorf("impl._IntFunc_0_ int decode: %w", err)
				}
				impl._IntFunc_0_ = int(binary.LittleEndian.Uint64(b))
			}
			{ // StringFunc()
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
					impl._StringFunc_0_ = string(sbuf)
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
	{ // customInterface
		var isNil bool
		if s.Param0_ == nil {
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
			{ // IntFunc()
				_IntFunc_0_ := s.Param0_.IntFunc()
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, uint64(_IntFunc_0_))
				if _, err := w.Write(b[:8]); err != nil {
					return fmt.Errorf("_IntFunc_0_ int write: %w", err)
				}
			}
			{ // StringFunc()
				_StringFunc_0_ := s.Param0_.StringFunc()
				var l int = len(_StringFunc_0_)
				b := make([]byte, 8)
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
func (s *_Irpc_interfaceTestpassCustomInterfaceAndReturnItModifiedResp) Deserialize(r io.Reader) error {
	{ // customInterface
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
			s.Param0_ = nil
		} else {
			var impl _customInterface_interfaceTest_irpcInterfaceImpl
			{ // IntFunc()
				b := make([]byte, 8)
				if _, err := io.ReadFull(r, b[:8]); err != nil {
					return fmt.Errorf("impl._IntFunc_0_ int decode: %w", err)
				}
				impl._IntFunc_0_ = int(binary.LittleEndian.Uint64(b))
			}
			{ // StringFunc()
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
					impl._StringFunc_0_ = string(sbuf)
				}
			}
			s.Param0_ = impl
		}
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
			var impl _error_interfaceTest_irpcInterfaceImpl
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

type _Irpc_customInterfaceIntFuncReq struct {
}

func (s _Irpc_customInterfaceIntFuncReq) Serialize(w io.Writer) error {
	return nil
}
func (s *_Irpc_customInterfaceIntFuncReq) Deserialize(r io.Reader) error {
	return nil
}

type _Irpc_customInterfaceIntFuncResp struct {
	Param0_ int
}

func (s _Irpc_customInterfaceIntFuncResp) Serialize(w io.Writer) error {
	{ // int
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(s.Param0_))
		if _, err := w.Write(b[:8]); err != nil {
			return fmt.Errorf("s.Param0_ int write: %w", err)
		}
	}
	return nil
}
func (s *_Irpc_customInterfaceIntFuncResp) Deserialize(r io.Reader) error {
	{ // int
		b := make([]byte, 8)
		if _, err := io.ReadFull(r, b[:8]); err != nil {
			return fmt.Errorf("s.Param0_ int decode: %w", err)
		}
		s.Param0_ = int(binary.LittleEndian.Uint64(b))
	}
	return nil
}

type _Irpc_customInterfaceStringFuncReq struct {
}

func (s _Irpc_customInterfaceStringFuncReq) Serialize(w io.Writer) error {
	return nil
}
func (s *_Irpc_customInterfaceStringFuncReq) Deserialize(r io.Reader) error {
	return nil
}

type _Irpc_customInterfaceStringFuncResp struct {
	Param0_ string
}

func (s _Irpc_customInterfaceStringFuncResp) Serialize(w io.Writer) error {
	{ // string
		var l int = len(s.Param0_)
		b := make([]byte, 8)
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
func (s *_Irpc_customInterfaceStringFuncResp) Deserialize(r io.Reader) error {
	{ // string
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
			s.Param0_ = string(sbuf)
		}
	}
	return nil
}
