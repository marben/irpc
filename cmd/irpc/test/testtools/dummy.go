package testtools

import (
	"github.com/marben/irpc/irpcgen"
)

var _ irpcgen.Serializable = DummySerializable{}

type DummySerializable struct{}

// Serialize implements irpcgen.Serializable.
func (d DummySerializable) Serialize(e *irpcgen.Encoder) error {
	return nil
}

var _ irpcgen.Deserializable = DummyDeserializable{}

type DummyDeserializable struct{}

// Deserialize implements irpcgen.Deserializable.
func (DummyDeserializable) Deserialize(d *irpcgen.Decoder) error {
	return nil
}
