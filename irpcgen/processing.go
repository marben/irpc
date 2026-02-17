package irpcgen

import "context"

type FuncId uint64

// Service defines a collection of RPC-callable functions.
type Service interface {
	// Id returns the unique identifier of the service.
	Id() []byte

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
