package etcd

import (
	"context"
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestSilencedStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		silenced := types.FixtureSilenced("checkname")
		silenced.Organization = "default"
		silenced.Environment = "default"
		silenced.Subscription = "subscription"
		silenced.ID = silenced.Subscription + ":" + silenced.CheckName
		ctx := context.WithValue(context.Background(), types.OrganizationKey, silenced.Organization)
		ctx = context.WithValue(ctx, types.EnvironmentKey, silenced.Environment)

		// We should receive an empty slice if no results were found
		silencedEntries, err := store.GetSilencedEntries(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, silencedEntries)

		err = store.UpdateSilencedEntry(ctx, silenced)
		if err != nil {
			fmt.Printf("error is %s \n", err)
			assert.FailNow(t, "failed to update entry due to error")
		}

		// Get all silenced entries
		entries, err := store.GetSilencedEntries(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, entries)
		assert.Equal(t, 1, len(entries))

		// Get silenced entry by id
		entries, err = store.GetSilencedEntryByID(ctx, silenced.ID)
		assert.NoError(t, err)
		entry := entries[0]
		assert.NotNil(t, entry)
		assert.Equal(t, silenced.CheckName, entry.CheckName)

		// Get silenced entry by subscription
		entries, err = store.GetSilencedEntriesBySubscription(ctx, silenced.Subscription)
		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.Equal(t, 1, len(entries))

		// Get silenced entry by check
		entries, err = store.GetSilencedEntriesByCheckName(ctx, silenced.CheckName)
		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.Equal(t, 1, len(entries))

		// Update silenced entry to "wildcard"
		silenced.CheckName = "*"
		silenced.ID = silenced.Subscription + ":" + silenced.CheckName
		err = store.UpdateSilencedEntry(ctx, silenced)
		assert.NoError(t, err)

		// Get silenced entry by id with "wildcard"
		entries, err = store.GetSilencedEntryByID(ctx, silenced.ID)
		assert.NoError(t, err)
		entry = entries[0]
		assert.NotNil(t, entry)
		assert.Equal(t, "subscription:*", entry.ID)

		// Delete silenced entry by id
		err = store.DeleteSilencedEntryByID(ctx, silenced.ID)
		assert.NoError(t, err)

		// Delete Silenced entry by subscription
		_ = store.UpdateSilencedEntry(ctx, silenced)
		err = store.DeleteSilencedEntriesBySubscription(ctx, silenced.Subscription)
		assert.NoError(t, err)

		// Delete silenced entry by checkName
		_ = store.UpdateSilencedEntry(ctx, silenced)
		err = store.DeleteSilencedEntriesByCheckName(ctx, silenced.CheckName)
		assert.NoError(t, err)

		// Updating a check in a nonexistent org and env should not work
		silenced.Organization = "missing"
		silenced.Environment = "missing"
		err = store.UpdateSilencedEntry(ctx, silenced)
		assert.Error(t, err)
	})
}
