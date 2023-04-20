package schedulerd

import (
	"testing"

	v2 "github.com/sensu/core/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestToggleIntervalSchedule(t *testing.T) {
	check := v2.FixtureCheckConfig("foobar")
	sched := &IntervalScheduler{
		check:			check,
		lastIntervalState:	check.Interval,
		logger:			logger.WithFields(logrus.Fields{}),
	}

	// no state change
	assert.False(t, sched.toggleSchedule())
	assert.Equal(t, uint32(60), sched.check.Interval)
	assert.Equal(t, uint32(60), sched.lastIntervalState)

	// interval -> interval change
	sched.check.Interval = 30
	assert.True(t, sched.toggleSchedule())
	assert.Equal(t, uint32(30), sched.check.Interval)
	assert.Equal(t, uint32(30), sched.lastIntervalState)

	// no state change
	assert.False(t, sched.toggleSchedule())
}

func TestToggleCronSchedule(t *testing.T) {
	check := v2.FixtureCheckConfig("foobar")
	sched := &CronScheduler{
		check:		check,
		lastCronState:	check.Cron,
		logger:		logger.WithFields(logrus.Fields{}),
	}

	// no state change
	assert.False(t, sched.toggleSchedule())

	// cron -> cron change
	sched.check.Interval = 0
	sched.check.Cron = "*/2 * * * *"
	assert.True(t, sched.toggleSchedule())
	assert.Equal(t, "*/2 * * * *", sched.check.Cron)
	assert.Equal(t, "*/2 * * * *", sched.lastCronState)

	// no state change
	assert.False(t, sched.toggleSchedule())
}
