package messaging

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWizardBus(t *testing.T) {
	b := &WizardBus{}
	b.Start()

	// Should be able to publish with no subscribers.
	err := b.Publish("topic", "message1")
	assert.NoError(t, err)

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
	close(c4)

	b.Stop()

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
		bus := &WizardBus{}
		bus.Start()
		b.Run(fmt.Sprintf("%d subscribers", tc), func(b *testing.B) {
			b.SetParallelism(tc)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					ch := make(chan interface{})
					subID := string(rand.Int())
					bus.Subscribe("topic", subID, ch)
				}
			})
		})
		bus.Stop()
	}
}
