package wrap

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

//go:generate go run ../../../scripts/check_protoc/main.go
//go:generate go build -o $GOPATH/bin/protoc-gen-gofast github.com/gogo/protobuf/protoc-gen-gofast
//go:generate -command protoc protoc --plugin $GOPATH/bin/protoc-gen-gofast --gofast_out=plugins:$GOPATH/src -I=$GOPATH/pkg/mod -I=$GOPATH/pkg/mod/github.com/gogo/protobuf@v1.3.1/protobuf -I=$GOPATH/src
//go:generate protoc github.com/sensu/sensu-go/backend/store/wrap/wrapper.proto

// tmGetter is useful for types that want to explicitly provide their
// TypeMeta - this is useful for lifters.
type tmGetter interface {
	GetTypeMeta() corev2.TypeMeta
}

func (e Encoding) Encode(v interface{}) ([]byte, error) {
	switch e {
	case Encoding_json:
		return json.Marshal(v)
	case Encoding_protobuf:
		m, ok := v.(proto.Message)
		if !ok {
			return nil, fmt.Errorf("protobuf encoding requested, but %T is not a proto.Message", v)
		}
		return proto.Marshal(m)
	}
	return nil, fmt.Errorf("invalid encoding: %s", e)
}

func (e Encoding) Decode(m []byte, v interface{}) error {
	switch e {
	case Encoding_json:
		return json.Unmarshal(m, v)
	case Encoding_protobuf:
		msg, ok := v.(proto.Message)
		if !ok {
			return fmt.Errorf("protobuf decoding requested, but %T is not a proto.Message", v)
		}
		return proto.Unmarshal(m, msg)
	}
	return fmt.Errorf("invalid encoding: %s", e)
}

func (c Compression) Compress(m []byte) []byte {
	switch c {
	case Compression_none:
		return m
	case Compression_snappy:
		return snappy.Encode(nil, m)
	}
	return m
}

func (c Compression) Decompress(m []byte) ([]byte, error) {
	switch c {
	case Compression_none:
		return m, nil
	case Compression_snappy:
		return snappy.Decode(nil, m)
	}
	return nil, fmt.Errorf("invalid compression: %s", c)
}

// Option is a functional option, for passing to wrap.Resource().
type Option func(*Wrapper, corev3.Resource) error

// EncodeProtobuf is an option for setting protobuf encoding. If the resource
// is not a proto.Message, an error will be returned.
var EncodeProtobuf Option = func(w *Wrapper, r corev3.Resource) error {
	if _, ok := r.(proto.Message); !ok {
		return fmt.Errorf("protobuf encoding requested, but %T is not a proto.Message", r)
	}
	w.Encoding = Encoding_protobuf
	return nil
}

// EncodeJSON is an option for setting JSON encoding.
var EncodeJSON Option = func(w *Wrapper, r corev3.Resource) error {
	w.Encoding = Encoding_json
	return nil
}

// EncodeDefault is the default encoder. It will be protobuf, unless the
// resource cannot be type asserted to proto.Message.
var EncodeDefault Option = func(w *Wrapper, r corev3.Resource) error {
	encoding := Encoding_json
	if _, ok := r.(proto.Message); ok {
		encoding = Encoding_protobuf
	}
	w.Encoding = encoding
	return nil
}

// CompressNone is an option for turning off compression.
var CompressNone Option = func(w *Wrapper, r corev3.Resource) error {
	w.Compression = Compression_none
	return nil
}

// CompressSnappy is an option for setting snappy compression.
var CompressSnappy Option = func(w *Wrapper, r corev3.Resource) error {
	w.Compression = Compression_snappy
	return nil
}

// CompressDefault is the default compression algorithm.
var CompressDefault = CompressSnappy

// Resource wraps the given resource in a wrapper designed for storage.
// By default, EncodeDefault and CompressDefault options are used. They can
// be overridden by supplying other options. Typically, protobuf-capable
// resources will be marshalled to protobuf and then compressed with snappy.
func Resource(r corev3.Resource, opts ...Option) (*Wrapper, error) {
	var tm corev2.TypeMeta
	if getter, ok := r.(tmGetter); ok {
		tm = getter.GetTypeMeta()
	} else {
		typ := reflect.Indirect(reflect.ValueOf(r)).Type()
		tm = corev2.TypeMeta{
			Type:       typ.Name(),
			APIVersion: types.ApiVersion(typ.PkgPath()),
		}
	}
	var w Wrapper
	w.Metadata = &tm
	opts = append([]Option{EncodeDefault, CompressDefault}, opts...)
	for _, opt := range opts {
		if err := opt(&w, r); err != nil {
			return nil, err
		}
	}

	message, err := w.Encoding.Encode(r)
	if err != nil {
		return nil, err
	}

	w.Value = w.Compression.Compress(message)

	return &w, nil
}

// Unwrap unmarshals the wrapper's value into a resource, according to the
// configuration of the wrapper.
func (w *Wrapper) Unwrap() (corev3.Resource, error) {
	r, err := types.ResolveType(w.Metadata.APIVersion, w.Metadata.Type)
	if err != nil {
		return nil, err
	}
	proxy, ok := r.(*corev3.V2ResourceProxy)
	if !ok {
		return nil, errors.New("only v3 resources are compatible with store wrappers")
	}
	resource := proxy.Resource
	message, err := w.Compression.Decompress(w.Value)
	if err != nil {
		return nil, fmt.Errorf("error unwrapping %T: %s", resource, err)
	}
	return resource, w.Encoding.Decode(message, &resource)
}
