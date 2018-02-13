// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		perr := types.FixtureError("name", "ms")
		ctx := context.Background()
		ctx = context.WithValue(ctx, types.OrganizationKey, perr.GetOrganization())
		ctx = context.WithValue(ctx, types.EnvironmentKey, perr.GetEnvironment())

		// Storage
		cerr := store.CreateError(ctx, perr)
		require.NoError(t, cerr)

		// GetErrors
		allErrors, err := store.GetErrors(ctx)
		require.NoError(t, err)
		assert.Len(t, allErrors, 1)
		assert.EqualValues(t, []*types.Error{perr}, allErrors)

		// GetErrorsByEntity
		entityErrors, err := store.GetErrorsByEntity(ctx, perr.Event.Entity.ID)
		require.NoError(t, err)
		assert.Len(t, entityErrors, 1)
		assert.EqualValues(t, []*types.Error{perr}, entityErrors)

		// GetErrorsByCheck
		checkErrors, err := store.GetErrorsByEntityCheck(ctx, perr.Event.Entity.ID, perr.Event.Check.Name)
		require.NoError(t, err)
		assert.Len(t, checkErrors, 1)
		assert.EqualValues(t, []*types.Error{perr}, checkErrors)

		// GetError
		perrRes, err := store.GetError(ctx, perr.Event.Entity.ID, perr.Event.Check.Name, string(perr.Timestamp))
		require.NoError(t, err)
		assert.EqualValues(t, perr, perrRes)

		// Delete error
		cerr = store.CreateError(ctx, perr) // Ensure error is present
		require.NoError(t, cerr)
		err = store.DeleteError(ctx, perr.Event.Entity.ID, perr.Event.Check.Name, string(perr.Timestamp))
		require.NoError(t, err)
		allErrors, err = store.GetErrors(ctx) // Check that error is gone
		require.NoError(t, err)
		require.Empty(t, allErrors)

		// Delete errors by check
		cerr = store.CreateError(ctx, perr) // Ensure error is present
		require.NoError(t, cerr)
		err = store.DeleteErrorsByEntityCheck(ctx, perr.Event.Entity.ID, perr.Event.Check.Name)
		require.NoError(t, err)
		allErrors, err = store.GetErrors(ctx) // Check that error is gone
		require.NoError(t, err)
		require.Empty(t, allErrors)

		// Delete errors by entity
		cerr = store.CreateError(ctx, perr) // Ensure error is present
		require.NoError(t, cerr)
		err = store.DeleteErrorsByEntity(ctx, perr.Event.Entity.ID)
		require.NoError(t, err)
		allErrors, err = store.GetErrors(ctx) // Check that error is gone
		require.NoError(t, err)
		require.Empty(t, allErrors)
	})
}
