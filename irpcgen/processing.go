package irpcgen

import (
	"context"
	"fmt"
)

type FuncId uint64

type ServiceId uint64

func (pt ServiceId) Serialize(e *Encoder) error {
	return e.uint64le(uint64(pt))
}
func (pt *ServiceId) Deserialize(d *Decoder) error {
	val, err := d.uint64le()
	if err != nil {
		return err
	}
	*pt = ServiceId(val)
	return nil
}

func (pt ServiceId) String() string {
	return fmt.Sprintf("%#016x", uint64(pt))
}

// Service defines a collection of RPC-callable functions.
type Service interface {
	// Id returns the unique identifier of the service.
	Id() ServiceId

	// GetFuncCall returns a deserializer for the given function ID.
	GetFuncCall(funcId FuncId) (ArgDeserializer, error)
}

// Serializable serializes its state using an Encoder
type Serializable interface {
	Serialize(e *Encoder) error
}

// Deserializable initializes its state from a Decoder.
type Deserializable interface {
	Deserialize(d *Decoder) error
}

// FuncExecutor wraps a function call and returns its result as Serializable.
type FuncExecutor func(ctx context.Context) Serializable

// ArgDeserializer deserializes function arguments from a Decoder and returns
// a FuncExecutor that performs the function call.
type ArgDeserializer func(d *Decoder) (FuncExecutor, error)

// EmptySerializable is a no-op implementation of [Serializable].
type EmptySerializable struct{}

func (EmptySerializable) Serialize(e *Encoder) error { return nil }

// EmptyDeserializable is a no-op implementation of [Deserializable].
type EmptyDeserializable struct{}

func (EmptyDeserializable) Deserialize(d *Decoder) error { return nil }
