package irpc

import (
	"context"
	"fmt"
)

/*
 * clientRegisterService is registered upon endpoint startup as serviceId 0
 * it is used to negotiate other services registrations
 * we don't version this service atm and it is not autogenerated
 */

var _ Service = &clientRegisterService{}

// service accomodating the client's registration
// only for endpoint's purposes
type clientRegisterService struct {
	ep *Endpoint
}

type clientRegisterReq struct {
	ServiceHash []byte
}

type clientRegisterResp struct {
	ServiceId RegisteredServiceId
	Err       string // todo: turn into proper error
}

// GetFuncCall implements Service.
func (c *clientRegisterService) GetFuncCall(funcId FuncId) (ArgDeserializer, error) {
	switch funcId {
	case 0:
		return func(d *Decoder) (FuncExecutor, error) {
			var args clientRegisterReq
			if err := args.Deserialize(d); err != nil {
				return nil, err
			}
			return func(ctx context.Context) Serializable {
				c.ep.m.Lock()
				defer c.ep.m.Unlock()

				var resp clientRegisterResp
				serviceId, found := c.ep.serviceHashes[string(args.ServiceHash)]
				if !found {
					resp.Err = fmt.Errorf("couldn't find service hash %q", string(args.ServiceHash)).Error()
				} else {
					resp.ServiceId = serviceId
				}

				return resp
			}, nil
		}, nil
	default:
		return nil, fmt.Errorf("registerService has no function %d", funcId)
	}
}

// Hash implements Service.
func (c *clientRegisterService) Hash() []byte {
	// currently, client register service is not versioned and it's hash is not used at all
	return nil
}

func (rp clientRegisterReq) Serialize(e *Encoder) error {
	if err := e.ByteSlice(rp.ServiceHash); err != nil {
		return err
	}

	return nil
}

func (rp *clientRegisterReq) Deserialize(d *Decoder) error {
	if err := d.ByteSlice(&rp.ServiceHash); err != nil {
		return err
	}

	return nil
}

func (rp clientRegisterResp) Serialize(e *Encoder) error {
	if err := e.Uint16(uint16(rp.ServiceId)); err != nil {
		return err
	}
	if err := e.String(rp.Err); err != nil {
		return err
	}
	return nil
}

func (rp *clientRegisterResp) Deserialize(d *Decoder) error {
	if err := d.Uint16((*uint16)(&rp.ServiceId)); err != nil {
		return err
	}
	if err := d.String(&rp.Err); err != nil {
		return err
	}
	return nil
}
