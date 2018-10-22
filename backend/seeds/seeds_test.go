package seeds

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store/etcd/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeedInitialData(t *testing.T) {
	// Setup store
	ctx := context.Background()
	storeInstance, serr := testutil.NewStoreInstance()
	if serr != nil {
		assert.FailNow(t, serr.Error())
	}
	defer storeInstance.Teardown()

	store := storeInstance.GetStore()
	storev2 := storeInstance.GetStoreV2()

	err := SeedInitialData(store, storev2)
	require.NoError(t, err, "seeding process should not raise an error")

	err = SeedInitialData(store, storev2)
	require.NoError(t, err, "seeding process should be able to be run more than once without error")

	admin, err := store.GetUser(ctx, "admin")
	require.NoError(t, err)
	assert.NotEmpty(t, admin, "admin user should be present after seed process")

	agent, err := store.GetUser(ctx, "agent")
	require.NoError(t, err)
	assert.NotEmpty(t, agent, "agent user should be present after seed process")

	defaultNamespace, err := store.GetNamespace(ctx, "default")
	require.NoError(t, err)
	assert.NotEmpty(t, defaultNamespace, "default namespace should be present after seed process")

	sensu, err := store.GetUser(ctx, "sensu")
	require.NoError(t, err)
	assert.NotEmpty(t, sensu, "sensu user should be present after seed process")
}
