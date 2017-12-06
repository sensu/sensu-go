package eventd

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// Add a list of all silenced subscriptions to the event.
func getSilenced(ctx context.Context, event *types.Event, s store.Store) error {
	var (
		silencedSubscriptions []*types.Silenced
		silencedChecks        []*types.Silenced
		silencedEntries       map[string]bool
		err                   error
	)

	// Get all silenced entries by agent subscription and check name. This takes
	// any wildcards into account (subscription:* or *checkName).
	// TODO: implement deletion of silenced entries that have ExpireOnResolve set
	// to true. As of this writing, what constitutes a check resolution is TBD,
	// but will probably involve the check state (passing/failing).
	// also add checkName to this entry list
	silencedEntries = make(map[string]bool)

	// Get the silenced entries for the entity
	for _, value := range event.Entity.Subscriptions {
		silencedSubscriptions, err = s.GetSilencedEntriesBySubscription(ctx, value)
		if err != nil {
			return err
		}
	}
	appendEntries(event, silencedSubscriptions, silencedEntries)

	// Get the silenced entries for the check
	silencedChecks, err = s.GetSilencedEntriesByCheckName(ctx, event.Check.Config.Name)
	if err != nil {
		return err
	}
	appendEntries(event, silencedChecks, silencedEntries)

	return nil
}

// iterate through silencedEntries and create an array of silenced entries in
// the event minus any duplicates.
func appendEntries(event *types.Event, silenced []*types.Silenced, silencedEntries map[string]bool) {
	for _, entry := range silenced {
		if _, value := silencedEntries[entry.ID]; !value {
			silencedEntries[entry.ID] = true
			event.Silenced = append(event.Silenced, entry.ID)
		}
	}
}
