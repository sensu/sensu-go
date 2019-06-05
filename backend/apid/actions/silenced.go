package actions

import (
	"context"

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
