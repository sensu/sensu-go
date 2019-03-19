package filter

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/fields"

	"k8s.io/apimachinery/pkg/selection"

	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/js"
	"github.com/sensu/sensu-go/types/dynamic"
	"k8s.io/apimachinery/pkg/labels"
)

func BenchmarkOtto(b *testing.B) {
	filter := "event.check.name == 'check42'"

	for i := 0; i < b.N; i++ {
		rand.Seed(time.Now().UnixNano())
		checkName := fmt.Sprintf("check%d", rand.Intn((1000/2)-1)+1)
		event := v2.FixtureEvent("foo", checkName)

		synth := dynamic.Synthesize(event)
		parameters := map[string]interface{}{"event": synth}
		_, _ = js.Evaluate(filter, parameters, nil)
	}
}

func BenchmarkLabelSelector(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rand.Seed(time.Now().UnixNano())
		checkName := fmt.Sprintf("check%d", rand.Intn((1000/2)-1)+1)
		event := v2.FixtureEvent("foo", checkName)
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
		rand.Seed(time.Now().UnixNano())
		checkName := fmt.Sprintf("check%d", rand.Intn((1000/2)-1)+1)
		event := v2.FixtureEvent("foo", checkName)

		fieldSet := eventsField(event)
		selector, _ := fields.ParseSelector("event.check.name=check42")
		_ = selector.Matches(fieldSet)
	}
}
