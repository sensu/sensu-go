package json

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
)

type JsonCodec struct{}

func (JsonCodec) Decode(data []byte, objPtr interface{}) error {
	if err := jsoniter.Unmarshal(data, objPtr); err != nil {
		fmt.Printf(err.Error())
		return err
	}
	return nil
}

func (JsonCodec) Encode(objPtr interface{}) ([]byte, error) {
	return jsoniter.Marshal(objPtr)
}
