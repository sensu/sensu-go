package schedulerd

import (
	"fmt"
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
		cfg.DataDir = tmpDir

		cfg.ListenClientURL = fmt.Sprintf("http://127.0.0.1:%d", p[0])
		cfg.ListenPeerURL = fmt.Sprintf("http://127.0.0.1:%d", p[1])
		cfg.InitialCluster = fmt.Sprintf("default=http://127.0.0.1:%d", p[1])
		e, err := etcd.NewEtcd(cfg)
		assert.NoError(t, err)
		defer e.Shutdown()

		st, err := e.NewStore()
		bus := &messaging.WizardBus{}
		assert.NoError(t, bus.Start())

		checker := &Schedulerd{
			Store:      st,
			MessageBus: bus,
		}
		checker.Start()

		ch := make(chan interface{}, 10)
		assert.NoError(t, bus.Subscribe("subscription", "channel", ch))

		check := types.FixtureCheckConfig("check_name")
		assert.NoError(t, check.Validate())
		assert.NoError(t, st.UpdateCheckConfig(check))

		time.Sleep(1 * time.Second)
		assert.NoError(t, st.DeleteCheckConfigByName(check.Organization, check.Name))
		time.Sleep(1 * time.Second)
		assert.NoError(t, checker.Stop())
		assert.NoError(t, bus.Stop())
		close(ch)

		for msg := range ch {
			result, ok := msg.(*types.CheckConfig)
			assert.True(t, ok)
			assert.EqualValues(t, check, result)
		}
	})
}
