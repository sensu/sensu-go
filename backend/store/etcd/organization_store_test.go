package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestOrgStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		ctx := context.Background()

		// We should receive the default organization (set in store_test.go)
		orgs, err := store.GetOrganizations(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(orgs))

		org := types.FixtureOrganization("acme")
		err = store.UpdateOrganization(ctx, org)
		assert.NoError(t, err)

		result, err := store.GetOrganizationByName(ctx, org.Name)
		assert.NoError(t, err)
		assert.Equal(t, org.Name, result.Name)

		// Missing organization
		_, err = store.GetOrganizationByName(ctx, "missing")
		assert.Error(t, err)

		// Create an environment within this new organization
		env := types.FixtureEnvironment("dev")
		err = store.UpdateEnvironment(ctx, org.Name, env)
		assert.NoError(t, err)

		// Get all organizations
		orgs, err = store.GetOrganizations(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, orgs)
		assert.Equal(t, 2, len(orgs))

		// Delete an org
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
