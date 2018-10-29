package core

import metav1 "github.com/sensu/sensu-go/internal/apis/meta/v1"

// A Namespace is a resource that defines where other resources are located.
type Namespace struct {
	metav1.TypeMeta   `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`
	metav1.ObjectMeta `json:"metadata" protobuf:"bytes,2,opt,name=metadata"`
}
