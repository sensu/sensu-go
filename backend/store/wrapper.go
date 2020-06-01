package store

import (
	"errors"

	"github.com/golang/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
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

// WrapResource wraps the given resource in a wrapper designed for storage
func WrapResource(r corev3.Resource) (Wrapper, error) {
	var tm corev2.TypeMeta
	if getter, ok := r.(tmGetter); ok {
		tm = getter.GetTypeMeta()
	} else {
		return Wrapper{}, errors.New("given resource does not have a GetTypeMeta method")
	}

	bytes, err := proto.Marshal(r)
	if err != nil {
		return Wrapper{}, err
	}

	return Wrapper{
		TypeMeta:   &tm,
		ObjectMeta: r.GetMetadata(),
		Value:      bytes,
	}, nil
}
