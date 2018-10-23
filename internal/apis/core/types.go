package core

import "github.com/sensu/sensu-go/internal/apis/meta"

// A Namespace is a resource that defines where other resources are located.
type Namespace struct {
	meta.TypeMeta   `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`
	meta.ObjectMeta `json:"metadata" protobuf:"bytes,2,opt,name=metadata"`
}
