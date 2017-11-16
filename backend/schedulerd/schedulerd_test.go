package schedulerd

import (
	"context"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store/etcd/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestSchedulerd(t *testing.T) {
	// Setup wizard bus
	bus := &messaging.WizardBus{}
	if berr := bus.Start(); berr != nil {
		assert.FailNow(t, berr.Error())
	}

	// Setup store
	st, serr := testutil.NewStoreInstance()
	if serr != nil {
		assert.FailNow(t, serr.Error())
	}
	defer st.Teardown()

	// Mock a default organization
	st.UpdateOrganization(context.Background(), types.FixtureOrganization("default"))

	// Mock a default environment
	st.UpdateEnvironment(context.Background(), types.FixtureEnvironment("default"))

	checker := &Schedulerd{
		Store:      st,
		MessageBus: bus,
	}
	checker.Start()

	ch := make(chan interface{}, 10)
	assert.NoError(t, bus.Subscribe("subscription", "channel", ch))

	check := types.FixtureCheckConfig("check_name")
	ctx := context.WithValue(context.Background(), types.OrganizationKey, check.Organization)
	ctx = context.WithValue(ctx, types.EnvironmentKey, check.Environment)

	assert.NoError(t, check.Validate())
	assert.NoError(t, st.UpdateCheckConfig(ctx, check))

	time.Sleep(1 * time.Second)

	err := st.DeleteCheckConfigByName(ctx, check.Name)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	assert.NoError(t, checker.Stop())
	assert.NoError(t, bus.Stop())
	close(ch)

	for msg := range ch {
		result, ok := msg.(*types.CheckConfig)
		assert.True(t, ok)
		assert.EqualValues(t, check, result)
	}
}
