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

	check1 := types.FixtureCheck("check1")
	store.On("GetChecks").Return([]*types.Check{check1}, nil)
	store.On("GetCheckByName", "check1").Return(check1, nil)

	c := &Schedulerd{
		Store:           store,
		schedulersMutex: &sync.Mutex{},
		schedulers:      map[string]*CheckScheduler{},
		wg:              &sync.WaitGroup{},
	}

	assert.Equal(t, 0, len(c.schedulers))

	c.reconcile()

	assert.Equal(t, 1, len(c.schedulers))

	var nilCheck *types.Check
	store.On("GetCheckByName", "check1").Return(nilCheck, nil)
	store.On("GetChecks").Return(nil, nil)

	assert.NoError(t, c.reconcile())
}
