package etcd

import (
	"fmt"
	"log"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/system"
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

func TestEntityStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		entity := &types.Entity{
			ID: "0",
		}
		err := store.UpdateEntity(entity)
		assert.NoError(t, err)
		retrieved, err := store.GetEntityByID(entity.ID)
		assert.NoError(t, err)
		assert.Equal(t, entity.ID, retrieved.ID)
		entities, err := store.GetEntities()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(entities))
		assert.Equal(t, entity.ID, entities[0].ID)
		err = store.DeleteEntity(entity)
		assert.NoError(t, err)
		retrieved, err = store.GetEntityByID(entity.ID)
		assert.Nil(t, retrieved)
		assert.NoError(t, err)
		// Nonexistent entity deletion should return no error.
		err = store.DeleteEntity(entity)
		assert.NoError(t, err)
	})

}

func TestCheckStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		check := &types.Check{
			Name:          "check1",
			Interval:      60,
			Subscriptions: []string{"subscription1"},
			Command:       "command1",
		}

		err := store.UpdateCheck(check)
		assert.NoError(t, err)
		retrieved, err := store.GetCheckByName("check1")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)

		assert.Equal(t, check.Name, retrieved.Name)
		assert.Equal(t, check.Interval, retrieved.Interval)
		assert.Equal(t, check.Subscriptions, retrieved.Subscriptions)
		assert.Equal(t, check.Command, retrieved.Command)

		checks, err := store.GetChecks()
		assert.NoError(t, err)
		assert.NotEmpty(t, checks)
		assert.Equal(t, 1, len(checks))
	})
}

func TestEventStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		sysinfo, _ := system.Info()

		event := &types.Event{
			Entity: &types.Entity{
				ID:     "entity1",
				Class:  "system",
				System: sysinfo,
			},
			Check: &types.Check{
				Name:          "check1",
				Interval:      60,
				Subscriptions: []string{"subscription1"},
				Command:       "command1",
			},
		}

		assert.NoError(t, store.UpdateEvent(event))

		newEv, err := store.GetEventByEntityCheck(event.Entity.ID, event.Check.Name)
		assert.NoError(t, err)
		assert.EqualValues(t, event, newEv)

		events, err := store.GetEvents()
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

		assert.NoError(t, store.DeleteEventByEntityCheck(event.Entity.ID, event.Check.Name))
		newEv, err = store.GetEventByEntityCheck(event.Entity.ID, event.Check.Name)
		assert.Nil(t, newEv)
		assert.NoError(t, err)

		assert.Error(t, store.DeleteEventByEntityCheck("", ""))
		assert.Error(t, store.DeleteEventByEntityCheck("", "foo"))
		assert.Error(t, store.DeleteEventByEntityCheck("foo", ""))
	})
}
