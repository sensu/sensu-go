package backend

import (
	"sync"
	"testing"

	"github.com/sensu/sensu-go/testing/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestCheckerReconcile(t *testing.T) {
	store := fixtures.NewFixtureStore()

	c := &Checker{
		Store:           store,
		schedulersMutex: &sync.Mutex{},
		schedulers:      map[string]*CheckScheduler{},
	}

	assert.Equal(t, 0, len(c.schedulers))

	c.reconcile()

	assert.Equal(t, 1, len(c.schedulers))

	store.DeleteCheckByName("check1")

	assert.NoError(t, c.reconcile())
}
