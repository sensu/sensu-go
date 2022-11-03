//go:build integration && !race
// +build integration,!race

package etcd

import (
	"context"
	"testing"

	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityStorage(t *testing.T) {
	testWithEtcd(t, func(s store.Store) {
		entity := types.FixtureEntity("entity")
		ctx := context.WithValue(context.Background(), types.NamespaceKey, entity.Namespace)
		pred := &store.SelectionPredicate{}

		// We should receive an empty slice if no results were found
		entities, err := s.GetEntities(ctx, pred)
		assert.NoError(t, err)
		assert.NotNil(t, entities)
		assert.Equal(t, pred.Continue, "")

		err = s.UpdateEntity(ctx, entity)
		assert.NoError(t, err)

		retrieved, err := s.GetEntityByName(ctx, entity.Name)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, entity.Name, retrieved.Name)

		entities, err = s.GetEntities(ctx, pred)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(entities))
		assert.Equal(t, entity.Name, entities[0].Name)
		assert.Equal(t, pred.Continue, "")

		err = s.DeleteEntity(ctx, entity)
		assert.NoError(t, err)

		retrieved, err = s.GetEntityByName(ctx, entity.Name)
		assert.Nil(t, retrieved)
		assert.NoError(t, err)

		// Nonexistent entity deletion should return no error.
		err = s.DeleteEntity(ctx, entity)
		assert.NoError(t, err)

		// Updating an enity in a nonexistent namespace should not work
		entity.Namespace = "missing"
		err = s.UpdateEntity(ctx, entity)
		assert.Error(t, err)
	})
}

func TestEntityIteration(t *testing.T) {
	configs := []corev3.EntityConfig{
		*corev3.FixtureEntityConfig("a"),
		*corev3.FixtureEntityConfig("b"),
		*corev3.FixtureEntityConfig("c"),
		*corev3.FixtureEntityConfig("d"),
		*corev3.FixtureEntityConfig("e"),
	}
	states := []corev3.EntityState{
		*corev3.FixtureEntityState("b"),
		*corev3.FixtureEntityState("c"),
		*corev3.FixtureEntityState("d"),
	}
	entities, err := entitiesFromConfigAndState(configs, states)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(entities), len(configs); got != want {
		t.Fatalf("bad entity count: got %d, want %d", got, want)
	}
	for i := range configs {
		if got, want := configs[i].Metadata.Name, entities[i].Name; got != want {
			t.Errorf("bad entity name: got %q, want %q", got, want)
		}
	}
}

func TestEntityIterationNoPanicMismatched(t *testing.T) {
	configs := []corev3.EntityConfig{
		*corev3.FixtureEntityConfig("b"),
		*corev3.FixtureEntityConfig("c"),
	}
	states := []corev3.EntityState{
		*corev3.FixtureEntityState("a"),
		*corev3.FixtureEntityState("b"),
		*corev3.FixtureEntityState("c"),
	}
	if _, err := entitiesFromConfigAndState(configs, states); err != nil {
		t.Fatal(err)
	}
}
