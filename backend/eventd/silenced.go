package eventd

import (
	"context"
	"fmt"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// Add a list of all silenced subscriptions to the event.
func getSilenced(ctx context.Context, event *types.Event, s store.Store) error {
	var (
		silencedEntries []*types.Silenced
		err             error
	)

	// Get all silenced entries by agent subscription - this takes any wildcards
	// (subscription:*) and check names into account.
	// TODO: implement deletion of silenced entries that have ExpireOnResolve set
	// to true. As of this writing, what constitutes a check resolution is TBD,
	// but will probably involve the check state (passing/failing).
	if event.Check.Status != 0 {
		for i := 0; i < len(event.Entity.Subscriptions); i++ {
			silencedEntries, err = s.GetSilencedEntriesBySubscription(ctx, event.Entity.Subscriptions[i])
		}
		fmt.Println("Silenced entries: %s", silencedEntries)
		if err != nil {
			return err
		}
	}
	// iterate through silencedEntries and create an array of silenced entries in
	// the event.
	for i := 0; i < len(silencedEntries); i++ {
		silencedEntry := silencedEntries[i]
		event.Silenced = append(event.Silenced, silencedEntry.Subscription)
	}
	return nil
}
