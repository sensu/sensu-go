package schedulerd

import (
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestSyncronizeChecks(t *testing.T) {
	assert := assert.New(t)

	check1 := types.FixtureCheckConfig("check1")
	store := &mockstore.MockStore{}
	store.On("GetCheckConfigs", "").Return([]*types.CheckConfig{check1}, nil)

	sync := SyncronizeChecks{
		Store: store,
		OnUpdate: func(res []*types.CheckConfig) {
			assert.NotEmpty(res)
			assert.Len(res, 1)
		},
	}
	sync.Sync()
}

func TestSyncronizeAssets(t *testing.T) {
	assert := assert.New(t)

	asset := types.FixtureAsset("asset1")
	store := &mockstore.MockStore{}
	store.On("GetAssets", "").Return([]*types.Asset{asset}, nil)

	sync := SyncronizeAssets{
		Store: store,
		OnUpdate: func(res []*types.Asset) {
			assert.NotEmpty(res)
			assert.Len(res, 1)
		},
	}
	sync.Sync()
}

func TestSyncScheduler(t *testing.T) {
	assert := assert.New(t)

	scheduler := &SyncResourceScheduler{Interval: 30}
	scheduler.Start()

	err := scheduler.Stop()
	assert.NoError(err)
}

// 	check1 := types.FixtureCheckConfig("check1")
// 	store.On("GetCheckConfigs", "").Return([]*types.CheckConfig{check1}, nil)
//
// 	c := &Schedulerd{Store: store, wg: &sync.WaitGroup{}}
// 	c.schedulers = newSchedulerCollection(nil, store)
//
// 	assert.Equal(t, 0, len(c.schedulers.items))
//
// 	c.reconcile()
//
// 	assert.Equal(t, 1, len(c.schedulers.items))
//
// 	var nilCheck *types.Check
// 	store.On("GetCheckConfigByName", "default", "check1").Return(nilCheck, nil)
// 	store.On("GetCheckConfigs", "").Return(nil, nil)
//
// 	assert.NoError(t, c.reconcile())
// }
