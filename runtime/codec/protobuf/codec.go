package protobuf

import (
	"errors"

	"github.com/gogo/protobuf/proto"
)

type ProtobufCodec struct{}

func (ProtobufCodec) Encode(objPtr interface{}) ([]byte, error) {
	if msg, ok := objPtr.(proto.Message); ok {
		return proto.Marshal(msg)
	}

	return nil, errors.New("cannot serialize object to protobuf")
}

func (ProtobufCodec) Decode(data []byte, objPtr interface{}) error {
	if msg, ok := objPtr.(proto.Message); ok {
		return proto.Unmarshal(data, msg)
	}

	return errors.New("cannot deserialize protobuf into object")
}
