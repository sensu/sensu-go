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
	// todo: uncomment after #4765 implements namespacestore on postgres store
	// NamespaceStore

	// CreateOrUpdate creates or updates the wrapped resource.
	CreateOrUpdate(ResourceRequest, Wrapper) error

	// UpdateIfExists updates the resource with the wrapped resource, but only
	// if it already exists in the store.
	UpdateIfExists(ResourceRequest, Wrapper) error

	// CreateIfNotExists writes the wrapped resource to the store, but only if
	// it does not already exist.
	CreateIfNotExists(ResourceRequest, Wrapper) error

	// Get gets a wrapped resource from the store.
	Get(ResourceRequest) (Wrapper, error)

	// Delete deletes a resource from the store.
	Delete(ResourceRequest) error

	// List lists all resources specified by the resource request, and the
	// selection predicate.
	List(ResourceRequest, *store.SelectionPredicate) (WrapList, error)

	// Exists returns true if the resource indicated by the request exists
	Exists(ResourceRequest) (bool, error)

	// Patch patches the resource given in the request
	Patch(ResourceRequest, Wrapper, patch.Patcher, *store.ETagCondition) error

	// Watch provides a channel for receiving updates to a particular resource
	// or resource collection
	// Watch(context.Context, ResourceRequest) <-chan []WatchEvent
}

type WatchActionType uint

const (
	Create WatchActionType = iota
	Update
	Delete
	WatchError
)

type WatchEvent struct {
	Type          WatchActionType
	Key           ResourceRequest
	Value         Wrapper
	PreviousValue Wrapper
	Revision      int64
	Err           error
}

type NamespaceStore interface {
	// CreateNamespace persists a corev3.Namespace resource
	CreateNamespace(context.Context, *corev3.Namespace) error
	// DeleteNamespace deletes the corresponding corev3.Namespace resource and
	// cleans up any store resources from that namespace.
	DeleteNamespace(context.Context, string) error
}
