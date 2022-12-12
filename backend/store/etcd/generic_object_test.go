package etcd

import corev2 "github.com/sensu/core/v2"

//go:generate -command protoc protoc --gofast_out=plugins:. -I=${GOPATH}/src:. -I=../../../vendor/ -I=./ -I=../../../vendor/github.com/gogo/protobuf/protobuf/
//go:generate protoc generic_object.proto

func (g *GenericObject) GetNamespace() string {
	return g.Namespace
}

func (g *GenericObject) SetObjectMeta(meta corev2.ObjectMeta) {
	g.ObjectMeta = meta
}

func (g *GenericObject) StorePrefix() string {
	return ""
}

func (g *GenericObject) URIPath() string {
	return ""
}

func (g *GenericObject) Validate() error {
	return nil
}

// SetNamespace sets the namespace of the resource.
func (g *GenericObject) SetNamespace(namespace string) {
	g.Namespace = namespace
}
