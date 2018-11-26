package schedulerd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/testing/mockbus"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRoundRobinScheduler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bus := mockbus.MockBus{}
	bus.On("PublishDirect", mock.Anything, "ramen-vending-machine", mock.Anything).Return(nil)
	sched := newRoundRobinScheduler(ctx, &bus)
	msg := &roundRobinMessage{
		req:          types.FixtureCheckRequest("foo"),
		subscription: "ramen-vending-machine",
	}
	wg, err := sched.Schedule(msg)
	require.NoError(t, err)
	wg.Wait()
	bus.AssertCalled(t, "PublishDirect", mock.Anything, "ramen-vending-machine", msg.req)
}
