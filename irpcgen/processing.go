package irpcgen

import "context"

type FuncId uint64

type Service interface {
	Id() []byte // unique id of the service
	GetFuncCall(funcId FuncId) (ArgDeserializer, error)
}

type Serializable interface {
	Serialize(e *Encoder) error
}

type Deserializable interface {
	Deserialize(d *Decoder) error
}

type FuncExecutor func(ctx context.Context) Serializable

type ArgDeserializer func(d *Decoder) (FuncExecutor, error)

type EmptySerializable struct{}

func (EmptySerializable) Serialize(e *Encoder) error { return nil }

type EmptyDeserializable struct{}

func (EmptyDeserializable) Deserialize(d *Decoder) error { return nil }
