package schedulerd

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/testing/util"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestSchedulerd(t *testing.T) {
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
		bus := &messaging.WizardBus{}
		assert.NoError(t, bus.Start())

		checker := &Schedulerd{
			Client:     cli,
			Store:      st,
			MessageBus: bus,
			wg:         &sync.WaitGroup{},
		}
		checker.Start()

		ch := make(chan interface{}, 10)
		assert.NoError(t, bus.Subscribe("subscription", "channel", ch))

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
		close(ch)

		for msg := range ch {
			evt, ok := msg.(*types.Event)
			assert.True(t, ok)
			assert.EqualValues(t, check, evt.Check)
		}
	})
}
