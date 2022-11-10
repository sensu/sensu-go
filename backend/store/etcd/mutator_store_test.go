//go:build integration && !race
// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev2 "github.com/sensu/core/v2"
)

func TestMutatorStorage(t *testing.T) {
	testWithEtcd(t, func(s store.Store) {
		mutator := corev2.FixtureMutator("mutator1")
		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, mutator.Namespace)

		// We should receive an empty slice if no results were found
		pred := &store.SelectionPredicate{}
		mutators, err := s.GetMutators(ctx, pred)
		assert.NoError(t, err)
		assert.NotNil(t, mutators)
		assert.Empty(t, pred.Continue)

		err = s.UpdateMutator(ctx, mutator)
		assert.NoError(t, err)

		retrieved, err := s.GetMutatorByName(ctx, "mutator1")
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, mutator.Name, retrieved.Name)
		assert.Equal(t, mutator.Command, retrieved.Command)
		assert.Equal(t, mutator.Timeout, retrieved.Timeout)

		mutators, err = s.GetMutators(ctx, pred)
		assert.NoError(t, err)
		assert.NotEmpty(t, mutators)
		assert.Equal(t, 1, len(mutators))
		assert.Empty(t, pred.Continue)

		// Updating a mutator in a nonexistent org and env should not work
		mutator.Namespace = "missing"
		err = s.UpdateMutator(ctx, mutator)
		assert.Error(t, err)
	})
}
