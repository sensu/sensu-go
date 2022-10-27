package v2

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
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

// EntityConfigStoreGetter gets you an EntityConfigStore.
type EntityConfigStoreGetter interface {
	GetEntityConfigStore() EntityConfigStore
}

// EntityStateStoreGetter gets you an EntityStateStore.
type EntityStateStoreGetter interface {
	GetEntityStateStore() EntityStateStore
}

// NamespaceStoreGetter gets you a NamespaceStore.
type NamespaceStoreGetter interface {
	GetNamespaceStore() NamespaceStore
}

// Interface specifies the interface of a v2 store.
type Interface interface {
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

	// Delete deletes the corev3.Namespace, from the store, corresponding with
	// the given name. It will return an error of the namespace is not empty.
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

// EntityConfigStore provides an interface for interacting with entity configs.
type EntityConfigStore interface {
	// CreateOrUpdate creates or updates a corev3.EntityConfig resource.
	CreateOrUpdate(context.Context, *corev3.EntityConfig) error

	// UpdateIfExists updates the corev3.EntityConfig resource, but only if it
	// already exists in the store.
	UpdateIfExists(context.Context, *corev3.EntityConfig) error

	// CreateIfNotExists writes the corev3.EntityConfig resource to the store, but
	// only if it does not already exist.
	CreateIfNotExists(context.Context, *corev3.EntityConfig) error

	// Get gets a corev3.EntityConfig from the store.
	Get(context.Context, string, string) (*corev3.EntityConfig, error)

	// Delete deletes the corev3.EntityConfig, from the store, corresponding
	// with the given namespace and name.
	Delete(context.Context, string, string) error

	// List lists all corev3.EntityConfig resources.
	List(context.Context, string, *store.SelectionPredicate) ([]*corev3.EntityConfig, error)

	// Exists returns true if the corev3.EntityConfig exists for the provided
	// namespace and name.
	Exists(context.Context, string, string) (bool, error)

	// Patch patches the corev3.EntityConfig resource with the provided
	// namespace and name.
	Patch(context.Context, string, string, patch.Patcher, *store.ETagCondition) error
}

// EntityStateStore provides an interface for interacting with entity states.
type EntityStateStore interface {
	// CreateOrUpdate creates or updates a corev3.EntityState resource.
	CreateOrUpdate(context.Context, *corev3.EntityState) error

	// UpdateIfExists updates the corev3.EntityState resource, but only if it
	// already exists in the store.
	UpdateIfExists(context.Context, *corev3.EntityState) error

	// CreateIfNotExists writes the corev3.EntityState resource to the store, but
	// only if it does not already exist.
	CreateIfNotExists(context.Context, *corev3.EntityState) error

	// Get gets a corev3.EntityState from the store.
	Get(context.Context, string, string) (*corev3.EntityState, error)

	// Delete deletes the corev3.EntityState, from the store, corresponding
	// with the given namespace and name.
	Delete(context.Context, string, string) error

	// List lists all corev3.EntityState resources.
	List(context.Context, string, *store.SelectionPredicate) ([]*corev3.EntityState, error)

	// Exists returns true if the corev3.EntityState exists for the provided
	// namespace and name.
	Exists(context.Context, string, string) (bool, error)

	// Patch patches the corev3.EntityState resource with the provided
	// namespace and name.
	Patch(context.Context, string, string, patch.Patcher, *store.ETagCondition) error
}

// KeepaliveStore provides an interface for interacting with keepalives.
type KeepaliveStore interface {
	// DeleteFailingKeepalive deletes a failing keepalive record for a given entity.
	DeleteFailingKeepalive(ctx context.Context, entityConfig *corev3.EntityConfig) error

	// GetFailingKeepalives returns a slice of failing keepalives.
	// TODO: create a corev3.KeepaliveRecord type so we can remove the
	// dependency on corev2
	GetFailingKeepalives(ctx context.Context) ([]*corev2.KeepaliveRecord, error)

	// UpdateFailingKeepalive updates the given entity keepalive with the given expiration
	// in unix timestamp format
	UpdateFailingKeepalive(ctx context.Context, entityConfig *corev3.EntityConfig, expiration int64) error
}

// Initialize sets up a cluster with the default resources & config.
type InitializeFunc func(context.Context, Interface, NamespaceStore) error
