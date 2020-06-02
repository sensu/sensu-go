package dynamic_test

import (
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/api/core/v2/internal/types/dynamic"
)

func TestSynthesizeEvent(t *testing.T) {
	event := corev2.FixtureEvent("foo", "bar")
	synth := dynamic.Synthesize(event).(map[string]interface{})
	if !reflect.DeepEqual(event.HasCheck(), synth["has_check"]) {
		t.Fatal("bad synthesis")
	}
}
