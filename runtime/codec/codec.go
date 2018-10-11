package codec

import (
	"github.com/sensu/sensu-go/runtime/codec/json"
	"github.com/sensu/sensu-go/runtime/codec/protobuf"
)

// A Codec is responsible for Encoding and Decoding data to
// a specific serialization format.
type Codec interface {
	Encode(objPtr interface{}) ([]byte, error)
	Decode(data []byte, objPtr interface{}) error
}

var (
	registeredCodecs = []Codec{
		json.JsonCodec{},
		protobuf.ProtobufCodec{},
	}
)

// UniversalCodec returns a Codec capable of serializing/deserializing
// all known serialization formats.
func UniversalCodec() Codec {
	return universalCodec{}
}

type universalCodec struct{}

func (universalCodec) Encode(objPtr interface{}) ([]byte, error) {
	var (
		bytes []byte
		err   error
	)

	for _, c := range registeredCodecs {
		bytes, err = c.Encode(objPtr)
		if err == nil {
			return bytes, nil
		}
	}

	return nil, err
}

func (universalCodec) Decode(data []byte, objPtr interface{}) error {
	var err error

	for _, c := range registeredCodecs {
		if err = c.Decode(data, objPtr); err == nil {
			return nil
		}
	}
	return err
}
