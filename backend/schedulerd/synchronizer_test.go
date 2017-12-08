package schedulerd

import (
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSyncronizeChecks(t *testing.T) {
	assert := assert.New(t)

	check1 := types.FixtureCheckConfig("check1")
	store := &mockstore.MockStore{}
	store.On("GetCheckConfigs", mock.AnythingOfType("*context.emptyCtx")).Return([]*types.CheckConfig{check1}, nil)

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
	store.On("GetAssets", mock.AnythingOfType("*context.emptyCtx")).Return([]*types.Asset{asset}, nil)

	sync := SyncronizeAssets{
		Store: store,
		OnUpdate: func(res []*types.Asset) {
			assert.NotEmpty(res)
			assert.Len(res, 1)
		},
	}
	sync.Sync()
}

func TestSyncronizeHooks(t *testing.T) {
	assert := assert.New(t)

	hook := types.FixtureHookConfig("hook1")
	store := &mockstore.MockStore{}
	store.On("GetHookConfigs", mock.AnythingOfType("*context.emptyCtx")).Return([]*types.HookConfig{hook}, nil)

	sync := SyncronizeHooks{
		Store: store,
		OnUpdate: func(res []*types.HookConfig) {
			assert.NotEmpty(res)
			assert.Len(res, 1)
		},
	}
	sync.Sync()
}

func TestSyncScheduler(t *testing.T) {
	assert := assert.New(t)

	scheduler := NewSynchronizeStateScheduler(30)
	scheduler.Start()

	err := scheduler.Stop()
	assert.NoError(err)
}
