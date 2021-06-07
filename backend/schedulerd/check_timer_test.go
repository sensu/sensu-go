package schedulerd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNextCronTime(t *testing.T) {
	now := mockTime.Now()

	// Valid cron string will return a time in the future, on an even minute
	nextCron, err := NextCronTime(now, "* * * * *")
	assert.Nil(t, err)
	assert.True(t, nextCron >= 0)
	assert.True(t, now.Add(nextCron).Second() == 0)

	// Valid cron string will return a time in the future, on an even hour
	nextCron, err = NextCronTime(now, "0 * * * *")
	assert.Nil(t, err)
	assert.True(t, nextCron >= 0)
	assert.True(t, now.Add(nextCron).Minute() == 0)

	// Valid cron string with timezone will return a time in the future, on an even hour
	nextCron, err = NextCronTime(now, "CRON_TZ=Asia/Tokyo 0 * * * *")
	assert.Nil(t, err)
	assert.True(t, nextCron >= 0)
	assert.True(t, now.Add(nextCron).Minute() == 0)

	// Valid cron Sunday as zero
	nextCron, err = NextCronTime(now, "0 * * * 0")
	assert.Nil(t, err)
	assert.True(t, nextCron >= 0)
	assert.True(t, now.Add(nextCron).Minute() == 0)

	// Valid cron Sunday as letter
	nextCron, err = NextCronTime(now, "0 * * * SUN")
	assert.Nil(t, err)
	assert.True(t, nextCron >= 0)
	assert.True(t, now.Add(nextCron).Minute() == 0)

	// Invalid cron string with timezone in the wrong spot
	nextCron, err = NextCronTime(now, "0 * * * * CRON_TZ=Asia/Tokyo")
	assert.NotNil(t, err)
	assert.True(t, nextCron == 0)

	// Invalid cron string will return an error
	nextCron, err = NextCronTime(now, "invalid")
	assert.NotNil(t, err)
	assert.True(t, nextCron == 0)
}

func TestSplay(t *testing.T) {
	timer := NewIntervalTimer("check1", 10)

	assert.Condition(t, func() bool { return timer.splay > 0 })

	timer2 := NewIntervalTimer("check1", 10)
	assert.Equal(t, timer.splay, timer2.splay)
}

func TestInitialOffset(t *testing.T) {
	inputs := []uint{1, 10, 60}
	for _, intervalSeconds := range inputs {
		now := mockTime.Now()
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
