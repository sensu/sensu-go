package wait

import (
	"errors"
	"time"
)

// Backoff contains the configuration for exponential backoff
type Backoff struct {
	// InitialDelayInterval represents the initial amount of time of sleep
	InitialDelayInterval time.Duration
	// MaxDelayInterval represents the maximal amount of time of sleep between
	// retries
	MaxDelayInterval time.Duration
	// MaxRetryAttempts is the maximal number of retries before exiting with
	// an error. A value of zero signifies unlimited retry attemps
	MaxRetryAttempts int
	// Multiplier is used to increment the current interval by multiplying it with
	// this multiplier
	Multiplier float64
}

// BackoffFunc represents a function to retry with exponential backoff, which
// returns true if the attempt is successful, or an error if the retry should be
// aborted
type BackoffFunc func() (done bool, err error)

// ErrMaxRetryAttempts is returned when the number of maximal retry attempts is
// reached
var ErrMaxRetryAttempts = errors.New("maximal number of retry attempts reached")

// ExponentialBackoff retries the provided func with exponential backoff, until
// the maximal number of retries is reached
func (b *Backoff) ExponentialBackoff(fn BackoffFunc) error {
	sleep := b.InitialDelayInterval

	for i := 0; i < b.MaxRetryAttempts || b.MaxRetryAttempts == 0; i++ {
		if i != 0 {
			// Sleep for the determined duration
			time.Sleep(sleep)

			// Exponentially increase that sleep duration
			sleep = time.Duration(float64(sleep) * b.Multiplier)
		}

		if ok, err := fn(); err != nil || ok {
			return err
		}
	}

	return ErrMaxRetryAttempts
}
