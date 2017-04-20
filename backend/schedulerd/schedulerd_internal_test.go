package schedulerd

import (
	"sync"
	"testing"

	"github.com/sensu/sensu-go/testing/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestSchedulerdReconcile(t *testing.T) {
	store := fixtures.NewFixtureStore()

	c := &Schedulerd{
		Store:           store,
		schedulersMutex: &sync.Mutex{},
		schedulers:      map[string]*CheckScheduler{},
		wg:              &sync.WaitGroup{},
	}

	assert.Equal(t, 0, len(c.schedulers))

	c.reconcile()

	assert.Equal(t, 1, len(c.schedulers))

	store.DeleteCheckByName("check1")

	assert.NoError(t, c.reconcile())
}
