package etcd

import (
	"context"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatchers(t *testing.T) {
	t.Parallel()

	testWithEtcd(t, func(st store.Store) {
		checkCfg := types.FixtureCheckConfig("check_config")

		ctx, cancel := context.WithCancel(context.Background())

		watchChan := st.GetCheckConfigWatcher(ctx)
		require.NotNil(t, watchChan)

		if err := st.UpdateCheckConfig(ctx, checkCfg); err != nil {
			require.NoError(t, err, "failed to create check config in store")
		}

		var ev store.WatchEventCheckConfig

		select {
		case ev = <-watchChan:
			assert.Equal(t, store.WatchCreate, ev.Action)
			assert.Equal(t, checkCfg.Organization, ev.CheckConfig.Organization)
			assert.Equal(t, checkCfg.Environment, ev.CheckConfig.Environment)
			assert.Equal(t, checkCfg.Name, ev.CheckConfig.Name)
		case <-time.After(5 * time.Second):
			assert.Fail(t, "failed to receive a watch event in 5 seconds")
		}

		cancel()

		select {
		case _, ok := <-watchChan:
			assert.False(t, ok, "watch channel wasn't closed")
		case <-time.After(5 * time.Second):
			assert.Fail(t, "failed to close watch channel in 5 seconds")
		}
	})
}
