package eventd

import (
	"context"
	"fmt"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	stringsutil "github.com/sensu/sensu-go/util/strings"
)

// addToSilencedBy takes a silenced entry ID and adds it to a silence of IDs if
// it's not already present in order to avoid duplicated elements
func addToSilencedBy(id string, ids []string) []string {
	if !stringsutil.InArray(id, ids) {
		ids = append(ids, id)
	}
	return ids
}

// getSilenced retrieves all silenced entries for a given event, using the
// entity subscription, the check subscription and the check name while
// supporting wildcard silenced entries (e.g. subscription:*)
func getSilenced(ctx context.Context, event *types.Event, s store.Store) error {
	if !event.HasCheck() {
		return nil
	}

	// Retrieve silenced entries using the entity subscription
	entitySubscription := types.GetEntitySubscription(event.Entity.Name)
	subscriptions := append(event.Check.Subscriptions, entitySubscription)

	entries, err := s.GetSilencedEntriesBySubscription(ctx, subscriptions...)
	if err != nil {
		return fmt.Errorf("error setting silenced entries: %s", err)
	}

	// Retrieve silenced entries using the check name
	results, err := s.GetSilencedEntriesByCheckName(ctx, event.Check.Name)
	if err != nil {
		return err
	}
	entries = append(entries, results...)

	// Determine which entries silence this event
	silencedIDs := silencedBy(event, entries)

	// Add to the event all silenced entries ID that actually silence it
	event.Check.Silenced = silencedIDs

	return nil
}

// silencedBy determines which of the given silenced entries silenced a given
// event and return a list of silenced entry IDs
func silencedBy(event *types.Event, silencedEntries []*types.Silenced) []string {
	silencedBy := event.SilencedBy(silencedEntries)
	names := make([]string, 0, len(silencedBy))
	for _, entry := range silencedBy {
		names = addToSilencedBy(entry.Name, names)
	}
	return names
}

func handleExpireOnResolveEntries(ctx context.Context, event *types.Event, store store.Store) error {
	// Make sure we have a check and that the event is a resolution
	if !event.HasCheck() || !event.IsResolution() {
		return nil
	}

	entries, err := store.GetSilencedEntriesByName(ctx, event.Check.Silenced...)
	if err != nil {
		return fmt.Errorf("couldn't resolve silences: %s", err)
	}
	toDelete := entries[:0]
	toRetain := []string{}
	for _, entry := range entries {
		if entry.ExpireOnResolve {
			toDelete = append(toDelete, entry)
		} else {
			toRetain = append(toRetain, entry.Name)
		}
	}

	event.Check.Silenced = toRetain

	return nil
}
