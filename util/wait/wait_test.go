package wait

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var errBackoff = errors.New("error")

func mockBackoffFunc(max int) func() (bool, error) {
	i := 0
	return func() (bool, error) {
		i++
		if i == max {
			return true, nil
		}

		return false, nil
	}
}

func mockBackoffFuncErr() func() (bool, error) {
	return func() (bool, error) {
		return false, errBackoff
	}
}

func TestExponentialBackoff(t *testing.T) {
	// It should reach MaxRetryAttempts
	fn := mockBackoffFunc(3)
	b := Backoff{
		InitialDelayInterval: 1 * time.Millisecond,
		MaxDelayInterval:     1 * time.Second,
		MaxRetryAttempts:     2,
		Multiplier:           1.5,
	}
	assert.Error(t, b.ExponentialBackoff(fn))

	// It should be successful
	b.MaxRetryAttempts = 3
	assert.NoError(t, b.ExponentialBackoff(fn))

	// It should return an error from our func
	errFn := mockBackoffFuncErr()
	assert.Equal(t, errBackoff, b.ExponentialBackoff(errFn))
}
