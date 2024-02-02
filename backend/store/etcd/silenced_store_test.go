//go:build integration && !race
// +build integration,!race

package etcd

import (
	"context"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSilencedStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		silenced := types.FixtureSilenced("*:checkname")
		silenced.Namespace = "default"
		silenced.Subscription = "subscription"
		silenced.Name = silenced.Subscription + ":" + silenced.Check
		ctx := context.WithValue(context.Background(), types.NamespaceKey, silenced.Namespace)

		// We should receive an empty slice if no results were found
		silencedEntries, err := store.GetSilencedEntries(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, silencedEntries)

		err = store.UpdateSilencedEntry(ctx, silenced)
		if err != nil {
			t.Fatalf("failed to update entry due to error: %s", err)
		}

		// Get all silenced entries
		entries, err := store.GetSilencedEntries(ctx)
		require.NoError(t, err)
		require.NotNil(t, entries)
		require.Equal(t, 1, len(entries))

		// Get silenced entry by name
		entry, err := store.GetSilencedEntryByName(ctx, silenced.Name)
		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.Equal(t, silenced.Check, entry.Check)

		// Get silenced entry by subscription
		entries, err = store.GetSilencedEntriesBySubscription(ctx, silenced.Subscription)
		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.Equal(t, 1, len(entries))

		// Get silenced entry by check
		entries, err = store.GetSilencedEntriesByCheckName(ctx, silenced.Check)
		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.Equal(t, 1, len(entries))

		// Update silenced entry to "wildcard"
		silenced.Check = "*"
		silenced.Name = silenced.Subscription + ":" + silenced.Check
		err = store.UpdateSilencedEntry(ctx, silenced)
		assert.NoError(t, err)

		// Get silenced entry by name with "wildcard"
		entry, err = store.GetSilencedEntryByName(ctx, silenced.Name)
		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.Equal(t, "subscription:*", entry.Name)

		// check entries without -1 expiry
		assert.NotEqual(t, int64(-1), entry.Expire)

		// Delete silenced entry by name
		err = store.DeleteSilencedEntryByName(ctx, silenced.Name)
		assert.NoError(t, err)

		// Update a silenced entry's expire time
		silenced.Expire = 2
		silenced.ExpireAt = 0
		err = store.UpdateSilencedEntry(ctx, silenced)
		assert.NoError(t, err)

		// Wait for the entry to expire
		time.Sleep(3 * time.Second)

		// Check that the entry is deleted
		entry, err = store.GetSilencedEntryByName(ctx, silenced.Name)
		assert.NoError(t, err)
		assert.Nil(t, entry)

		// Updating a silenced entry in a nonexistent org and env should not work
		silenced.Namespace = "missing"
		err = store.UpdateSilencedEntry(ctx, silenced)
		assert.Error(t, err)

	})
}

func TestSilencedStorageWithExpire(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		silenced := types.FixtureSilenced("subscription:checkname")
		silenced.Namespace = "default"
		silenced.Expire = 15
		silenced.ExpireAt = 0
		ctx := context.WithValue(context.Background(), types.NamespaceKey, silenced.Namespace)

		err := store.UpdateSilencedEntry(ctx, silenced)
		if err != nil {
			t.Fatalf("failed to update entry due to error: %s", err)
		}

		// Get silenced entry and check that expire time is not zero
		entry, err := store.GetSilencedEntryByName(ctx, silenced.Name)
		require.NoError(t, err)
		require.NotNil(t, entry)
		assert.NotZero(t, entry.Expire)
	})
}

func TestSilencedStorageWithBegin(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		silenced := types.FixtureSilenced("subscription:checkname")
		silenced.Namespace = "default"
		// set a begin time in the future
		silenced.Begin = time.Now().Add(time.Duration(1) * time.Second).Unix()
		// current time is before the start time
		currentTime := time.Now().Unix()
		ctx := context.WithValue(context.Background(), types.NamespaceKey, silenced.Namespace)

		err := store.UpdateSilencedEntry(ctx, silenced)
		if err != nil {
			t.Fatalf("failed to update entry due to error: %s", err)
		}

		// Get silenced entry and check that it is not yet ready to start
		// silencing
		entry, err := store.GetSilencedEntryByName(ctx, silenced.Name)
		require.NoError(t, err)
		require.NotNil(t, entry)
		assert.False(t, entry.Begin < currentTime)

		// Wait for begin time to elapse current time. i.e let silencing begin
		time.Sleep(3 * time.Second)

		// reset current time to be ahead of begin time
		currentTime = time.Now().Unix()
		assert.True(t, entry.Begin < currentTime)
	})
}

func TestSilencedStorageWithBeginAndExpire(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		silenced := types.FixtureSilenced("subscription:checkname")
		silenced.Namespace = "default"
		silenced.Expire = 15
		currentTime := time.Now().UTC().Unix()
		// set a begin time in the future
		silenced.Begin = currentTime + 3600
		// current time is before the start time
		ctx := context.WithValue(context.Background(), types.NamespaceKey, silenced.Namespace)

		err := store.UpdateSilencedEntry(ctx, silenced)
		if err != nil {
			t.Fatalf("failed to update entry due to error: %s", err)
		}

		entry, err := store.GetSilencedEntryByName(ctx, silenced.Name)
		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.False(t, entry.Begin < currentTime)
		// Check that the ttl includes the expire time and delta between current
		// and begin time
		assert.Equal(t, entry.Expire, int64(15))
	})
}

func TestSilencedStorageWithMaxAllowedThresholdExpiry(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		silenced := types.FixtureSilenced("subscription:checkname")
		silenced.Namespace = "default"
		silenced.ExpireAt = time.Now().Add(time.Duration(30000) * time.Second).Unix()
		// set a begin time
		silenced.Begin = time.Now().Unix()
		ctx := context.WithValue(context.Background(), types.NamespaceKey, silenced.Namespace)

		err := store.UpdateSilencedEntry(ctx, silenced)

		// assert that error is thrown for breaching max expiry time allowed
		assert.Error(t, err)

		entry, err := store.GetSilencedEntryByName(ctx, silenced.Name)

		// assert that entry is nil
		assert.NoError(t, err)
		assert.Nil(t, entry)

	})
}

func TestSilencedStorageWithMaxAllowedThresholdExpiryWithError(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		silenced := types.FixtureSilenced("subscription:checkname")
		silenced.Namespace = "default"
		silenced.ExpireAt = 0
		silenced.Expire = 3001
		silenced.Begin = time.Now().Unix()
		ctx := context.WithValue(context.Background(), types.NamespaceKey, silenced.Namespace)

		err := store.UpdateSilencedEntry(ctx, silenced)

		// assert that error is thrown for breaching max expiry time allowed
		assert.Error(t, err)

		entry, err := store.GetSilencedEntryByName(ctx, silenced.Name)

		// assert that entry is nil
		assert.NoError(t, err)
		assert.Nil(t, entry)

	})
}

func TestSilencedStorageWithMaxAllowedThresholdExpiryAndWithoutError(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		silenced := types.FixtureSilenced("subscription:checkname")
		silenced.Namespace = "default"
		silenced.ExpireAt = 0
		silenced.Expire = 100
		silenced.Begin = time.Now().Unix()
		ctx := context.WithValue(context.Background(), types.NamespaceKey, silenced.Namespace)

		err := store.UpdateSilencedEntry(ctx, silenced)

		// assert that error is nil
		assert.Nil(t, err)

		entry, err := store.GetSilencedEntryByName(ctx, silenced.Name)

		// assert that entry is not nil
		assert.NoError(t, err)
		assert.NotNil(t, entry)

	})
}
