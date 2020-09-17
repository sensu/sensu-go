package actions

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
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
