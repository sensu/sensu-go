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

func TestOrgStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		ctx := context.Background()

		// We should receive the default organization (set in store_test.go)
		orgs, err := store.GetOrganizations(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(orgs))

		// We should be able to create a new organization
		org := types.FixtureOrganization("acme")
		err = store.CreateOrganization(ctx, org)
		assert.NoError(t, err)

		result, err := store.GetOrganizationByName(ctx, org.Name)
		assert.NoError(t, err)
		assert.Equal(t, org.Name, result.Name)

		// Missing organization
		result, err = store.GetOrganizationByName(ctx, "missing")
		assert.NoError(t, err)
		assert.Nil(t, result)

		// Create an environment within this new organization
		env := types.FixtureEnvironment("dev")
		env.Organization = org.Name
		err = store.UpdateEnvironment(ctx, env)
		assert.NoError(t, err)

		// Get all organizations
		orgs, err = store.GetOrganizations(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, orgs)
		assert.Equal(t, 2, len(orgs))

		// Delete a non-empty org
		err = store.DeleteOrganizationByName(ctx, org.Name)
		assert.Error(t, err)

		// Delete a non-empty org w/ roles
		require.NoError(t, store.DeleteEnvironment(ctx, env))
		require.NoError(t, store.UpdateRole(ctx, types.FixtureRole("1", org.Name, env.Name)))
		require.NoError(t, store.UpdateRole(ctx, types.FixtureRole("2", "asdf", "asdf")))
		err = store.DeleteOrganizationByName(ctx, org.Name)
		assert.Error(t, err)

		// Delete an empty org
		require.NoError(t, store.DeleteRoleByName(ctx, "1"))
		err = store.DeleteOrganizationByName(ctx, org.Name)
		assert.NoError(t, err)

		// Delete a missing org
		err = store.DeleteOrganizationByName(ctx, "missing")
		assert.Error(t, err)

		// Get again all organizations
		orgs, err = store.GetOrganizations(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(orgs))
	})
}
