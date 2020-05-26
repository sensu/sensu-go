package v3

import corev2 "github.com/sensu/sensu-go/api/core/v2"

var _ corev2.Resource = v2ResourceProxy{}

// Resource represents a Sensu resource.
type Resource interface {
	// GetMetadata returns the object metadata for the resource.
	GetMetadata() *corev2.ObjectMeta

	// StoreSuffix gives the path suffix to this resource type in the store.
	StoreSuffix() string

	// RBACName describes the name of the resource for RBAC purposes.
	RBACName() string

	// URIPath gives the path to the resource, e.g. /checks/checkname
	URIPath() string

	// Validate checks if the fields in the resource are valid.
	Validate() error
}

type v2ResourceProxy struct {
	Resource
}

func V3ToV2Resource(resource Resource) corev2.Resource {
	return v2ResourceProxy{Resource: resource}
}

func (v v2ResourceProxy) GetObjectMeta() corev2.ObjectMeta {
	meta := v.GetMetadata()
	if meta == nil {
		return corev2.ObjectMeta{}
	}
	return *meta
}

func (v v2ResourceProxy) SetObjectMeta(meta corev2.ObjectMeta) {
	ptr := v.GetMetadata()
	*ptr = meta
}

func (v v2ResourceProxy) SetNamespace(ns string) {
	v.GetMetadata().Namespace = ns
}

func (v v2ResourceProxy) StorePrefix() string {
	return v.StoreSuffix()
}
