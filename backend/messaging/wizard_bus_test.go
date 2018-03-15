package messaging

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/sensu/sensu-go/testing/mockring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWizardBus(t *testing.T) {
	b, err := NewWizardBus(WizardBusConfig{
		RingGetter: &mockring.Getter{},
	})
	require.NoError(t, err)
	require.NoError(t, b.Start())

	// Should be able to publish with no subscribers.
	require.NoError(t, b.Publish("topic", "message1"))

	c1 := make(chan interface{}, 100)
	c2 := make(chan interface{}, 100)

	// Topic publisher should not be blocked by a channel.
	c3 := make(chan interface{}, 1)

	// Will unsubscribe from the topic after receiving messages.
	c4 := make(chan interface{}, 100)

	assert.NoError(t, b.Subscribe("topic", "consumer1", c1))
	assert.NoError(t, b.Subscribe("topic", "consumer2", c2))
	assert.NoError(t, b.Subscribe("topic", "consumer3", c3))
	assert.NoError(t, b.Subscribe("topic", "consumer4", c4))

	err = b.Publish("topic", "message2")
	assert.NoError(t, err)
	err = b.Publish("topic", "message3")
	assert.NoError(t, err)
	err = b.Publish("topic", "message4")
	assert.NoError(t, err)

	assert.NoError(t, b.Unsubscribe("topic", "consumer4"))

	require.NoError(t, b.Stop())
	close(c1)
	close(c2)
	close(c3)

	received := 0
	messages := []string{}
	for m := range c1 {
		received++
		messages = append(messages, m.(string))
	}

	for m := range c2 {
		received++
		messages = append(messages, m.(string))
	}

	for m := range c3 {
		received++
		messages = append(messages, m.(string))
	}
	assert.Equal(t, 7, len(messages))
	assert.Equal(t, 7, received)
	assert.Equal(t, "message2", messages[0])
	assert.Equal(t, "message3", messages[1])
	assert.Equal(t, "message4", messages[2])
	assert.Equal(t, "message2", messages[3])
	assert.Equal(t, "message2", messages[6])
}

func BenchmarkSubscribe(b *testing.B) {
	testCases := []int{100, 1000, 10000}

	rand.Seed(time.Now().UnixNano())

	for _, tc := range testCases {
		bus, _ := NewWizardBus(WizardBusConfig{
			RingGetter: &mockring.Getter{},
		})
		_ = bus.Start()
		b.Run(fmt.Sprintf("%d subscribers", tc), func(b *testing.B) {
			b.SetParallelism(tc)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					ch := make(chan interface{})
					subID := string(rand.Int())
					_ = bus.Subscribe("topic", subID, ch)
				}
			})
		})
		_ = bus.Stop()
	}
}

func TestPublishDirect(t *testing.T) {
	ring := &mockring.Ring{}
	ring.On("Add", mock.Anything, mock.Anything).Return(nil)
	ring.On("Next", mock.Anything).Return("a", nil)
	getter := &mockring.Getter{"topic": ring}
	bus, err := NewWizardBus(WizardBusConfig{
		RingGetter: getter,
	})
	require.NoError(t, err)

	require.NoError(t, bus.Start())
	defer bus.Stop()

	subscribers := map[string]chan interface{}{
		"a": make(chan interface{}, 1),
		"b": make(chan interface{}, 1),
		"c": make(chan interface{}, 1),
	}
	for sub, ch := range subscribers {
		require.NoError(t, bus.Subscribe("topic", sub, ch))
	}
	require.NoError(t, bus.PublishDirect("topic", "hello, world"))
	select {
	case msg := <-subscribers["a"]:
		assert.Equal(t, msg, "hello, world")
	case <-subscribers["b"]:
		t.Error("got message on b")
	case <-subscribers["c"]:
		t.Error("got message on c")
	default:
		t.Error("no message sent")
	}
	select {
	case <-subscribers["a"]:
		t.Error("got message on a")
	case <-subscribers["b"]:
		t.Error("got message on b")
	case <-subscribers["c"]:
		t.Error("got message on c")
	default:
	}
}
