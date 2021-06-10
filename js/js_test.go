package js

import (
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types/dynamic"
	"github.com/stretchr/testify/assert"
)

// This is a unit test to cover the race condition found in
// https://github.com/sensu/sensu-go/issues/4073
func TestEvaluateRaceCondition(t *testing.T) {
	entity := corev2.FixtureEntity("foo")
	synth := dynamic.Synthesize(entity)
	params := map[string]interface{}{"entity": synth}

	go func() {
		_, _ = Evaluate("true", params, nil)
	}()
	go func() {
		_, _ = Evaluate("true", params, nil)
	}()
}

func TestTimeFunctions(t *testing.T) {
	timestamp := int64(7323)
	t.Run("hour() function", func(t *testing.T) {
		entity := corev2.FixtureEvent("foo", "bar")
		entity.Timestamp = timestamp
		synth := dynamic.Synthesize(entity)
		params := map[string]interface{}{"event": synth}

		result, err := Evaluate("hour(event.timestamp) == 2", params, nil)
		assert.NoError(t, err)
		assert.Equal(t, true, result)
	})
	t.Run("minute() function", func(t *testing.T) {
		entity := corev2.FixtureEvent("foo", "bar")
		entity.Timestamp = timestamp
		synth := dynamic.Synthesize(entity)
		params := map[string]interface{}{"event": synth}

		result, err := Evaluate("minute(event.timestamp) == 2", params, nil)
		assert.NoError(t, err)
		assert.Equal(t, true, result)
	})
	t.Run("second() function", func(t *testing.T) {
		entity := corev2.FixtureEvent("foo", "bar")
		entity.Timestamp = timestamp
		synth := dynamic.Synthesize(entity)
		params := map[string]interface{}{"event": synth}

		result, err := Evaluate("second(event.timestamp) == 3", params, nil)
		assert.NoError(t, err)
		assert.Equal(t, true, result)
	})
	t.Run("seconds_since() function", func(t *testing.T) {
		entity := corev2.FixtureEvent("foo", "bar")
		entity.Timestamp = timestamp
		synth := dynamic.Synthesize(entity)
		params := map[string]interface{}{"event": synth}

		result, err := Evaluate("seconds_since(event.timestamp) > 10000", params, nil)
		assert.NoError(t, err)
		assert.Equal(t, true, result)
	})
}
