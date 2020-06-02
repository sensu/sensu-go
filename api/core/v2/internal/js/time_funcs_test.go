package js_test

import (
	"fmt"
	"testing"

	time "github.com/echlebek/timeproxy"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/api/core/v2/internal/js"
	"github.com/sensu/sensu-go/api/core/v2/internal/types/dynamic"
)

func TestTimeFuncs(t *testing.T) {
	check := corev2.FixtureCheck("foo")
	synth := dynamic.Synthesize(check)
	expr := fmt.Sprintf("hour(executed) == %d", time.Unix(check.Executed, 0).UTC().Hour())
	result, err := js.Evaluate(expr, synth, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("result should be true")
	}
}
