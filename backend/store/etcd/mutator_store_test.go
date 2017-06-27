package etcd

import (
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestMutatorStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		// We should receive an empty slice if no results were found
		mutators, err := store.GetMutators("default")
		assert.NoError(t, err)
		assert.NotNil(t, mutators)

		mutator := types.FixtureMutator("mutator1")

		err = store.UpdateMutator(mutator)
		assert.NoError(t, err)
		retrieved, err := store.GetMutatorByName("default", "mutator1")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)

		assert.Equal(t, mutator.Name, retrieved.Name)
		assert.Equal(t, mutator.Command, retrieved.Command)
		assert.Equal(t, mutator.Timeout, retrieved.Timeout)

		mutators, err = store.GetMutators("default")
		assert.NoError(t, err)
		assert.NotEmpty(t, mutators)
		assert.Equal(t, 1, len(mutators))
	})
}
