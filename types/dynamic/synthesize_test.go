package dynamic_test

import (
	"reflect"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/dynamic"
)

func TestSynthesizeEvent(t *testing.T) {
	event := types.FixtureEvent("foo", "bar")
	synth := dynamic.Synthesize(event).(map[string]interface{})
	if !reflect.DeepEqual(event.HasCheck(), synth["has_check"]) {
		t.Fatal("bad synthesis")
	}
}
