package actions

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// ExtensionController expose actions in which a viewer can perform.
type ExtensionController struct {
	store store.ExtensionRegistry
}

// NewExtensionController returns new ExtensionController
func NewExtensionController(store store.ExtensionRegistry) ExtensionController {
	return ExtensionController{
		store: store,
	}
}

// List returns resources available to the viewer filter by given params.
func (e ExtensionController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	// Fetch from store
	results, err := e.store.GetExtensions(ctx, pred)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	resources := make([]corev2.Resource, len(results))
	for i, v := range results {
		resources[i] = corev2.Resource(v)
	}

	return resources, nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (e ExtensionController) Find(ctx context.Context, name string) (*types.Extension, error) {
	// Validate params
	if id := name; id == "" {
		return nil, NewErrorf(InternalErr, "'id' param missing")
	}

	// Fetch from store
	result, serr := e.store.GetExtension(ctx, name)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}
	if result == nil {
		return nil, NewErrorf(NotFound)
	}

	return result, nil
}

// Register creates or replaces the extension given.
func (e ExtensionController) Register(ctx context.Context, extension types.Extension) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &extension)

	// Validate
	if err := extension.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist Changes
	if serr := e.store.RegisterExtension(ctx, &extension); serr != nil {
		return NewError(InternalErr, serr)
	}

	return nil
}

// Deregister deletes the extension from the registry.
func (e ExtensionController) Deregister(ctx context.Context, name string) error {
	if err := e.store.DeregisterExtension(ctx, name); err != nil {
		return NewError(InternalErr, nil)
	}
	return nil
}
