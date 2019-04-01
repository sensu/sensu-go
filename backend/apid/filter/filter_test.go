package filter

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/js"
	"github.com/sensu/sensu-go/types/dynamic"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

// Tweak this number as required
const eventsCount = 1000000

var events = make([]*v2.Event, eventsCount)

func init() {
	for i := 0; i < eventsCount; i++ {
		rand.Seed(time.Now().UnixNano())
		checkName := fmt.Sprintf("check%d", rand.Intn((1000/2)-1)+1)
		events[i] = v2.FixtureEvent("foo", checkName)
	}
}
func BenchmarkOtto(b *testing.B) {
	filter := "event.check.name == 'check42'"

	for i := 0; i < b.N; i++ {
		event := events[i]

		synth := dynamic.Synthesize(event)
		parameters := map[string]interface{}{"event": synth}
		_, _ = js.Evaluate(filter, parameters, nil)
	}
}

func BenchmarkOttoWithReusedVM(b *testing.B) {
	filter := "event.check.name == 'check42'"
	vm := otto.New()

	for i := 0; i < b.N; i++ {
		event := events[i]

		synth := dynamic.Synthesize(event)
		parameters := map[string]interface{}{"event": synth}

		for name, value := range parameters {
			if err := vm.Set(name, value); err != nil {
				b.Fatal(err)
			}
		}

		value, _ := vm.Run(filter)
		_, _ = value.ToBoolean()

		// Cleanup
		for name := range parameters {
			_ = vm.Set(name, otto.UndefinedValue())
		}
	}
}

func BenchmarkOttoWithoutSynthesize(b *testing.B) {
	filter := "event.Check.ObjectMeta.Name == 'check42'"
	vm := otto.New()

	for i := 0; i < b.N; i++ {
		event := events[i]

		parameters := map[string]interface{}{"event": event}

		for name, value := range parameters {
			if err := vm.Set(name, value); err != nil {
				b.Fatal(err)
			}
		}

		value, _ := vm.Run(filter)
		_, _ = value.ToBoolean()

		// Cleanup
		for name := range parameters {
			_ = vm.Set(name, otto.UndefinedValue())
		}
	}
}

func BenchmarkLabelSelector(b *testing.B) {
	for i := 0; i < b.N; i++ {
		event := events[i]

		event.Labels["check"] = event.Check.ObjectMeta.Name
		event.Labels["entity"] = event.Entity.ObjectMeta.Name
		event.Labels["status"] = string(event.Check.Status)
		// TODO: figure out how to deal with lists
		//event.Labels["subscriptions"] = strings.Join(event.Check.Subscriptions, ",")

		labelSet := labels.Set(event.Labels)
		requirement, _ := labels.NewRequirement("check", selection.Equals, []string{"check42"})
		_ = requirement.Matches(labelSet)
	}
}

func BenchmarkFieldSelector(b *testing.B) {
	for i := 0; i < b.N; i++ {
		event := events[i]

		fieldSet := eventsField(event)
		selector, _ := fields.ParseSelector("event.check.name=check42")
		_ = selector.Matches(fieldSet)
	}
}

func TestDurationOtto(t *testing.T) {
	filter := "event.check.name == 'check42'"

	for i := 0; i < eventsCount; i++ {
		event := events[i]

		synth := dynamic.Synthesize(event)
		parameters := map[string]interface{}{"event": synth}
		_, _ = js.Evaluate(filter, parameters, nil)
	}
}

func TestDurationOttoWithReusedVM(t *testing.T) {
	filter := "event.check.name == 'check42'"

	vm := otto.New()

	for i := 0; i < eventsCount; i++ {
		event := events[i]

		synth := dynamic.Synthesize(event)
		parameters := map[string]interface{}{"event": synth}

		for name, value := range parameters {
			if err := vm.Set(name, value); err != nil {
				t.Fatal(err)
			}
		}

		_, _ = vm.Run(filter)

		// Cleanup
		for name := range parameters {
			_ = vm.Set(name, otto.UndefinedValue())
		}
	}
}

func TestDurationOttoWithoutSynthesize(t *testing.T) {
	filter := "event.Check.ObjectMeta.Name == 'check42'"
	vm := otto.New()

	for i := 0; i < eventsCount; i++ {
		event := events[i]

		parameters := map[string]interface{}{"event": event}

		for name, value := range parameters {
			if err := vm.Set(name, value); err != nil {
				t.Fatal(err)
			}
		}

		value, _ := vm.Run(filter)
		_, _ = value.ToBoolean()

		// Cleanup
		for name := range parameters {
			_ = vm.Set(name, otto.UndefinedValue())
		}
	}
}

func TestDurationLabelSelector(t *testing.T) {
	for i := 0; i < eventsCount; i++ {
		event := events[i]

		event.Labels["check"] = event.Check.ObjectMeta.Name
		event.Labels["entity"] = event.Entity.ObjectMeta.Name
		event.Labels["status"] = string(event.Check.Status)
		// TODO: figure out how to deal with lists
		//event.Labels["subscriptions"] = strings.Join(event.Check.Subscriptions, ",")

		labelSet := labels.Set(event.Labels)
		requirement, _ := labels.NewRequirement("check", selection.Equals, []string{"check42"})
		_ = requirement.Matches(labelSet)
	}
}

func TestDurationFieldSelector(t *testing.T) {
	for i := 0; i < eventsCount; i++ {
		event := events[i]

		fieldSet := eventsField(event)
		selector, _ := fields.ParseSelector("event.check.name=check42")
		_ = selector.Matches(fieldSet)
	}
}
