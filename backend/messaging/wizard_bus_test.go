package messaging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryBus(t *testing.T) {
	b := &WizardBus{}
	b.Start()

	err := b.Publish("topic", []byte("message1"))
	assert.NoError(t, err)
	// should be able to publish with no subscribers

	sub1, ch1, err := b.NewSubscriber()
	assert.NoError(t, err)
	sub2, ch2, err := b.NewSubscriber()
	assert.NoError(t, err)
	assert.NoError(t, b.Subscribe("topic", sub1))

	assert.NoError(t, b.Subscribe("topic", sub2))

	err = b.Publish("topic", []byte("message2"))
	assert.NoError(t, err)
	err = b.Publish("topic", []byte("message3"))
	assert.NoError(t, err)

	b.Stop()

	received := 0
	messages := []string{}
	for m := range ch1 {
		received++
		messages = append(messages, string(m))
	}

	for m := range ch2 {
		received++
		messages = append(messages, string(m))
	}
	assert.Equal(t, "message2", messages[0])
	assert.Equal(t, "message3", messages[3])
	assert.Equal(t, 4, len(messages))
	assert.Equal(t, 4, received)
}
