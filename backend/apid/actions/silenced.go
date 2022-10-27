package actions

import (
	"context"
	"errors"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// SilencedController exposes actions in which a viewer can perform.
type SilencedController struct {
	Store store.SilenceStore
}

// NewSilencedController returns new SilencedController
func NewSilencedController(store store.SilenceStore) SilencedController {
	return SilencedController{
		Store: store,
	}
}

// List returns resources available to the viewer.
func (c SilencedController) List(ctx context.Context, sub, check string) ([]*corev2.Silenced, error) {
	var results []*types.Silenced
	namespace := corev2.ContextNamespace(ctx)
	var serr error
	if sub != "" {
		results, serr = c.Store.GetSilencesBySubscription(ctx, namespace, []string{sub})
	} else if check != "" {
		results, serr = c.Store.GetSilencesByCheck(ctx, namespace, check)
	} else {
		results, serr = c.Store.GetSilences(ctx, namespace)
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

	namespace := corev2.ContextNamespace(ctx)

	// Validate the silenced entry
	if err := entry.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	if claims := jwt.GetClaimsFromContext(ctx); claims != nil {
		entry.CreatedBy = claims.StandardClaims.Subject
	}

	// Check for existing
	if e, serr := c.Store.GetSilenceByName(ctx, namespace, entry.Name); serr != nil {
		return NewError(InternalErr, serr)
	} else if e != nil {
		return NewErrorf(AlreadyExistsErr)
	}

	// Persist
	if err := c.Store.UpdateSilence(ctx, entry); err != nil {
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
	if err := c.Store.UpdateSilence(ctx, entry); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

func (c SilencedController) Get(ctx context.Context, name string) (*corev2.Silenced, error) {
	entry, err := c.Store.GetSilenceByName(ctx, corev2.ContextNamespace(ctx), name)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}
	if entry == nil {
		return nil, NewError(NotFound, errors.New("silenced entry not found"))
	}
	return entry, nil
}
