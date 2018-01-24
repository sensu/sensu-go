package schedulerd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNextCronTime(t *testing.T) {
	now := time.Now()

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

	// Invalid cron string will return an error
	nextCron, err = NextCronTime(now, "invalid")
	assert.NotNil(t, err)
	assert.True(t, nextCron == 0)
}
