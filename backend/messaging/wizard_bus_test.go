package messaging

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type channelSubscriber struct {
	Channel chan interface{}
}

func (c channelSubscriber) Receiver() chan<- interface{} {
	return c.Channel
}

func TestWizardBus(t *testing.T) {
	b, err := NewWizardBus(WizardBusConfig{})
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

	_ = subscr4.Cancel()
	close(sub4.Channel)

	_ = subscr1.Cancel()
	close(sub1.Channel)

	_ = subscr2.Cancel()
	close(sub2.Channel)

	_ = subscr3.Cancel()
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
		bus, _ := NewWizardBus(WizardBusConfig{})
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

func TestBug1407(t *testing.T) {
	bus, err := NewWizardBus(WizardBusConfig{})
	require.NoError(t, err)

	require.NoError(t, bus.Start())
	defer func() {
		_ = bus.Stop()
	}()

	subscriber := channelSubscriber{make(chan interface{}, 1)}
	subscription, err := bus.Subscribe("topic", "a", subscriber)
	require.NoError(t, err)
	require.NoError(t, subscription.Cancel())
	_, err = bus.Subscribe("topic", "a", subscriber)
	require.NoError(t, err)
	value, _ := bus.topics.Load("topic")
	topic := value.(*wizardTopic)
	assert.False(t, topic.IsClosed())

}
