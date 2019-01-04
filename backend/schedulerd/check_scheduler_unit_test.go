package schedulerd

import (
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestToggleSchedule(t *testing.T) {
	check := types.FixtureCheckConfig("foobar")
	sched := &CheckScheduler{
		check:             check,
		lastCronState:     check.Cron,
		lastIntervalState: check.Interval,
		logger:            logger.WithFields(logrus.Fields{}),
	}

	// no state change
	assert.False(t, sched.toggleSchedule())
	assert.Equal(t, uint32(60), sched.check.Interval)
	assert.Equal(t, uint32(60), sched.lastIntervalState)
	assert.Equal(t, "", sched.check.Cron)
	assert.Equal(t, "", sched.lastCronState)

	// interval -> interval change
	sched.check.Interval = 30
	assert.True(t, sched.toggleSchedule())
	assert.Equal(t, uint32(30), sched.check.Interval)
	assert.Equal(t, uint32(30), sched.lastIntervalState)
	assert.Equal(t, "", sched.check.Cron)
	assert.Equal(t, "", sched.lastCronState)

	// interval -> cron change
	sched.check.Interval = 0
	sched.check.Cron = "* * * * *"
	assert.True(t, sched.toggleSchedule())
	assert.Equal(t, uint32(0), sched.check.Interval)
	assert.Equal(t, uint32(0), sched.lastIntervalState)
	assert.Equal(t, "* * * * *", sched.check.Cron)
	assert.Equal(t, "* * * * *", sched.lastCronState)

	// cron -> cron change
	sched.check.Interval = 0
	sched.check.Cron = "*/2 * * * *"
	assert.True(t, sched.toggleSchedule())
	assert.Equal(t, uint32(0), sched.check.Interval)
	assert.Equal(t, uint32(0), sched.lastIntervalState)
	assert.Equal(t, "*/2 * * * *", sched.check.Cron)
	assert.Equal(t, "*/2 * * * *", sched.lastCronState)

	// cron -> interval change
	sched.check.Interval = 60
	sched.check.Cron = ""
	assert.True(t, sched.toggleSchedule())
	assert.Equal(t, uint32(60), sched.check.Interval)
	assert.Equal(t, uint32(60), sched.lastIntervalState)
	assert.Equal(t, "", sched.check.Cron)
	assert.Equal(t, "", sched.lastCronState)

	// no state change
	assert.False(t, sched.toggleSchedule())
	assert.Equal(t, uint32(60), sched.check.Interval)
	assert.Equal(t, uint32(60), sched.lastIntervalState)
	assert.Equal(t, "", sched.check.Cron)
	assert.Equal(t, "", sched.lastCronState)
}
