package schedulerd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSplay(t *testing.T) {
	timer := NewIntervalTimer("check1", 10)

	assert.Condition(t, func() bool { return timer.splay > 0 })

	timer2 := NewIntervalTimer("check1", 10)
	assert.Equal(t, timer.splay, timer2.splay)
}

func TestInitialOffset(t *testing.T) {
	inputs := []uint{1, 10, 60}
	for _, intervalSeconds := range inputs {
		now := time.Now()
		timer := NewIntervalTimer("check1", intervalSeconds)
		nextExecution := timer.calcInitialOffset()
		executionTime := now.Add(nextExecution)

		// We've scheduled it in the future.
		assert.Condition(t, func() bool { return executionTime.Sub(now) > 0 })
		// The offset is less than the check interval.
		assert.Condition(t, func() bool { return nextExecution < (time.Duration(intervalSeconds) * time.Second) })
		// The next execution occurs _before_ now + interval.
		assert.Condition(t, func() bool { return executionTime.Before(now.Add(time.Duration(intervalSeconds) * time.Second)) })
	}
}
