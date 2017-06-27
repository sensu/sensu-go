package etcd

import (
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestOrgStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		// We should receive the default organization (set in store_test.go)
		orgs, err := store.GetOrganizations()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(orgs))

		org := types.FixtureOrganization("foo")
		err = store.UpdateOrganization(org)
		assert.NoError(t, err)

		result, err := store.GetOrganizationByName(org.Name)
		assert.NoError(t, err)
		assert.Equal(t, org.Name, result.Name)

		// Missing organization
		_, err = store.GetOrganizationByName("foobar")
		assert.Error(t, err)

		// Get all organizations
		orgs, err = store.GetOrganizations()
		assert.NoError(t, err)
		assert.NotEmpty(t, orgs)
		assert.Equal(t, 2, len(orgs))

		// Delete an org
		err = store.DeleteOrganizationByName(org.Name)
		assert.NoError(t, err)

		// Delete a missing org
		err = store.DeleteOrganizationByName("foobar")
		assert.Error(t, err)

		// Get again all organizations
		orgs, err = store.GetOrganizations()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(orgs))
	})
}
