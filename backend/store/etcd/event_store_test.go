package etcd

import (
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestEventStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		// We should receive an empty slice if no results were found
		events, err := store.GetEvents("default")
		assert.NoError(t, err)
		assert.NotNil(t, events)

		event := types.FixtureEvent("entity1", "check1")
		assert.NoError(t, store.UpdateEvent(event))

		newEv, err := store.GetEventByEntityCheck("default", "entity1", "check1")
		assert.NoError(t, err)
		assert.EqualValues(t, event, newEv)

		events, err = store.GetEvents("default")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events))
		assert.EqualValues(t, event, events[0])

		newEv, err = store.GetEventByEntityCheck("default", "", "foo")
		assert.Nil(t, newEv)
		assert.Error(t, err)

		newEv, err = store.GetEventByEntityCheck("default", "foo", "")
		assert.Nil(t, newEv)
		assert.Error(t, err)

		newEv, err = store.GetEventByEntityCheck("default", "foo", "foo")
		assert.Nil(t, newEv)
		assert.Nil(t, err)

		events, err = store.GetEventsByEntity("default", "entity1")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events))
		assert.EqualValues(t, event, events[0])

		assert.NoError(t, store.DeleteEventByEntityCheck("default", "entity1", "check1"))
		newEv, err = store.GetEventByEntityCheck("default", "entity1", "check1")
		assert.Nil(t, newEv)
		assert.NoError(t, err)

		assert.Error(t, store.DeleteEventByEntityCheck("", "", ""))
		assert.Error(t, store.DeleteEventByEntityCheck("default", "", "foo"))
		assert.Error(t, store.DeleteEventByEntityCheck("default", "foo", ""))
	})
}
