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

type InitializeFunc func(context.Context) error

// Initializer provides methods to verify if a store is initialized
type Initializer interface {
	// Close closes the session to the store and unlock any mutex
	Close(context.Context) error

	// FlagAsInitialized marks the store as initialized
	FlagAsInitialized(context.Context) error

	// IsInitialized returns a boolean error that indicates if the store has been
	// initialized or not
	IsInitialized(context.Context) (bool, error)

	// Lock locks a mutex to avoid competing writes
	Lock(context.Context) error
}
