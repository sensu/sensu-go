package actions

import (
	"context"
	"encoding/base64"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// ExtensionController expose actions in which a viewer can perform.
type ExtensionController struct {
	Store store.ExtensionRegistry
}

// NewExtensionController returns new ExtensionController
func NewExtensionController(store store.ExtensionRegistry) ExtensionController {
	return ExtensionController{
		Store: store,
	}
}

// Query returns resources available to the viewer filter by given params.
func (e ExtensionController) Query(ctx context.Context) ([]*types.Extension, string, error) {
	pageSize := corev2.PageSizeFromContext(ctx)
	continueToken := corev2.PageContinueFromContext(ctx)

	// Fetch from store
	results, newContinueToken, serr := e.Store.GetExtensions(ctx, int64(pageSize), continueToken)
	if serr != nil {
		return nil, "", NewError(InternalErr, serr)
	}

	// Encode the continue token with base64url (RFC 4648), without padding
	encodedNewContinueToken := base64.RawURLEncoding.EncodeToString([]byte(newContinueToken))

	return results, encodedNewContinueToken, nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (e ExtensionController) Find(ctx context.Context, name string) (*types.Extension, error) {
	// Validate params
	if id := name; id == "" {
		return nil, NewErrorf(InternalErr, "'id' param missing")
	}

	// Fetch from store
	result, serr := e.Store.GetExtension(ctx, name)
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
	if serr := e.Store.RegisterExtension(ctx, &extension); serr != nil {
		return NewError(InternalErr, serr)
	}

	return nil
}

// Deregister deletes the extension from the registry.
func (e ExtensionController) Deregister(ctx context.Context, name string) error {
	if err := e.Store.DeregisterExtension(ctx, name); err != nil {
		return NewError(InternalErr, nil)
	}
	return nil
}
