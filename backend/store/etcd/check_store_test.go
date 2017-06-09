package etcd

import (
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestCheckConfigStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		// We should receive an empty slice if no results were found
		checks, err := store.GetCheckConfigs()
		assert.NoError(t, err)
		assert.NotNil(t, checks)

		check := &types.CheckConfig{
			Name:          "check1",
			Interval:      60,
			Subscriptions: []string{"subscription1"},
			Command:       "command1",
		}

		err = store.UpdateCheckConfig(check)
		assert.NoError(t, err)
		retrieved, err := store.GetCheckConfigByName("check1")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)

		assert.Equal(t, check.Name, retrieved.Name)
		assert.Equal(t, check.Interval, retrieved.Interval)
		assert.Equal(t, check.Subscriptions, retrieved.Subscriptions)
		assert.Equal(t, check.Command, retrieved.Command)

		checks, err = store.GetCheckConfigs()
		assert.NoError(t, err)
		assert.NotEmpty(t, checks)
		assert.Equal(t, 1, len(checks))
	})
}
