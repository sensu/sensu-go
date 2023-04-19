package dynamic_test

import (
	"reflect"
	"testing"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/dynamic"
)

func TestSynthesizeEvent(t *testing.T) {
	event := v2.FixtureEvent("foo", "bar")
	synth := dynamic.Synthesize(event).(map[string]interface{})
	if !reflect.DeepEqual(event.HasCheck(), synth["has_check"]) {
		t.Fatal("bad synthesis")
	}
}

type hasMethods struct {
	callLog map[string]bool
}

func (h *hasMethods) Meth1() {
	if h.callLog == nil {
		h.callLog = make(map[string]bool)
	}
	h.callLog["Meth1"] = true
}

func (h *hasMethods) Meth2() {
	if h.callLog == nil {
		h.callLog = make(map[string]bool)
	}
	h.callLog["Meth2"] = true
}

func TestSynthesizeMethods(t *testing.T) {
	hm := new(hasMethods)
	synth := dynamic.SynthesizeMethods(hm)
	if got, want := len(synth), 2; got != want {
		t.Fatalf("wrong length: got %d, want %d", got, want)
	}
	for _, v := range synth {
		v.(func())()
	}
	exp := map[string]bool{
		"Meth1": true,
		"Meth2": true,
	}
	if got, want := hm.callLog, exp; !reflect.DeepEqual(got, want) {
		t.Errorf("didn't get expected calls: got %v, want %v", got, want)
	}
}
