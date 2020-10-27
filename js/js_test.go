package js

import (
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types/dynamic"
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
