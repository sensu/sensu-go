package backend

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/testing/fixtures"
	"github.com/sensu/sensu-go/testing/util"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestCheckScheduler(t *testing.T) {
	bus := &messaging.MemoryBus{}
	assert.NoError(t, bus.Start())

	st := fixtures.NewFixtureStore()
	check, _ := st.GetCheckByName("check1")
	check.Interval = 1

	scheduler := &CheckScheduler{
		MessageBus: bus,
		Store:      fixtures.NewFixtureStore(),
		Check:      check,
	}

	c1, err := bus.Subscribe("subscription1", "")
	assert.NoError(t, err)

	assert.NoError(t, scheduler.Start())
	time.Sleep(1 * time.Second)
	scheduler.Stop()
	assert.NoError(t, bus.Stop())

	messages := [][]byte{}
	for msg := range c1 {
		messages = append(messages, msg)
	}
	assert.Equal(t, 1, len(messages))
	evt := &types.Event{}
	assert.NoError(t, json.Unmarshal(messages[0], evt))
	assert.Equal(t, "check1", evt.Check.Name)
}

func TestCheckerd(t *testing.T) {
	util.WithTempDir(func(tmpDir string) {
		p := make([]int, 2)
		err := util.RandomPorts(p)
		if err != nil {
			assert.FailNow(t, "failed to get ports for testing: ", err.Error())
		}

		cfg := etcd.NewConfig()
		cfg.StateDir = tmpDir

		cfg.ClientListenURL = fmt.Sprintf("http://127.0.0.1:%d", p[0])
		cfg.PeerListenURL = fmt.Sprintf("http://127.0.0.1:%d", p[1])
		cfg.InitialCluster = fmt.Sprintf("default=http://127.0.0.1:%d", p[1])
		e, err := etcd.NewEtcd(cfg)
		assert.NoError(t, err)
		defer e.Shutdown()

		cli, err := e.NewClient()
		assert.NoError(t, err)
		st, err := e.NewStore()
		bus := &messaging.MemoryBus{}
		assert.NoError(t, bus.Start())

		checker := &Checker{
			Client:     cli,
			Store:      st,
			MessageBus: bus,
		}
		checker.Start()

		ch, err := bus.Subscribe("subscription", "")
		assert.NoError(t, err)
		assert.NotNil(t, ch)

		check := &types.Check{
			Name:          "check_name",
			Interval:      1,
			Command:       "command",
			Subscriptions: []string{"subscription"},
		}
		assert.NoError(t, check.Validate())
		assert.NoError(t, st.UpdateCheck(check))

		time.Sleep(1 * time.Second)
		assert.NoError(t, st.DeleteCheckByName(check.Name))
		time.Sleep(1 * time.Second)
		assert.NoError(t, checker.Stop())
		assert.NoError(t, bus.Stop())

		for msg := range ch {
			evt := &types.Event{}
			assert.NoError(t, json.Unmarshal(msg, evt))
			assert.EqualValues(t, check, evt.Check)
		}
	})
}
