//go:build integration && !race
// +build integration,!race

package etcd

import (
	"context"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckConfigStorage(t *testing.T) {
	testWithEtcd(t, func(s store.Store) {
		check := corev2.FixtureCheckConfig("check1")
		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, check.Namespace)

		pred := &store.SelectionPredicate{}

		// We should receive an empty slice if no results were found
		checks, err := s.GetCheckConfigs(ctx, pred)
		assert.NoError(t, err)
		assert.NotNil(t, checks)
		assert.Empty(t, pred.Continue)

		err = s.UpdateCheckConfig(ctx, check)
		require.NoError(t, err)

		retrieved, err := s.GetCheckConfigByName(ctx, "check1")
		assert.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, check.Name, retrieved.Name)
		assert.Equal(t, check.Interval, retrieved.Interval)
		assert.Equal(t, check.Subscriptions, retrieved.Subscriptions)
		assert.Equal(t, check.Command, retrieved.Command)
		assert.Equal(t, check.Stdin, retrieved.Stdin)

		checks, err = s.GetCheckConfigs(ctx, pred)
		assert.NoError(t, err)
		assert.NotEmpty(t, checks)
		assert.Equal(t, 1, len(checks))
		assert.Empty(t, pred.Continue)

		// Updating a check in a nonexistent org and env should not work
		check.Namespace = "missing"
		err = s.UpdateCheckConfig(ctx, check)
		assert.Error(t, err)
	})
}

func TestCheckConfigSchedulerProperty(t *testing.T) {
	testWithEtcd(t, func(s store.Store) {
		check := corev2.FixtureCheckConfig("interval")
		rrCheck := corev2.FixtureCheckConfig("roundrobin-interval")
		rrCheck.RoundRobin = true

		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, check.Namespace)
		pred := &store.SelectionPredicate{}

		if err := s.UpdateCheckConfig(ctx, check); err != nil {
			t.Fatal(err)
		}

		if err := s.UpdateCheckConfig(ctx, rrCheck); err != nil {
			t.Fatal(err)
		}

		checks, err := s.GetCheckConfigs(ctx, pred)
		if err != nil {
			t.Fatal(err)
		}

		for _, check := range checks {
			if check.Scheduler == "" {
				t.Errorf("expected non-zero scheduler for %q", check.Name)
			}
		}

		chk, err := s.GetCheckConfigByName(ctx, "interval")
		if err != nil {
			t.Fatal(err)
		}

		if got, want := chk.Scheduler, corev2.MemoryScheduler; got != want {
			t.Errorf("bad scheduler: got %q, want %q", got, want)
		}

		chk, err = s.GetCheckConfigByName(ctx, "roundrobin-interval")
		if err != nil {
			t.Fatal(err)
		}

		if got, want := chk.Scheduler, corev2.EtcdScheduler; got != want {
			t.Errorf("bad scheduler: got %q, want %q", got, want)
		}

	})
}
