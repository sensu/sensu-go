package retry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"
	"unsafe"
)

const DefaultMultiplier float64 = 2.0

// JSONTimeDuration is like time.Duration, but with friendly JSON methods.
type JSONTimeDuration time.Duration

// MarshalJSON always returns non-nil error.
func (j JSONTimeDuration) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	if _, err := fmt.Fprintf(buf, "%q", time.Duration(j).String()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (j *JSONTimeDuration) UnmarshalJSON(b []byte) error {
	if len(b) == 1 && b[0] == '0' {
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	t, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*j = JSONTimeDuration(t)
	return nil
}

func (j JSONTimeDuration) String() string {
	return time.Duration(j).String()
}

// ExponentialBackoff contains the configuration for exponential backoff
type ExponentialBackoff struct {
	// If Ctx is canceled, the retry method will terminate early with the
	// error from Ctx.Err().
	Ctx context.Context `json:"-"`

	// InitialDelayInterval represents the initial amount of time of sleep
	InitialDelayInterval time.Duration `json:"initial_delay_interval,omitempty"`

	// MaxDelayInterval represents the maximal amount of time of sleep between
	// retries
	MaxDelayInterval time.Duration `json:"max_delay_interval,omitempty"`

	// MaxElapsedTime represents the maximal amount of time allowed to retry. A
	// value of zero signifies no limit
	MaxElapsedTime time.Duration `json:"max_elapsed_time,omitempty"`

	// MaxRetryAttempts is the maximal number of retries before exiting with
	// an error. A value of zero signifies unlimited retry attemps
	MaxRetryAttempts int `json:"max_retry_attempts,omitempty"`

	// Multiplier is used to increment the current interval by multiplying it with
	// this multiplier. If not supplied, it will be set to DefaultMultiplier.
	Multiplier float64 `json:"multiplier"`

	// start contains the starting time of the retry attempts
	start time.Time
}

func (e *ExponentialBackoff) UnmarshalJSON(b []byte) error {
	blob := map[string]*json.RawMessage{}
	if err := json.Unmarshal(b, &blob); err != nil {
		return err
	}
	if maxRetry, ok := blob["max_retry_attempts"]; ok {
		if err := json.Unmarshal(*maxRetry, &e.MaxRetryAttempts); err != nil {
			return err
		}
	}
	if mult, ok := blob["multiplier"]; ok {
		if err := json.Unmarshal(*mult, &e.Multiplier); err != nil {
			return err
		}
	}
	if initDelay, ok := blob["initial_delay_interval"]; ok {
		var td JSONTimeDuration
		if err := json.Unmarshal(*initDelay, &td); err != nil {
			return err
		}
		e.InitialDelayInterval = time.Duration(td)
	}
	if maxDelay, ok := blob["max_delay_interval"]; ok {
		var td JSONTimeDuration
		if err := json.Unmarshal(*maxDelay, &td); err != nil {
			return err
		}
		e.MaxDelayInterval = time.Duration(td)
	}
	if maxElapsed, ok := blob["max_elapsed_time"]; ok {
		var td JSONTimeDuration
		if err := json.Unmarshal(*maxElapsed, &td); err != nil {
			return err
		}
		e.MaxElapsedTime = time.Duration(td)
	}
	return nil
}

func (e ExponentialBackoff) MarshalJSON() (out []byte, err error) {
	type ebFacade struct {
		Ctx                  context.Context  `json:"-"`
		InitialDelayInterval JSONTimeDuration `json:"initial_delay_interval,omitempty"`
		MaxDelayInterval     JSONTimeDuration `json:"max_delay_interval,omitempty"`
		MaxElapsedTime       JSONTimeDuration `json:"max_elapsed_time,omitempty"`
		MaxRetryAttempts     int              `json:"max_retry_attempts,omitempty"`
		Multiplier           float64          `json:"multiplier"`
		start                time.Time
	}
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("couldn't marshal value: %s", err)
		}
	}()
	var eb = (*ebFacade)(unsafe.Pointer(&e))
	return json.Marshal(eb)
}

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
	wait := time.Duration(b.InitialDelayInterval)
	ctx := context.Background()
	if b.Ctx != nil {
		ctx = b.Ctx
	}

	for i := 0; i < b.MaxRetryAttempts || b.MaxRetryAttempts == 0; i++ {
		if i != 0 {
			// Verify if we reached the MaxElapsedTime
			if b.MaxElapsedTime != 0 && time.Since(b.start) > time.Duration(b.MaxElapsedTime) {
				return ErrMaxElapsedTime
			}

			// Sleep for the determined duration
			if b.MaxDelayInterval > 0 && wait > time.Duration(b.MaxDelayInterval) {
				wait = time.Duration(b.MaxDelayInterval)
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(wait)):
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
