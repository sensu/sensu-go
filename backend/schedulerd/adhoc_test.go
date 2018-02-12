package schedulerd

import (
	"context"
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
	newAdhocExec := NewAdhocRequestExecutor(store, bus)
	if err = newAdhocExec.Start(context.Background()); err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.NoError(t, newAdhocExec.bus.Start())

	goodCheck := types.FixtureCheckConfig("goodCheck")
	ch := make(chan interface{}, 10)
	assert.NoError(t, bus.Subscribe("subscription", "channel", ch))

	if err = newAdhocExec.adhocQueue.Enqueue(context.Background(), goodCheck.String()); err != nil {
		assert.FailNow(t, err.Error())
	}

	for msg := range ch {
		result, ok := msg.(*types.CheckConfig)
		assert.True(t, ok)
		assert.EqualValues(t, goodCheck, result)
	}
}
