package compat

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
)

// URIPath gets the URIPath from either a core/v2 resource or a core/v3 resource.
// It panics if the value passed is not a corev2 or corev3 resource.
func URIPath(value interface{}) string {
	switch value := value.(type) {
	case corev2.Resource:
		return value.URIPath()
	case corev3.Resource:
		return value.URIPath()
	}
	// impossible unless the type resolver is broken. fatal error.
	panic("got neither corev2 resource nor corev3 resource")
}

// GetObjectMeta gets the object metadata from either a corev2 or corev3 resource.
// It panics if the value passed is not a corev2 or corev3 resource.
func GetObjectMeta(value interface{}) *corev2.ObjectMeta {
	switch value := value.(type) {
	case corev2.Resource:
		meta := value.GetObjectMeta()
		return &meta
	case corev3.Resource:
		return value.GetMetadata()
	}
	// impossible unless the type resolver is broken. fatal error.
	panic("got neither corev2 resource nor corev3 resource")
}

// SetNamespace sets the namespace.
// It panics if the value passed is not a corev2 or corev3 resource.
func SetNamespace(value interface{}, namespace string) {
	switch value := value.(type) {
	case corev2.Resource:
		value.SetNamespace(namespace)
	case corev3.Resource:
		value.GetMetadata().Namespace = namespace
	default:
		// impossible unless the type resolver is broken. fatal error.
		panic("got neither corev2 resource nor corev3 resource")
	}
}

// SetObjectMeta sets the object metadata.
// It panics if the value passed is not a corev2 or corev3 resource.
func SetObjectMeta(value interface{}, meta *corev2.ObjectMeta) {
	switch value := value.(type) {
	case corev2.Resource:
		value.SetObjectMeta(*meta)
	case corev3.Resource:
		value.SetMetadata(meta)
	default:
		// impossible unless the type resolver is broken. fatal error.
		panic("got neither corev2 resource nor corev3 resource")
	}
}

// V2Resource returns its input if it's already a v2 resource, or converts from
// v3 otherwise.
// It panics if the value passed is not a corev2 or corev3 resource.
func V2Resource(value interface{}) corev2.Resource {
	switch value := value.(type) {
	case corev2.Resource:
		return value
	case corev3.Resource:
		return corev3.V3ToV2Resource(value)
	default:
		// impossible unless the type resolver is broken. fatal error.
		panic("got neither corev2 resource nor corev3 resource")
	}
}
