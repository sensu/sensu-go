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
	st, serr := testutil.NewStoreInstance()
	if serr != nil {
		assert.FailNow(t, serr.Error())
	}
	defer st.Teardown()

	err := SeedInitialData(st)
	require.NoError(t, err, "seeding process should not raise an error")

	err = SeedInitialData(st)
	require.NoError(t, err, "seeding process should be able to be run more than once without error")

	admin, err := st.GetUser(ctx, "admin")
	require.NoError(t, err)
	assert.NotEmpty(t, admin, "admin user should be present after seed process")

	adminRole, err := st.GetRoleByName(ctx, "admin")
	require.NoError(t, err)
	assert.NotEmpty(t, adminRole, "admin role should be present after seed process")

	agent, err := st.GetUser(ctx, "agent")
	require.NoError(t, err)
	assert.NotEmpty(t, agent, "agent user should be present after seed process")

	defaultNamespace, err := st.GetNamespace(ctx, "default")
	require.NoError(t, err)
	assert.NotEmpty(t, defaultNamespace, "default namespace should be present after seed process")

	sensu, err := st.GetUser(ctx, "sensu")
	require.NoError(t, err)
	assert.NotEmpty(t, sensu, "sensu user should be present after seed process")

	readOnlyRole, err := st.GetRoleByName(ctx, "read-only")
	require.NoError(t, err)
	assert.NotEmpty(t, readOnlyRole, "read-only role should be present after seed process")
}
