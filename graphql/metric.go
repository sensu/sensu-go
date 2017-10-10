package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
	"github.com/sensu/sensu-go/types"
)

var metricEventType *graphql.Object

func init() {
	metricEventType = graphql.NewObject(graphql.ObjectConfig{
		Name: "MetricEvent",
		Fields: graphql.Fields{
			"id":        relay.GlobalIDField("MetricEvent", nil),
			"entity":    &graphql.Field{Type: entityType},
			"timestamp": &graphql.Field{Type: timeScalar},
		},
		IsTypeOf: func(p graphql.IsTypeOfParams) bool {
			if e, ok := p.Value.(*types.Event); ok {
				return e.Metrics != nil
			}
			return false
		},
	})
}
