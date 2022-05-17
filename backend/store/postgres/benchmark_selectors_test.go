package postgres

import (
	"testing"

	"github.com/sensu/sensu-go/backend/selector"
)

func BenchmarkCreateGetEventsQuery(b *testing.B) {
	namespace := "default"
	entity := "entity"
	check := "check"
	selector := &selector.Selector{
		Operations: []selector.Operation{
			{
				LValue:   "foo",
				Operator: selector.InOperator,
				RValues:  []string{"labels.bar"},
			},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = CreateGetEventsQuery(namespace, entity, check, selector, nil)
	}
}
