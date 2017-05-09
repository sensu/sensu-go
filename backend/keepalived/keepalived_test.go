package keepalived

import (
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
)

func TestStartStop(t *testing.T) {
	k := &Keepalived{}
	assert.Error(t, k.Start())

	messageBus := &messaging.WizardBus{}
	k.MessageBus = messageBus
	assert.Error(t, k.Start())
	messageBus.Start()
	defer messageBus.Stop()

	store := &mockstore.MockStore{}
	k.KeepaliveStore = store
	assert.NoError(t, k.Start())

	assert.NoError(t, k.Status())

	var err error
	select {
	case err = <-k.Err():
	default:
	}
	assert.NoError(t, err)

	assert.NoError(t, k.Stop())
}
