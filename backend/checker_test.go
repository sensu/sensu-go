package backend

import (
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/stretchr/testify/assert"
)

func TestMessageScheduler(t *testing.T) {
	bus := &messaging.MemoryBus{}
	assert.NoError(t, bus.Start())

	ms := &MessageScheduler{
		MessageBus: bus,
		HashKey:    []byte("hash_key"),
		Interval:   1,
		MsgBytes:   []byte("message"),
		Topics:     []string{"topic"},
	}

	c1, err := bus.Subscribe("topic", "channel")
	assert.NoError(t, err)

	assert.NoError(t, ms.Start())
	time.Sleep(1 * time.Second)
	ms.Stop()
	assert.NoError(t, bus.Stop())

	messages := []string{}
	for msg := range c1 {
		messages = append(messages, string(msg))
	}
	assert.Equal(t, 1, len(messages))
	assert.Equal(t, "message", messages[0])
}

func TestCheckerd(t *testing.T) {

}
