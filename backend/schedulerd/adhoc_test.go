// +build integration,race

package schedulerd

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store/etcd/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestAdhocExecutor(t *testing.T) {
	store, err := testutil.NewStoreInstance()

	if err != nil {
		assert.FailNow(t, err.Error())
	}
	bus := &messaging.WizardBus{}
	newAdhocExec := NewAdhocRequestExecutor(context.Background(), store, bus)
	defer newAdhocExec.Stop()
	assert.NoError(t, newAdhocExec.bus.Start())

	goodCheck := types.FixtureCheckConfig("goodCheck")
	goodCheck.Subscriptions = []string{"subscription1"}

	goodCheckRequest := &types.CheckRequest{}
	goodCheckRequest.Config = goodCheck
	ch := make(chan interface{}, 1)
	topic := messaging.SubscriptionTopic(goodCheck.Organization, goodCheck.Environment, "subscription1")
	assert.NoError(t, bus.Subscribe(topic, "channel", ch))

	marshaledCheck, err := json.Marshal(goodCheck)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	if err = newAdhocExec.adhocQueue.Enqueue(context.Background(), string(marshaledCheck)); err != nil {
		assert.FailNow(t, err.Error())
	}

	msg := <-ch
	result, ok := msg.(*types.CheckRequest)
	assert.True(t, ok)
	assert.EqualValues(t, goodCheckRequest, result)
}
