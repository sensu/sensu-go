package wrap

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	//nolint:staticcheck // SA1004 Replacing this will take some planning.
	"github.com/golang/protobuf/proto"

	"github.com/golang/snappy"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

//go:generate go run ../../../../scripts/check_protoc/main.go
//go:generate go build -o $GOPATH/bin/protoc-gen-gofast github.com/gogo/protobuf/protoc-gen-gofast
//go:generate -command protoc protoc --plugin $GOPATH/bin/protoc-gen-gofast --gofast_out=plugins:$GOPATH/src -I=$GOPATH/pkg/mod -I=$GOPATH/pkg/mod/github.com/gogo/protobuf@v1.3.1/protobuf -I=$GOPATH/src
//go:generate protoc github.com/sensu/sensu-go/backend/store/v2/wrap/wrapper.proto

// tmGetter is useful for types that want to explicitly provide their
// TypeMeta - this is useful for lifters.
type tmGetter interface {
	GetTypeMeta() corev2.TypeMeta
}

type validatable interface {
	Validate() error
}

var ErrValidateMethodMissing = errors.New("resource is missing required Validate() method")

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
type Option func(wrapper *Wrapper, resource interface{}) error

// EncodeProtobuf is an option for setting protobuf encoding. If the resource
// is not a proto.Message, an error will be returned.
var EncodeProtobuf Option = func(w *Wrapper, r interface{}) error {
	if _, ok := r.(proto.Message); !ok {
		return fmt.Errorf("protobuf encoding requested, but %T is not a proto.Message", r)
	}
	w.Encoding = Encoding_protobuf
	return nil
}

// EncodeJSON is an option for setting JSON encoding.
var EncodeJSON Option = func(w *Wrapper, r interface{}) error {
	w.Encoding = Encoding_json
	return nil
}

// EncodeDefault is the default encoder. It will be protobuf, unless the
// resource cannot be type asserted to proto.Message.
var EncodeDefault Option = func(w *Wrapper, r interface{}) error {
	encoding := Encoding_json
	if _, ok := r.(proto.Message); ok {
		encoding = Encoding_protobuf
	}
	w.Encoding = encoding
	return nil
}

// CompressNone is an option for turning off compression.
var CompressNone Option = func(w *Wrapper, r interface{}) error {
	w.Compression = Compression_none
	return nil
}

// CompressSnappy is an option for setting snappy compression.
var CompressSnappy Option = func(w *Wrapper, r interface{}) error {
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
	return wrap(r, opts...)
}

func ResourceWithoutValidation(r corev3.Resource, opts ...Option) (*Wrapper, error) {
	return wrapWithoutValidation(r, opts...)
}

// V2Resource is like Resource, but works on older core v2 resources.
func V2Resource(r corev2.Resource, opts ...Option) (*Wrapper, error) {
	return wrap(r, opts...)
}

func V2ResourceWithoutValidation(r corev2.Resource, opts ...Option) (*Wrapper, error) {
	return wrapWithoutValidation(r, opts...)
}

func wrapWithoutValidation(r interface{}, opts ...Option) (*Wrapper, error) {
	if proxy, ok := r.(*corev3.V2ResourceProxy); ok {
		r = proxy.Resource
	}
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
	w := Wrapper{
		TypeMeta: &tm,
	}
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

func wrap(r interface{}, opts ...Option) (*Wrapper, error) {
	if v, ok := r.(validatable); ok {
		if err := v.Validate(); err != nil {
			return nil, err
		}
	} else {
		return nil, ErrValidateMethodMissing
	}
	return wrapWithoutValidation(r, opts...)
}

// Unwrap unmarshals the wrapper's value into a resource, according to the
// configuration of the wrapper. The unwrapped data structure will have
// its labels and annotations set to non-nil empty slices, if they are nil.
func (w *Wrapper) Unwrap() (corev3.Resource, error) {
	r, err := w.UnwrapRaw()
	if err != nil {
		return nil, err
	}
	resource, ok := r.(corev3.Resource)
	if !ok {
		return nil, fmt.Errorf("only v3 resources can be unwrapped")
	}
	meta := resource.GetMetadata()
	if meta == nil {
		meta = new(corev2.ObjectMeta)
		resource.SetMetadata(meta)
	}
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	return resource, nil
}

// UnwrapRaw is like Unwrap, but returns a raw interface{} value.
func (w *Wrapper) UnwrapRaw() (interface{}, error) {
	resource, err := types.ResolveRaw(w.TypeMeta.APIVersion, w.TypeMeta.Type)
	if err != nil {
		return nil, err
	}
	message, err := w.Compression.Decompress(w.Value)
	if err != nil {
		return nil, fmt.Errorf("error unwrapping %T: %s", resource, err)
	}
	if err := w.Encoding.Decode(message, resource); err != nil {
		return nil, err
	}
	return resource, nil
}

// UnwrapInto unwraps a wrapper into a user-defined data structure. Most users
// should use Unwrap.
func (w *Wrapper) UnwrapInto(p interface{}) error {
	if proxy, ok := p.(*corev3.V2ResourceProxy); ok {
		p = proxy.Resource
	}
	message, err := w.Compression.Decompress(w.Value)
	if err != nil {
		return fmt.Errorf("error unwrapping %T: %s", p, err)
	}
	if err := w.Encoding.Decode(message, p); err != nil {
		return err
	}
	if resource, ok := p.(corev3.Resource); ok {
		meta := resource.GetMetadata()
		if meta.Labels == nil {
			meta.Labels = make(map[string]string)
		}
		if meta.Annotations == nil {
			meta.Annotations = make(map[string]string)
		}
	}
	return nil
}

// List is a slice of wrappers.
type List []*Wrapper

// Len tells the length of the wrap list.
func (l List) Len() int {
	return len(l)
}

// Unwrap unwraps each item in the list and returns a slice of resources of the
// same size.
func (l List) Unwrap() ([]corev3.Resource, error) {
	result := make([]corev3.Resource, len(l))
	for i := range result {
		p, err := l[i].Unwrap()
		if err != nil {
			return nil, fmt.Errorf("wrap list item %d: %s", i, err)
		}
		result[i] = p
	}
	return result, nil
}

func (l List) UnwrapInto(ptr interface{}) error {
	if len(l) == 0 {
		// if there are no elements to work on, modify nothing
		return nil
	}
	// Assume that encoding and compression are the same throughout the range
	encoding := l[0].Encoding
	compression := l[0].Compression
	// Make sure the interface is a pointer, and that the element at this address
	// is a slice.
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr {
		return &store.ErrNotValid{Err: fmt.Errorf("expected pointer, but got %v type", v.Type())}
	}
	if v.Elem().Kind() != reflect.Slice {
		return &store.ErrNotValid{Err: fmt.Errorf("expected slice, but got %s", v.Elem().Kind())}
	}
	v = v.Elem()
	if v.Cap() < len(l) {
		v.Set(reflect.MakeSlice(v.Type(), len(l), len(l)))
	}
	if v.Len() < v.Cap() {
		v.SetLen(v.Cap())
	}
	for i, w := range l {
		value, err := compression.Decompress(w.Value)
		if err != nil {
			return err
		}
		elt := v.Index(i)
		if elt.Kind() != reflect.Ptr {
			elt = elt.Addr()
		}
		if elt.IsNil() {
			elt.Set(reflect.New(elt.Type().Elem()))
		}
		if err := encoding.Decode(value, elt.Interface()); err != nil {
			return err
		}
	}
	return nil
}
