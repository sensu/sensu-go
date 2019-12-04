package retry

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

// ExponentialBackoff contains the configuration for exponential backoff
type ExponentialBackoff struct {
	// If Ctx is canceled, the retry method will terminate early with the
	// error from Ctx.Err().
	Ctx context.Context

	// InitialDelayInterval represents the initial amount of time of sleep
	InitialDelayInterval time.Duration

	// MaxDelayInterval represents the maximal amount of time of sleep between
	// retries
	MaxDelayInterval time.Duration

	// MaxElapsedTime represents the maximal amount of time allowed to retry. A
	// value of zero signifies no limit
	MaxElapsedTime time.Duration

	// MaxRetryAttempts is the maximal number of retries before exiting with
	// an error. A value of zero signifies unlimited retry attemps
	MaxRetryAttempts int

	// Multiplier is used to increment the current interval by multiplying it with
	// this multiplier. If not supplied, it will be set to DefaultMultiplier.
	Multiplier float64

	// start contains the starting time of the retry attempts
	start time.Time
}

const DefaultMultiplier float64 = 2.0

// Func represents a function to retry, which returns true if the attempt is
// successful, or an error if the retry should be aborted
type Func func(retry int) (done bool, err error)

// ErrMaxRetryAttempts is returned when the number of maximal retry attempts is
// reached
var ErrMaxRetryAttempts = errors.New("maximal number of retry attempts reached")

// ErrMaxElapsedTime is returned when the maximal elapsed time is reached
var ErrMaxElapsedTime = errors.New("maximal elapsed time reached")

// Retry retries the provided func with exponential backoff, until
// the maximal number of retries is reached
func (b *ExponentialBackoff) Retry(fn Func) error {
	wait := b.InitialDelayInterval
	ctx := context.Background()
	if b.Ctx != nil {
		ctx = b.Ctx
	}

	for i := 0; i < b.MaxRetryAttempts || b.MaxRetryAttempts == 0; i++ {
		if i != 0 {
			// Verify if we reached the MaxElapsedTime
			if b.MaxElapsedTime != 0 && time.Since(b.start) > b.MaxElapsedTime {
				return ErrMaxElapsedTime
			}

			// Sleep for the determined duration
			if b.MaxDelayInterval > 0 && wait > b.MaxDelayInterval {
				wait = b.MaxDelayInterval
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}

			// Exponentially increase that sleep duration
			multiplier := b.Multiplier
			if multiplier == 0 {
				multiplier = DefaultMultiplier
			}
			wait = time.Duration(float64(wait) * multiplier)

			// Add a jitter (randomized delay) for the next attempt, to prevent
			// potential collisions
			wait = wait + time.Duration(rand.Float64()*float64(wait))
		} else {
			// Save the current time, in order to measure the total execution time
			b.start = time.Now()
		}

		if ok, err := fn(i); err != nil || ok {
			return err
		}
	}

	return ErrMaxRetryAttempts
}
