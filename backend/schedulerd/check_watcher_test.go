package schedulerd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	cachev2 "github.com/sensu/sensu-go/backend/store/cache/v2"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
)

func TestCheckWatcherSmoke(t *testing.T) {
	t.Skip("skip")
	st := &mockstore.V2MockStore{}

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, bus.Start())
	defer func() {
		require.NoError(t, bus.Stop())
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	checkA := corev2.FixtureCheckConfig("a")
	checkB := corev2.FixtureCheckConfig("b")

	st.On("List", mock.Anything, mock.MatchedBy(isCheckResourceRequest), &store.SelectionPredicate{}).
		Return(mockstore.WrapList[*corev2.CheckConfig]{checkA, checkB}, nil)

	watcherChan := make(chan []storev2.WatchEvent)
	st.On("Watch", mock.Anything, mock.Anything).Return((<-chan []storev2.WatchEvent)(watcherChan), nil)

	pm := secrets.NewProviderManager(&mockEventReceiver{})
	cache, err := cachev2.New[*corev3.EntityConfig](ctx, st, true)
	if err != nil {
		t.Fatal(err)
	}
	watcher := NewCheckWatcher(ctx, bus, st, nil, cache, pm)
	require.NoError(t, watcher.Start())

	checkAA := corev2.FixtureCheckConfig("a")
	checkAA.Interval = 5
	reqCheckAA := storev2.NewResourceRequestFromResource(checkAA)
	wrappedCheckAA, err := storev2.WrapResource(checkAA)
	require.NoError(t, err)

	checkBB := corev2.FixtureCheckConfig("b")
	checkBB.Interval = 0
	checkBB.Cron = "* * * * *"
	reqCheckBB := storev2.NewResourceRequestFromResource(checkBB)
	wrappedCheckBB, err := storev2.WrapResource(checkBB)
	require.NoError(t, err)

	watcherChan <- []storev2.WatchEvent{
		{
			Type:  storev2.WatchUpdate,
			Key:   reqCheckAA,
			Value: wrappedCheckAA,
		},
	}

	watcherChan <- []storev2.WatchEvent{
		{
			Type:  storev2.WatchUpdate,
			Key:   reqCheckBB,
			Value: wrappedCheckBB,
		},
	}

	watcherChan <- []storev2.WatchEvent{
		{
			Type:  storev2.WatchDelete,
			Key:   reqCheckAA,
			Value: wrappedCheckAA,
		},
	}

	watcherChan <- []storev2.WatchEvent{
		{
			Type:  storev2.WatchCreate,
			Key:   reqCheckBB,
			Value: wrappedCheckBB,
		},
	}
}
