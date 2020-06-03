package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/types"
)

//go:generate go run ../../scripts/check_protoc/main.go
//go:generate go build -o $GOPATH/bin/protoc-gen-gofast github.com/gogo/protobuf/protoc-gen-gofast
//go:generate -command protoc protoc --plugin $GOPATH/bin/protoc-gen-gofast --gofast_out=plugins:$GOPATH/src -I=$GOPATH/pkg/mod -I=$GOPATH/pkg/mod/github.com/gogo/protobuf@v1.3.1/protobuf -I=$GOPATH/src
//go:generate protoc github.com/sensu/sensu-go/backend/store/wrapper.proto

// tmGetter is useful for types that want to explicitly provide their
// TypeMeta - this is useful for lifters.
type tmGetter interface {
	GetTypeMeta() corev2.TypeMeta
}

func handleProtoMarshal(r corev3.Resource) ([]byte, error) {
	if value, ok := r.(proto.Message); ok {
		return proto.Marshal(value)
	}
	return nil, fmt.Errorf("%T is not a proto.Message", r)
}

func handleJSONMarshal(r corev3.Resource) ([]byte, error) {
	return json.Marshal(r)
}

func handleUndefinedMarshal(corev3.Resource) ([]byte, error) {
	return nil, errors.New("undefined encoding")
}

var encoders = map[corev2.TypeMeta_Encoding]func(corev3.Resource) ([]byte, error){
	corev2.TypeMeta_undefined: handleUndefinedMarshal,
	corev2.TypeMeta_json:      handleJSONMarshal,
	corev2.TypeMeta_protobuf:  handleProtoMarshal,
}

func handleCompressionNone(in []byte) ([]byte, error) {
	return in, nil
}

func handleCompressionSnappy(in []byte) ([]byte, error) {
	return snappy.Encode(nil, in), nil
}

var compressors = map[corev2.TypeMeta_Compression]func([]byte) ([]byte, error){
	corev2.TypeMeta_none:   handleCompressionNone,
	corev2.TypeMeta_snappy: handleCompressionSnappy,
}

// WrapResource wraps the given resource in a wrapper designed for storage
func WrapResource(r corev3.Resource) (*Wrapper, error) {
	var tm corev2.TypeMeta
	if getter, ok := r.(tmGetter); ok {
		tm = getter.GetTypeMeta()
	} else {
		typ := reflect.Indirect(reflect.ValueOf(r)).Type()
		encoding := corev2.TypeMeta_json
		if _, ok := r.(proto.Message); ok {
			encoding = corev2.TypeMeta_protobuf
		}
		tm = corev2.TypeMeta{
			Type:       typ.Name(),
			APIVersion: types.ApiVersion(typ.PkgPath()),
			Encoding:   encoding,
		}
	}

	encoder, ok := encoders[tm.Encoding]
	if !ok {
		return nil, fmt.Errorf("invalid encoding: %v", tm.Encoding)
	}
	message, err := encoder(r)
	if err != nil {
		return nil, err
	}

	compressor, ok := compressors[tm.Compression]
	if !ok {
		return nil, fmt.Errorf("invalid compression: %v", tm.Compression)
	}

	message, err = compressor(message)
	if err != nil {
		return nil, err
	}

	return &Wrapper{
		Metadata: &tm,
		Value:    message,
	}, nil
}

func handleUndefinedUnmarshal([]byte, interface{}) error {
	return errors.New("undefined decoder")
}

func handleProtoUnmarshal(message []byte, value interface{}) error {
	v, ok := value.(proto.Message)
	if !ok {
		return fmt.Errorf("%T is not a proto.Message", value)
	}
	return proto.Unmarshal(message, v)
}

var decoders = map[corev2.TypeMeta_Encoding]func([]byte, interface{}) error{
	corev2.TypeMeta_undefined: handleUndefinedUnmarshal,
	corev2.TypeMeta_json:      json.Unmarshal,
	corev2.TypeMeta_protobuf:  handleProtoUnmarshal,
}

func handleDecompressionNone(in []byte) ([]byte, error) {
	return in, nil
}

func handleDecompressionSnappy(in []byte) ([]byte, error) {
	return snappy.Decode(nil, in)
}

var decompressors = map[corev2.TypeMeta_Compression]func([]byte) ([]byte, error){
	corev2.TypeMeta_none:   handleDecompressionNone,
	corev2.TypeMeta_snappy: handleDecompressionSnappy,
}

func (w *Wrapper) Unwrap() (corev3.Resource, error) {
	resource, err := corev3.ResolveResource(w.Metadata.Type)
	if err != nil {
		return nil, err
	}
	decompressor, ok := decompressors[w.Metadata.Compression]
	if !ok {
		return nil, fmt.Errorf("invalid compression: %v", w.Metadata.Compression)
	}
	message, err := decompressor(w.Value)
	if err != nil {
		return nil, fmt.Errorf("error unwrapping %T: %s", resource, err)
	}
	decoder, ok := decoders[w.Metadata.Encoding]
	if !ok {
		return nil, fmt.Errorf("invalid encoding: %v", w.Metadata.Encoding)
	}
	return resource, decoder(message, &resource)
}
