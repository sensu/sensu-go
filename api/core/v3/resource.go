package v3

import corev2 "github.com/sensu/sensu-go/api/core/v2"

var _ corev2.Resource = v2ResourceProxy{}

// Resource represents a Sensu resource.
type Resource interface {
	// GetMetadata returns the object metadata for the resource.
	GetMetadata() *corev2.ObjectMeta
	GeneratedType
}

// GeneratedType is an interface that specifies all the methods that are
// automatically generated.
type GeneratedType interface {
	// SetMetadata sets metadata on its receiver, if the receiver has metadata
	// to set. If the receiver does not have metadata to set, nothing happens.
	SetMetadata(*corev2.ObjectMeta)

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
		return corev2.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		}
	}
	return *meta
}

func (v v2ResourceProxy) SetObjectMeta(meta corev2.ObjectMeta) {
	v.SetMetadata(&meta)
}

func (v v2ResourceProxy) SetNamespace(ns string) {
	if v.GetMetadata() == nil {
		v.SetMetadata(&corev2.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		})
	}
	v.GetMetadata().Namespace = ns
}

func (v v2ResourceProxy) StorePrefix() string {
	return v.StoreSuffix()
}
