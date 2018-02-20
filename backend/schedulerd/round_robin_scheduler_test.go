package schedulerd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/testing/mockbus"
	"github.com/sensu/sensu-go/testing/mockring"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRoundRobinScheduler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bus := mockbus.MockBus{}
	bus.On("Publish", "sensu:check:default:default:entity:agent1", mock.Anything).Return(nil)
	ring := mockring.Ring{}
	getter := mockring.Getter{
		"subscription/ramen-vending-machine": &ring,
	}
	ring.On("Next", mock.Anything).Return("agent1", nil)
	sched := newRoundRobinScheduler(ctx, &bus, getter)
	msg := &roundRobinMessage{
		req:          types.FixtureCheckRequest("foo"),
		subscription: "ramen-vending-machine",
	}
	wg, err := sched.Schedule(msg)
	require.NoError(t, err)
	wg.Wait()
	ring.AssertCalled(t, "Next", mock.Anything)
	bus.AssertCalled(t, "Publish", "sensu:check:default:default:entity:agent1", msg.req)
}
