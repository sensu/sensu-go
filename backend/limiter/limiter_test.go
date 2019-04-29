package limiter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimiter(t *testing.T) {
	limiter := NewEntityLimiter()
	assert.NotEmpty(t, limiter)
	assert.Equal(t, entityLimit, limiter.Limit())
	assert.Equal(t, 0, len(limiter.CountHistory()))
	for i := 0; i < 12; i++ {
		limiter.AddCount(i)
	}
	assert.Equal(t, 12, len(limiter.CountHistory()))
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, limiter.CountHistory())
	limiter.AddCount(99)
	assert.Equal(t, 12, len(limiter.CountHistory()))
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 99}, limiter.CountHistory())
}
