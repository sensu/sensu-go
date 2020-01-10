package retry

import (
	"encoding/json"
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

func mockBackoffFuncSleep() func(retry int) (bool, error) {
	return func(retry int) (bool, error) {
		time.Sleep(50 * time.Millisecond)
		return false, nil
	}
}

func TestExponentialBackoff(t *testing.T) {
	// It should reach the MaxRetryAttempts
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

	// It should reach the MaxElapsedTime, since our mockBackoffFuncSleep func
	// sleeps for 50ms and we have 3 retry attempts
	b.MaxElapsedTime = 60 * time.Millisecond
	sleepFn := mockBackoffFuncSleep()
	assert.Equal(t, ErrMaxElapsedTime, b.Retry(sleepFn))
}

func TestJSONTimeDurationUnmarshal(t *testing.T) {
	data := []byte(`"5s"`)
	var tm JSONTimeDuration
	if err := json.Unmarshal(data, &tm); err != nil {
		t.Fatal(err)
	}
	if got, want := tm, JSONTimeDuration(time.Second*5); got != want {
		t.Fatalf("bad JSONTimeDuration: got %s, want %s", got, want)
	}
}

func TestJSONTimeDurationMarshal(t *testing.T) {
	tm := JSONTimeDuration(time.Second * 5)
	b, err := json.Marshal(tm)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(b), `"5s"`; got != want {
		t.Fatalf("bad JSONTimeDuration: got %s, want %s", got, want)
	}
}

func TestUnmarshalExponentialBackoff(t *testing.T) {
	var doc = []byte(`{
	"initial_delay_interval": "10ms",
	"max_delay_interval": 0,
	"max_elapsed_time": "1h",
	"max_retry_attempts": 5,
	"multiplier": 1.5
	}`)
	var eb ExponentialBackoff
	if err := json.Unmarshal(doc, &eb); err != nil {
		t.Fatal(err)
	}
	if got, want := eb.InitialDelayInterval, 10*time.Millisecond; got != want {
		t.Fatalf("bad InitialDelayInterval: got %s, want %s", got, want)
	}
	if got, want := eb.MaxDelayInterval, time.Duration(0); got != want {
		t.Fatalf("bad MaxDelayInterval: got %s, want %s", got, want)
	}
	if got, want := eb.MaxElapsedTime, time.Hour; got != want {
		t.Fatalf("bad MaxElapsedTime: got %s, want %s", got, want)
	}
}

func TestMarshalExponentialBackoff(t *testing.T) {
	eb := &ExponentialBackoff{
		InitialDelayInterval: time.Second,
		Multiplier:           1.5,
	}
	b, err := json.Marshal(eb)
	if err != nil {
		t.Fatal(err)
	}
	rm := make(map[string]*json.RawMessage)
	if err := json.Unmarshal(b, &rm); err != nil {
		t.Fatal(err)
	}
	if _, ok := rm["multiplier"]; !ok {
		t.Fatal("missing multiplier")
	}
	if _, ok := rm["initial_delay_interval"]; !ok {
		t.Fatal("missing initial_delay_interval")
	}
}
