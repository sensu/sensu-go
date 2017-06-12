package etcd

import (
	"fmt"
	"log"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/util"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func testWithEtcd(t *testing.T, f func(store.Store)) {
	util.WithTempDir(func(tmpDir string) {
		ports := make([]int, 2)
		err := util.RandomPorts(ports)
		if err != nil {
			log.Panic(err)
		}
		clURL := fmt.Sprintf("http://127.0.0.1:%d", ports[0])
		apURL := fmt.Sprintf("http://127.0.0.1:%d", ports[1])
		initCluster := fmt.Sprintf("default=%s", apURL)

		cfg := NewConfig()
		cfg.StateDir = tmpDir
		cfg.ClientListenURL = clURL
		cfg.PeerListenURL = apURL
		cfg.InitialCluster = initCluster

		e, err := NewEtcd(cfg)
		assert.NoError(t, err)
		if e != nil {
			defer e.Shutdown()
		}

		s, err := e.NewStore()
		assert.NoError(t, err)
		if err != nil {
			assert.FailNow(t, "failed to get store from etcd")
		}

		f(s)
	})
}

func TestHandlerStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		// We should receive an empty slice if no results were found
		handlers, err := store.GetHandlers("default")
		assert.NoError(t, err)
		assert.NotNil(t, handlers)

		handler := types.FixtureHandler("handler1")

		err = store.UpdateHandler(handler)
		assert.NoError(t, err)

		retrieved, err := store.GetHandlerByName("default", "handler1")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)

		assert.Equal(t, handler.Name, retrieved.Name)
		assert.Equal(t, handler.Type, retrieved.Type)
		assert.Equal(t, handler.Command, retrieved.Command)
		assert.Equal(t, handler.Timeout, retrieved.Timeout)

		handlers, err = store.GetHandlers("default")
		assert.NoError(t, err)
		assert.NotEmpty(t, handlers)
		assert.Equal(t, 1, len(handlers))
	})
}

func TestMutatorStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		// We should receive an empty slice if no results were found
		mutators, err := store.GetMutators()
		assert.NoError(t, err)
		assert.NotNil(t, mutators)

		mutator := &types.Mutator{
			Name:    "mutator1",
			Command: "command1",
			Timeout: 10,
		}

		err = store.UpdateMutator(mutator)
		assert.NoError(t, err)
		retrieved, err := store.GetMutatorByName("mutator1")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)

		assert.Equal(t, mutator.Name, retrieved.Name)
		assert.Equal(t, mutator.Command, retrieved.Command)
		assert.Equal(t, mutator.Timeout, retrieved.Timeout)

		mutators, err = store.GetMutators()
		assert.NoError(t, err)
		assert.NotEmpty(t, mutators)
		assert.Equal(t, 1, len(mutators))
	})
}

func TestEventStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		// We should receive an empty slice if no results were found
		events, err := store.GetEvents()
		assert.NoError(t, err)
		assert.NotNil(t, events)

		event := types.FixtureEvent("entity1", "check1")
		assert.NoError(t, store.UpdateEvent(event))

		newEv, err := store.GetEventByEntityCheck(event.Entity.ID, event.Check.Config.Name)
		assert.NoError(t, err)
		assert.EqualValues(t, event, newEv)

		events, err = store.GetEvents()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events))
		assert.EqualValues(t, event, events[0])

		newEv, err = store.GetEventByEntityCheck("", "foo")
		assert.Nil(t, newEv)
		assert.Error(t, err)

		newEv, err = store.GetEventByEntityCheck("foo", "")
		assert.Nil(t, newEv)
		assert.Error(t, err)

		newEv, err = store.GetEventByEntityCheck("foo", "foo")
		assert.Nil(t, newEv)
		assert.Nil(t, err)

		events, err = store.GetEventsByEntity(event.Entity.ID)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events))
		assert.EqualValues(t, event, events[0])

		assert.NoError(t, store.DeleteEventByEntityCheck(event.Entity.ID, event.Check.Config.Name))
		newEv, err = store.GetEventByEntityCheck(event.Entity.ID, event.Check.Config.Name)
		assert.Nil(t, newEv)
		assert.NoError(t, err)

		assert.Error(t, store.DeleteEventByEntityCheck("", ""))
		assert.Error(t, store.DeleteEventByEntityCheck("", "foo"))
		assert.Error(t, store.DeleteEventByEntityCheck("foo", ""))
	})
}
