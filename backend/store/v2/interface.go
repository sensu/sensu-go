package v2

import (
	"context"

	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
)

// Wrapper is an abstraction of a store wrapper.
type Wrapper interface {
	Unwrap() (corev3.Resource, error)
	UnwrapInto(interface{}) error
}

// WrapList is a list of Wrappers.
type WrapList interface {
	Unwrap() ([]corev3.Resource, error)
	UnwrapInto(interface{}) error
	Len() int
}

// Interface specifies the interface of a v2 store.
type Interface interface {
	// NamespaceStore returns a NamespaceStore.
	NamespaceStore() NamespaceStore

	// CreateOrUpdate creates or updates the wrapped resource.
	CreateOrUpdate(context.Context, ResourceRequest, Wrapper) error

	// UpdateIfExists updates the resource with the wrapped resource, but only
	// if it already exists in the store.
	UpdateIfExists(context.Context, ResourceRequest, Wrapper) error

	// CreateIfNotExists writes the wrapped resource to the store, but only if
	// it does not already exist.
	CreateIfNotExists(context.Context, ResourceRequest, Wrapper) error

	// Get gets a wrapped resource from the store.
	Get(context.Context, ResourceRequest) (Wrapper, error)

	// Delete deletes a resource from the store.
	Delete(context.Context, ResourceRequest) error

	// List lists all resources specified by the resource request, and the
	// selection predicate.
	List(context.Context, ResourceRequest, *store.SelectionPredicate) (WrapList, error)

	// Exists returns true if the resource indicated by the request exists
	Exists(context.Context, ResourceRequest) (bool, error)

	// Patch patches the resource given in the request
	Patch(context.Context, ResourceRequest, Wrapper, patch.Patcher, *store.ETagCondition) error

	// Watch provides a channel for receiving updates to a particular resource
	// or resource collection
	Watch(context.Context, ResourceRequest) <-chan []WatchEvent

	// Initialize
	Initialize(context.Context, InitializeFunc) error
}

// NamespaceStore provides an interface for interacting with namespaces.
type NamespaceStore interface {
	// CreateOrUpdate creates or updates a corev3.Namespace resource.
	CreateOrUpdate(context.Context, *corev3.Namespace) error

	// UpdateIfExists updates the corev3.Namespace resource, but only if it
	// already exists in the store.
	UpdateIfExists(context.Context, *corev3.Namespace) error

	// CreateIfNotExists writes the corev3.Namespace resource to the store, but
	// only if it does not already exist.
	CreateIfNotExists(context.Context, *corev3.Namespace) error

	// Get gets a corev3.Namespace from the store.
	Get(context.Context, string) (*corev3.Namespace, error)

	// Delete deletes the corresponding corev3.Namespace resource. It will
	// return an error of the namespace is not empty.
	Delete(context.Context, string) error

	// List lists all corev3.Namespace resources.
	List(context.Context, *store.SelectionPredicate) ([]*corev3.Namespace, error)

	// Exists returns true if the corev3.Namespace with the provided name exists.
	Exists(context.Context, string) (bool, error)

	// Patch patches the corev3.Namespace resource with the provided name.
	Patch(context.Context, string, patch.Patcher, *store.ETagCondition) error

	// IsEmpty returns whether the corev3.Namespace with the provided name
	// is empty or if it contains other resources.
	IsEmpty(context.Context, string) (bool, error)
}

// Initialize sets up a cluster with the default resources & config.
type InitializeFunc func(context.Context) error
