package v2

import (
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/wrap"
)

// Interface specifies the interface of a v2 store.
type Interface interface {
	// CreateOrUpdate creates or updates the wrapped resource.
	CreateOrUpdate(ResourceRequest, *wrap.Wrapper) error

	// UpdateIfExists updates the resource with the wrapped resource, but only
	// if it already exists in the store.
	UpdateIfExists(ResourceRequest, *wrap.Wrapper) error

	// CreateIfNotExists writes the wrapped resource to the store, but only if
	// it does not already exist.
	CreateIfNotExists(ResourceRequest, *wrap.Wrapper) error

	// Get gets a wrapped resource from the store.
	Get(ResourceRequest) (*wrap.Wrapper, error)

	// GetIfExists gets a wrapped resource from the store, and returns an error
	// if it does not exist already.
	GetIfExists(ResourceRequest) (*wrap.Wrapper, error)

	// Delete deletes a resource from the store.
	Delete(ResourceRequest) error

	// DeleteIfExists deletes a resource from the store, and returns an error
	// if it does not exist.
	DeleteIfExists(ResourceRequest) error

	// List lists all resources specified by the resource request, and the
	// selection predicate.
	List(ResourceRequest, *store.SelectionPredicate) ([]*wrap.Wrapper, error)
}
