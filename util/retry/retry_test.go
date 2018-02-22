package retry

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var errBackoff = errors.New("error")

func mockBackoffFunc(max int) func(retry int) (bool, error) {
	i := 0
	return func(retry int) (bool, error) {
		i++
		if i == max {
			return true, nil
		}

		return false, nil
	}
}

func mockBackoffFuncErr() func(retry int) (bool, error) {
	return func(retry int) (bool, error) {
		return false, errBackoff
	}
}

func TestExponentialBackoff(t *testing.T) {
	// It should reach MaxRetryAttempts
	fn := mockBackoffFunc(3)
	b := ExponentialBackoff{
		InitialDelayInterval: 1 * time.Millisecond,
		MaxDelayInterval:     1 * time.Second,
		MaxRetryAttempts:     2,
		Multiplier:           1.5,
	}
	assert.Error(t, b.Retry(fn))

	// It should be successful
	b.MaxRetryAttempts = 3
	assert.NoError(t, b.Retry(fn))

	// It should return an error from our func
	errFn := mockBackoffFuncErr()
	assert.Equal(t, errBackoff, b.Retry(errFn))
}
