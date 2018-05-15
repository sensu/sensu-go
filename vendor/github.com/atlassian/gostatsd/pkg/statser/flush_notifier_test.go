package statser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFlushNotifierCycle(t *testing.T) {
	t.Parallel()
	fn := &flushNotifier{}

	_, unregister := fn.RegisterFlush()
	assert.Equal(t, len(fn.flushTargets), 1)
	unregister()
	assert.Equal(t, len(fn.flushTargets), 0)
}

func TestFlushNotifierClosesChannel(t *testing.T) {
	t.Parallel()
	fn := &flushNotifier{}

	ch, unregister := fn.RegisterFlush()
	assert.Equal(t, len(fn.flushTargets), 1)
	unregister()
	assert.Equal(t, len(fn.flushTargets), 0)

	select {
	case _, ok := <-ch:
		assert.False(t, ok)
	default:
		assert.Fail(t, "channel not closed")
	}
}

func TestFlushNotifierMultiCycle(t *testing.T) {
	t.Parallel()
	fn := &flushNotifier{}

	ch1, unregister1 := fn.RegisterFlush()
	ch2, unregister2 := fn.RegisterFlush()
	_, unregister3 := fn.RegisterFlush()
	assert.Equal(t, len(fn.flushTargets), 3)

	unregister2()
	assert.Equal(t, len(fn.flushTargets), 2)
	for ch := range fn.flushTargets {
		assert.NotEqual(t, ch, ch2)
	}

	unregister1()
	assert.Equal(t, len(fn.flushTargets), 1)
	for ch := range fn.flushTargets {
		assert.NotEqual(t, ch, ch1)
	}

	unregister3()
	assert.Equal(t, len(fn.flushTargets), 0)
}

func TestFlushNotifierFires(t *testing.T) {
	t.Parallel()
	fn := &flushNotifier{}

	ch, unregister := fn.RegisterFlush()

	go func() {
		// Just enough to be certain that the select below is ready.
		time.Sleep(10 * time.Millisecond)
		fn.NotifyFlush(0)
	}()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	select {
	case <-ch:
		// happy
	case <-ticker.C:
		assert.Fail(t, "nothing received")
	}

	unregister()
	assert.Equal(t, len(fn.flushTargets), 0)
}

func TestFlushNotifierDoesNotBlock(t *testing.T) {
	t.Parallel()
	fn := &flushNotifier{}

	_, unregister := fn.RegisterFlush()

	deadline := time.Now().Add(10 * time.Millisecond)
	fn.NotifyFlush(0)
	assert.Truef(t, time.Now().Before(deadline), "NotifyFlush ran too long")

	unregister()
}
