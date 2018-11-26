package actions

import (
	"context"
	"time"

	"github.com/sensu/sensu-go/backend/authentication/jwt"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// silencedUpdateFields whitelists fields allowed to be updated for Silences
var silencedUpdateFields = []string{
	"Expire",
	"ExpireOnResolve",
	"Reason",
	"Begin",
}

// SilencedController exposes actions in which a viewer can perform.
type SilencedController struct {
	Store store.SilencedStore
}

// NewSilencedController returns new SilencedController
func NewSilencedController(store store.SilencedStore) SilencedController {
	return SilencedController{
		Store: store,
	}
}

// Query returns resources available to the viewer.
func (a SilencedController) Query(ctx context.Context, sub, check string) ([]*types.Silenced, error) {
	var results []*types.Silenced
	var serr error
	if sub != "" {
		results, serr = a.Store.GetSilencedEntriesBySubscription(ctx, sub)
	} else if check != "" {
		results, serr = a.Store.GetSilencedEntriesByCheckName(ctx, check)
	} else {
		results, serr = a.Store.GetSilencedEntries(ctx)
	}
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	return results, nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (a SilencedController) Find(ctx context.Context, id string) (*types.Silenced, error) {
	// Fetch from store
	result, err := a.findSilencedEntry(ctx, id)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, NewErrorf(NotFound)
	}

	return result, nil
}

// Create creates a new silenced entry. It returns an error if the entry already exists.
func (a SilencedController) Create(ctx context.Context, newSilence *types.Silenced) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, newSilence)

	// Populate newSilence.Name with the subscription and checkName. Substitute a
	// splat if one of the values does not exist. If both values are empty, the
	// validator will return an error when attempting to update it in the store.
	newSilence.Name, _ = types.SilencedName(newSilence.Subscription, newSilence.Check)

	// If begin timestamp was not already provided set it to the current time.
	if newSilence.Begin == 0 {
		newSilence.Begin = time.Now().Unix()
	}

	// Retrieve the subject of the JWT, which represents the logged on user, in
	// order to set it as the creator of the silenced entry
	if claims := jwt.GetClaimsFromContext(ctx); claims != nil {
		newSilence.Creator = claims.Subject
	}

	// Check for existing
	if e, serr := a.Store.GetSilencedEntryByName(ctx, newSilence.Name); serr != nil {
		return NewError(InternalErr, serr)
	} else if e != nil {
		return NewErrorf(AlreadyExistsErr)
	}

	// Validate
	if err := newSilence.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := a.Store.UpdateSilencedEntry(ctx, newSilence); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// CreateOrReplace creates or replaces a silenced entry.
func (a SilencedController) CreateOrReplace(ctx context.Context, newSilence types.Silenced) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &newSilence)

	// Populate newSilence.Name with the subscription and checkName. Substitute a
	// splat if one of the values does not exist. If both values are empty, the
	// validator will return an error when attempting to update it in the store.
	newSilence.Name, _ = types.SilencedName(newSilence.Subscription, newSilence.Check)

	// If begin timestamp was not already provided set it to the current time.
	if newSilence.Begin == 0 {
		newSilence.Begin = time.Now().Unix()
	}

	// Retrieve the subject of the JWT, which represents the logged on user, in
	// order to set it as the creator of the silenced entry
	if claims := jwt.GetClaimsFromContext(ctx); claims != nil {
		newSilence.Creator = claims.Subject
	}

	// Validate
	if err := newSilence.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := a.Store.UpdateSilencedEntry(ctx, &newSilence); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Destroy removes a resource if viewer has access.
func (a SilencedController) Destroy(ctx context.Context, id string) error {
	// Fetch from store
	result, serr := a.Store.GetSilencedEntryByName(ctx, id)
	if serr != nil {
		return NewError(InternalErr, serr)
	} else if result == nil {
		return NewErrorf(NotFound)
	}

	if err := a.Store.DeleteSilencedEntryByName(ctx, id); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

func (a SilencedController) findSilencedEntry(ctx context.Context, id string) (*types.Silenced, error) {
	result, serr := a.Store.GetSilencedEntryByName(ctx, id)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}
	if result != nil {
		return result, nil
	}

	return nil, NewErrorf(NotFound)
}
