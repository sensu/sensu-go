package schedulerd

import (
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestToggleCron(t *testing.T) {
	sched := &CheckScheduler{}

	// In this case, cron variables are not set and there is no side-effect
	check := types.FixtureCheckConfig("foobar")
	check.Cron = ""
	check.Interval = 5
	assert.False(t, sched.toggleCron(check))
	assert.Equal(t, uint32(0), sched.checkInterval)
	assert.Equal(t, sched.checkCron, "")
	assert.Equal(t, sched.lastCronState, "")

	// In this case, cron variables are set
	check.Cron = "* * * * *"
	check.Interval = 5
	assert.True(t, sched.toggleCron(check))
	assert.Equal(t, uint32(5), sched.checkInterval)
	assert.Equal(t, "* * * * *", sched.checkCron)
	assert.Equal(t, "* * * * *", sched.lastCronState)

	// In this case, cron variables are not set and there is no side-effect
	check.Cron = "1 * * * *"
	assert.False(t, sched.toggleCron(check))
	assert.Equal(t, "* * * * *", sched.checkCron)
	assert.Equal(t, "* * * * *", sched.lastCronState)
}
