package schedulerd

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestSchedulerd(t *testing.T) {
	tmpDir, remove := testutil.TempDir(t)
	defer remove()

	p := make([]int, 2)
	err := testutil.RandomPorts(p)
	if err != nil {
		assert.FailNow(t, "failed to get ports for testing: ", err.Error())
	}

	cfg := etcd.NewConfig()
	cfg.DataDir = tmpDir

	peerURL := fmt.Sprintf("http://127.0.0.1:%d", p[1])

	cfg.ListenClientURL = fmt.Sprintf("http://127.0.0.1:%d", p[0])
	cfg.ListenPeerURL = peerURL
	cfg.InitialCluster = fmt.Sprintf("default=http://127.0.0.1:%d", p[1])
	cfg.InitialAdvertisePeerURL = peerURL
	e, err := etcd.NewEtcd(cfg)
	assert.NoError(t, err)
	defer e.Shutdown()

	st, err := e.NewStore()
	bus := &messaging.WizardBus{}
	assert.NoError(t, bus.Start())

	// Mock a default organization
	st.UpdateOrganization(
		context.Background(),
		&types.Organization{
			Name: "default",
		})

	// Mock a default environment
	st.UpdateEnvironment(
		context.Background(),
		"default",
		&types.Environment{
			Name: "default",
		})

	checker := &Schedulerd{
		Store:      st,
		MessageBus: bus,
	}
	checker.Start()

	ch := make(chan interface{}, 10)
	assert.NoError(t, bus.Subscribe("subscription", "channel", ch))

	check := types.FixtureCheckConfig("check_name")
	ctx := context.WithValue(context.Background(), types.OrganizationKey, check.Organization)
	ctx = context.WithValue(ctx, types.EnvironmentKey, check.Environment)

	assert.NoError(t, check.Validate())
	assert.NoError(t, st.UpdateCheckConfig(ctx, check))

	time.Sleep(1 * time.Second)

	err = st.DeleteCheckConfigByName(ctx, check.Name)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	assert.NoError(t, checker.Stop())
	assert.NoError(t, bus.Stop())
	close(ch)

	for msg := range ch {
		result, ok := msg.(*types.CheckConfig)
		assert.True(t, ok)
		assert.EqualValues(t, check, result)
	}
}
