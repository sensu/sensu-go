package schedulerd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNextCronTime(t *testing.T) {
	// Valid cron string will return a time in the future
	nextCron, err := NextCronTime("* * * * *")
	assert.Nil(t, err)
	assert.True(t, nextCron > 0)

	// Invalid cron string will return an error
	nextCron, err = NextCronTime("invalid")
	assert.NotNil(t, err)
	assert.True(t, nextCron == 0)
}
