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
	bus.On("Publish", "topic:agent1", mock.Anything).Return(nil)
	ring := mockring.Ring{}
	getter := mockring.Getter{
		"topic": &ring,
	}
	ring.On("Next", mock.Anything).Return("agent1", nil)
	sched := newRoundRobinScheduler(ctx, &bus, getter)
	msg := &roundRobinMessage{
		req:   types.FixtureCheckRequest("foo"),
		topic: "topic",
	}
	wg, err := sched.Schedule(msg)
	require.NoError(t, err)
	wg.Wait()
	ring.AssertCalled(t, "Next", mock.Anything)
	bus.AssertCalled(t, "Publish", "topic:agent1", msg.req)
}
