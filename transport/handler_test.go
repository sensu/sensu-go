package transport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuncAddHandler(t *testing.T) {
	handler := NewMessageHandler()
	handler.AddHandler("MessageType", func(payload []byte) error {
		assert.EqualValues(t, []byte{0}, payload)
		return nil
	})

	msg := &Message{"MessageType", []byte{0}}
	assert.NoError(t, handler.Handle(msg))
}

func BenchmarkHandleMessage(b *testing.B) {
	handler := NewMessageHandler()
	handler.AddHandler("MessageType", func(payload []byte) error {
		return nil
	})

	msg := &Message{"MessageType", []byte{0}}
	for n := 0; n < b.N; n++ {
		handler.Handle(msg)
	}
}
