//go:build integration && !race
// +build integration,!race

package schedulerd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/backend/store/etcd/testutil"
	"github.com/sensu/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchedulerd(t *testing.T) {
	// Setup wizard bus
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	if berr := bus.Start(); berr != nil {
		assert.FailNow(t, berr.Error())
	}

	// Setup store
	st, serr := testutil.NewStoreInstance()
	if serr != nil {
		assert.FailNow(t, serr.Error())
	}
	defer st.Teardown()

	// Mock a default namespace
	require.NoError(t, st.CreateNamespace(context.Background(), types.FixtureNamespace("default")))

	schedulerd, err := New(context.Background(), Config{
		Store:       st,
		QueueGetter: queue.NewMemoryGetter(),
		Bus:         bus,
		Client:      st.Client,
	})
	require.NoError(t, err)
	require.NoError(t, schedulerd.Start())

	tsub := testSubscriber{
		ch: make(chan interface{}, 10),
	}
	sub, err := bus.Subscribe("subscription", "testSubscriber", tsub)
	if err != nil {
		assert.FailNow(t, "could not subscribe", err)
	}

	check := types.FixtureCheckConfig("check_name")
	ctx := context.WithValue(context.Background(), types.NamespaceKey, check.Namespace)

	assert.NoError(t, check.Validate())
	assert.NoError(t, st.UpdateCheckConfig(ctx, check))

	require.NoError(t, st.DeleteCheckConfigByName(ctx, check.Name))

	assert.NoError(t, sub.Cancel())
	close(tsub.ch)

	assert.NoError(t, schedulerd.Stop())
	assert.NoError(t, bus.Stop())

	for msg := range tsub.ch {
		result, ok := msg.(*types.CheckConfig)
		assert.True(t, ok)
		assert.EqualValues(t, check, result)
	}
}
