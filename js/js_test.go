package js_test

import (
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/js"
	"github.com/sensu/sensu-go/types/dynamic"
)

func TestEvaluateEntityFilters(t *testing.T) {
	entity := dynamic.Synthesize(corev2.FixtureEntity("foo"))
	expression := "entity.system.arch == 'amd64'"
	entities := []interface{}{entity}
	expressions := []string{expression}

	results, err := js.EvaluateEntityFilters(expressions, entities)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(results), 1; got != want {
		t.Fatalf("wrong number of results: got %d, want %d", got, want)
	}
}

func BenchmarkEvaluateEntityFilters1000(b *testing.B) {
	entity := dynamic.Synthesize(corev2.FixtureEntity("foo"))
	expression := "entity.system.arch == 'amd64'"

	entities := make([]interface{}, 1000)
	expressions := make([]string, 1000)

	for i := 0; i < 1000; i++ {
		entities[i] = entity
		expressions[i] = expression
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = js.EvaluateEntityFilters(expressions, entities)
	}
}

func BenchmarkEvaluate1000(b *testing.B) {
	entity := map[string]interface{}{
		"entity": dynamic.Synthesize(corev2.FixtureEntity("foo")),
	}
	expression := "entity.system.arch == 'amd64'"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 1000; j++ {
			_, _ = js.Evaluate(expression, entity, nil)
		}
	}
}
