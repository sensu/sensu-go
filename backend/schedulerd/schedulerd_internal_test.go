package schedulerd

import (
	"sync"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestSchedulerdReconcile(t *testing.T) {
	store := &mockstore.MockStore{}

	check1 := types.FixtureCheckConfig("check1")
	store.On("GetCheckConfigs", "").Return([]*types.CheckConfig{check1}, nil)
	store.On("GetCheckConfigsByName", "default", "check1").Return(check1, nil)

	c := &Schedulerd{Store: store, wg: &sync.WaitGroup{}}
	c.schedulers = newSchedulerCollection(nil, store)

	assert.Equal(t, 0, len(c.schedulers.items))

	c.reconcile()

	assert.Equal(t, 1, len(c.schedulers.items))

	var nilCheck *types.Check
	store.On("GetCheckConfigByName", "default", "check1").Return(nilCheck, nil)
	store.On("GetCheckConfigs", "").Return(nil, nil)

	assert.NoError(t, c.reconcile())
}
