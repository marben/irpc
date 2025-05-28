package irpcgen

import "context"

type Serializable interface {
	Serialize(e *Encoder) error
}

type Deserializable interface {
	Deserialize(d *Decoder) error
}

type FuncExecutor func(ctx context.Context) Serializable

type ArgDeserializer func(d *Decoder) (FuncExecutor, error)
