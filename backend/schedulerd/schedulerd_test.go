package schedulerd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/queue"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/etcdstore/testutil"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
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
	require.NoError(t, st.NamespaceStore().CreateOrUpdate(context.Background(), corev3.FixtureNamespace("default")))

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

	check := corev2.FixtureCheckConfig("check_name")
	assert.NoError(t, check.Validate())

	req := storev2.NewResourceRequestFromResource(check)
	wrappedCheck, err := wrap.Resource(check)
	if err != nil {
		assert.FailNow(t, "could not create check", err)
	}

	assert.NoError(t, st.CreateOrUpdate(context.Background(), req, wrappedCheck))
	assert.NoError(t, st.Delete(context.Background(), req))

	assert.NoError(t, sub.Cancel())
	close(tsub.ch)

	assert.NoError(t, schedulerd.Stop())
	assert.NoError(t, bus.Stop())

	for msg := range tsub.ch {
		result, ok := msg.(*corev2.CheckConfig)
		assert.True(t, ok)
		assert.EqualValues(t, check, result)
	}
}
