package messaging

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/sensu/sensu-go/testing/mockring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type channelSubscriber struct {
	Channel chan interface{}
}

func (c channelSubscriber) Receiver() chan<- interface{} {
	return c.Channel
}

func TestWizardBus(t *testing.T) {
	b, err := NewWizardBus(WizardBusConfig{
		RingGetter: &mockring.Getter{},
	})
	require.NoError(t, err)
	require.NoError(t, b.Start())

	// Should be able to publish with no subscribers.
	require.NoError(t, b.Publish("topic", "message1"))

	sub1 := channelSubscriber{make(chan interface{}, 100)}
	sub2 := channelSubscriber{make(chan interface{}, 100)}
	sub3 := channelSubscriber{make(chan interface{}, 100)}

	// Will unsubscribe from the topic after receiving messages.
	sub4 := channelSubscriber{make(chan interface{}, 100)}

	subscr1, err := b.Subscribe("topic", "1", sub1)
	assert.NoError(t, err)

	subscr2, err := b.Subscribe("topic", "2", sub2)
	assert.NoError(t, err)

	subscr3, err := b.Subscribe("topic", "3", sub3)
	assert.NoError(t, err)

	subscr4, err := b.Subscribe("topic", "4", sub4)
	assert.NoError(t, err)

	err = b.Publish("topic", "message2")
	assert.NoError(t, err)
	err = b.Publish("topic", "message3")
	assert.NoError(t, err)
	err = b.Publish("topic", "message4")
	assert.NoError(t, err)

	subscr4.Cancel()
	close(sub4.Channel)

	subscr1.Cancel()
	close(sub1.Channel)

	subscr2.Cancel()
	close(sub2.Channel)

	subscr3.Cancel()
	close(sub3.Channel)

	require.NoError(t, b.Stop())

	received := 0
	messages := []string{}
	for m := range sub1.Channel {
		received++
		messages = append(messages, m.(string))
	}

	for m := range sub2.Channel {
		received++
		messages = append(messages, m.(string))
	}

	for m := range sub3.Channel {
		received++
		messages = append(messages, m.(string))
	}
	assert.Equal(t, 9, len(messages))
	assert.Equal(t, 9, received)
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
				i := 0
				for pb.Next() {
					sub := channelSubscriber{make(chan interface{})}
					_, _ = bus.Subscribe("topic", fmt.Sprintf("client-%d", i), sub)
					i++
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

	subscribers := map[string]channelSubscriber{
		"a": channelSubscriber{make(chan interface{}, 1)},
		"b": channelSubscriber{make(chan interface{}, 1)},
		"c": channelSubscriber{make(chan interface{}, 1)},
	}
	for id, ch := range subscribers {
		_, err := bus.Subscribe("topic", id, ch)
		require.NoError(t, err)
	}
	require.NoError(t, bus.PublishDirect(context.TODO(), "topic", "hello, world"))
	select {
	case msg := <-subscribers["a"].Channel:
		assert.Equal(t, msg, "hello, world")
	case <-subscribers["b"].Channel:
		t.Error("got message on b")
	case <-subscribers["c"].Channel:
		t.Error("got message on c")
	default:
		t.Error("no message sent")
	}
	select {
	case <-subscribers["a"].Channel:
		t.Error("got message on a")
	case <-subscribers["b"].Channel:
		t.Error("got message on b")
	case <-subscribers["c"].Channel:
		t.Error("got message on c")
	default:
	}
}

func TestBug1407(t *testing.T) {
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

	subscriber := channelSubscriber{make(chan interface{}, 1)}
	subscription, err := bus.Subscribe("topic", "a", subscriber)
	require.NoError(t, err)
	require.NoError(t, subscription.Cancel())
	subscription, err = bus.Subscribe("topic", "a", subscriber)
	require.NoError(t, err)
	closed := bus.topics["topic"].IsClosed()
	assert.False(t, closed)

}
