// +build integration

package schedulerd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCheckWatcherSmoke(t *testing.T) {
	st := &mockstore.MockStore{}

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, bus.Start())
	defer bus.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	checkA := types.FixtureCheckConfig("a")
	checkB := types.FixtureCheckConfig("b")
	st.On("GetCheckConfigs", mock.Anything, &store.SelectionPredicate{}).Return([]*types.CheckConfig{checkA, checkB}, nil)
	st.On("GetCheckConfigByName", mock.Anything, "a").Return(checkA, nil)
	st.On("GetCheckConfigByName", mock.Anything, "b").Return(checkB, nil)
	st.On("GetAssets", mock.Anything, &store.SelectionPredicate{}).Return([]*types.Asset{}, nil)
	st.On("GetHookConfigs", mock.Anything, &store.SelectionPredicate{}).Return([]*types.HookConfig{}, nil)

	watcherChan := make(chan store.WatchEventCheckConfig)
	st.On("GetCheckConfigWatcher", mock.Anything).Return((<-chan store.WatchEventCheckConfig)(watcherChan), nil)

	watcher := NewCheckWatcher(ctx, bus, st, nil, &EntityCache{})
	require.NoError(t, watcher.Start())

	checkAA := types.FixtureCheckConfig("a")
	checkAA.Interval = 5
	watcherChan <- store.WatchEventCheckConfig{
		CheckConfig: checkAA,
		Action:      store.WatchUpdate,
	}

	checkB.Cron = "* * * * *"
	watcherChan <- store.WatchEventCheckConfig{
		CheckConfig: checkB,
		Action:      store.WatchUpdate,
	}

	watcherChan <- store.WatchEventCheckConfig{
		CheckConfig: checkAA,
		Action:      store.WatchDelete,
	}

	watcherChan <- store.WatchEventCheckConfig{
		CheckConfig: checkB,
		Action:      store.WatchCreate,
	}
}
