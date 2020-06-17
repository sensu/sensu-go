package actions

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

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

// List returns resources available to the viewer.
func (c SilencedController) List(ctx context.Context, sub, check string) ([]*corev2.Silenced, error) {
	var results []*types.Silenced
	var serr error
	if sub != "" {
		results, serr = c.Store.GetSilencedEntriesBySubscription(ctx, sub)
	} else if check != "" {
		results, serr = c.Store.GetSilencedEntriesByCheckName(ctx, check)
	} else {
		results, serr = c.Store.GetSilencedEntries(ctx)
	}
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	return results, nil
}

// Create creates a new silenced entry. It returns an error if the entry already exists.
func (c SilencedController) Create(ctx context.Context, entry *corev2.Silenced) error {
	// Prepare the silenced entry for storage
	entry.Prepare(ctx)

	// Validate the silenced entry
	if err := entry.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	if claims := jwt.GetClaimsFromContext(ctx); claims != nil {
		entry.CreatedBy = claims.StandardClaims.Subject
	}

	// Check for existing
	if e, serr := c.Store.GetSilencedEntryByName(ctx, entry.Name); serr != nil {
		return NewError(InternalErr, serr)
	} else if e != nil {
		return NewErrorf(AlreadyExistsErr)
	}

	// Persist
	if err := c.Store.UpdateSilencedEntry(ctx, entry); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// CreateOrReplace creates or replaces a silenced entry.
func (c SilencedController) CreateOrReplace(ctx context.Context, entry *corev2.Silenced) error {
	// Prepare the silenced entry for storage
	entry.Prepare(ctx)

	// Validate the silenced entry
	if err := entry.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	if claims := jwt.GetClaimsFromContext(ctx); claims != nil {
		entry.CreatedBy = claims.StandardClaims.Subject
	}

	// Persist
	if err := c.Store.UpdateSilencedEntry(ctx, entry); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}
