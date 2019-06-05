package js_test

import (
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/js"
	"github.com/sensu/sensu-go/types/dynamic"
)

func BenchmarkMatchEntities1000(b *testing.B) {
	entity := dynamic.Synthesize(corev2.FixtureEntity("foo"))
	// non-matching expression to avoid short-circuiting behaviour
	expression := "entity.system.arch == 'amd65'"

	entities := make([]interface{}, 100)
	expressions := make([]string, 10)

	for i := range entities {
		entities[i] = entity
	}
	for i := range expressions {
		expressions[i] = expression
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = js.MatchEntities(expressions, entities)
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
