package json

import (
	jsoniter "github.com/json-iterator/go"
)

type JsonCodec struct{}

func (JsonCodec) Decode(data []byte, objPtr interface{}) error {
	return jsoniter.Unmarshal(data, objPtr)
}

func (JsonCodec) Encode(objPtr interface{}) ([]byte, error) {
	return jsoniter.Marshal(objPtr)
}
