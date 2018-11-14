package js_test

import (
	"fmt"
	"testing"

	time "github.com/echlebek/timeproxy"
	"github.com/sensu/sensu-go/js"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/dynamic"
)

func TestTimeFuncs(t *testing.T) {
	check := types.FixtureCheck("foo")
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
