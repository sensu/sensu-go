package schedulerd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCheckSplayCalculation(t *testing.T) {
	// ensure that the splay is constant for a given check
	splay := calcExecutionSplay("check1")
	assert.Condition(t, func() bool { return splay > 0 })

	splay2 := calcExecutionSplay("check1")
	assert.Equal(t, splay, splay2)
}

func TestCheckNextExecutionCalculation(t *testing.T) {
	inputs := []int{1, 10, 60}
	splay := calcExecutionSplay("check1")
	for _, intervalSeconds := range inputs {
		now := time.Now()
		nextExecution := calcNextExecution(splay, intervalSeconds)
		executionTime := now.Add(nextExecution)
		// We've scheduled it in the future.
		assert.Condition(t, func() bool { return executionTime.Sub(now) > 0 })
		// The offset is less than the check interval.
		assert.Condition(t, func() bool { return nextExecution < (time.Duration(intervalSeconds) * time.Second) })
		// The next execution occurs _before_ now + interval.
		assert.Condition(t, func() bool { return executionTime.Before(now.Add(time.Duration(intervalSeconds) * time.Second)) })
	}
}
