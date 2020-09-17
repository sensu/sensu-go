package v2

import (
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
)

// Type alias for documentation clarity
type Wrapper = wrap.Wrapper

// Interface specifies the interface of a v2 store.
type Interface interface {
	// CreateOrUpdate creates or updates the wrapped resource.
	CreateOrUpdate(ResourceRequest, *Wrapper) error

	// UpdateIfExists updates the resource with the wrapped resource, but only
	// if it already exists in the store.
	UpdateIfExists(ResourceRequest, *Wrapper) error

	// CreateIfNotExists writes the wrapped resource to the store, but only if
	// it does not already exist.
	CreateIfNotExists(ResourceRequest, *Wrapper) error

	// Get gets a wrapped resource from the store.
	Get(ResourceRequest) (*Wrapper, error)

	// Delete deletes a resource from the store.
	Delete(ResourceRequest) error

	// List lists all resources specified by the resource request, and the
	// selection predicate.
	List(ResourceRequest, *store.SelectionPredicate) (wrap.List, error)

	// Exists returns true if the resource indicated by the request exists
	Exists(ResourceRequest) (bool, error)

	// Patch patches the resource given in the request
	Patch(ResourceRequest, *Wrapper, patch.Patcher, *store.ETagCondition) error
}
