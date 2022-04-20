package v3

import (
	"encoding/json"
	"reflect"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	types "github.com/sensu/sensu-go/types"
)

var _ corev2.Resource = &V2ResourceProxy{}

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

	// StoreName gives the name of the resource as it pertains to a store.
	StoreName() string

	// RBACName describes the name of the resource for RBAC purposes.
	RBACName() string

	// URIPath gives the path to the resource, e.g. /checks/checkname
	URIPath() string

	// Validate checks if the fields in the resource are valid.
	Validate() error
}

// GlobalResource  is an interface for indicating
// a resource's namespace strategy
type GlobalResource interface {
	// IsGlobalResource returns true when the resource
	// is not namespaced.
	IsGlobalResource() bool
}

// V2ResourceProxy is a compatibility shim for converting from a v3 resource to
// a v2 resource.
//sensu:nogen
type V2ResourceProxy struct {
	Resource
}

func V3ToV2Resource(resource Resource) corev2.Resource {
	return &V2ResourceProxy{Resource: resource}
}

func (v *V2ResourceProxy) GetObjectMeta() corev2.ObjectMeta {
	meta := v.GetMetadata()
	if meta == nil {
		return corev2.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		}
	}
	return *meta
}

func (v *V2ResourceProxy) SetObjectMeta(meta corev2.ObjectMeta) {
	v.SetMetadata(&meta)
}

func (v *V2ResourceProxy) SetNamespace(ns string) {
	// SetNamespace expected to be a no-op for
	// core/v2 Resources. sensu-go and sensu-go/types
	// both depend on this property.
	if gr, ok := v.Resource.(GlobalResource); ok {
		if gr.IsGlobalResource() {
			return
		}
	}
	if v.GetMetadata() == nil {
		v.SetMetadata(&corev2.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		})
	}
	v.GetMetadata().Namespace = ns
}

func (v *V2ResourceProxy) StorePrefix() string {
	return v.StoreName()
}

func (v V2ResourceProxy) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Resource)
}

func (v *V2ResourceProxy) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &v.Resource)
}

// tmGetter is useful for types that want to explicitly provide their
// TypeMeta - this is useful for lifters.
type tmGetter interface {
	GetTypeMeta() corev2.TypeMeta
}

func (v V2ResourceProxy) GetTypeMeta() corev2.TypeMeta {
	var tm corev2.TypeMeta
	if getter, ok := v.Resource.(tmGetter); ok {
		tm = getter.GetTypeMeta()
	} else {
		typ := reflect.Indirect(reflect.ValueOf(v.Resource)).Type()
		tm = corev2.TypeMeta{
			Type:       typ.Name(),
			APIVersion: types.ApiVersion(typ.PkgPath()),
		}
	}
	return tm
}
