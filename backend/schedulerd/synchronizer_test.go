package schedulerd

import (
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSynchronizeChecks(t *testing.T) {
	assert := assert.New(t)

	check1 := types.FixtureCheckConfig("check1")
	store := &mockstore.MockStore{}
	store.On("GetCheckConfigs", mock.AnythingOfType("*context.emptyCtx")).Return([]*types.CheckConfig{check1}, nil)

	sync := SynchronizeChecks{
		Store: store,
		OnUpdate: func(res []*types.CheckConfig) {
			assert.NotEmpty(res)
			assert.Len(res, 1)
		},
	}
	require.NoError(t, sync.Sync())
}

func TestSynchronizeAssets(t *testing.T) {
	assert := assert.New(t)

	asset := types.FixtureAsset("asset1")
	store := &mockstore.MockStore{}
	store.On("GetAssets", mock.AnythingOfType("*context.emptyCtx")).Return([]*types.Asset{asset}, nil)

	sync := SynchronizeAssets{
		Store: store,
		OnUpdate: func(res []*types.Asset) {
			assert.NotEmpty(res)
			assert.Len(res, 1)
		},
	}
	require.NoError(t, sync.Sync())
}

func TestSynchronizeHooks(t *testing.T) {
	assert := assert.New(t)

	hook := types.FixtureHookConfig("hook1")
	store := &mockstore.MockStore{}
	store.On("GetHookConfigs", mock.AnythingOfType("*context.emptyCtx")).Return([]*types.HookConfig{hook}, nil)

	sync := SynchronizeHooks{
		Store: store,
		OnUpdate: func(res []*types.HookConfig) {
			assert.NotEmpty(res)
			assert.Len(res, 1)
		},
	}
	require.NoError(t, sync.Sync())
}

func TestSynchronizeEntities(t *testing.T) {
	assert := assert.New(t)

	entity := types.FixtureEntity("entity1")
	store := &mockstore.MockStore{}
	store.On("GetEntities", mock.AnythingOfType("*context.emptyCtx")).Return([]*types.Entity{entity}, nil)

	sync := SynchronizeEntities{
		Store: store,
		OnUpdate: func(res []*types.Entity) {
			assert.NotEmpty(res)
			assert.Len(res, 1)
		},
	}
	require.NoError(t, sync.Sync())
}

func TestSyncScheduler(t *testing.T) {
	assert := assert.New(t)

	scheduler := NewSynchronizeStateScheduler(30)
	scheduler.Start()

	err := scheduler.Stop()
	assert.NoError(err)
}
