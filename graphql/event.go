package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var eventType *graphql.Union

func init() {
	initCheckEventType()
	initMetricEventType()

	eventType = graphql.NewUnion(graphql.UnionConfig{
		Name:        "Event",
		Description: "Describes check result or metric collected.",
		Types: []*graphql.Object{
			checkEventType,
			metricEventType,
		},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			// TODO
			if event, ok := p.Value.(*types.Event); ok {
				if event.Check != nil {
					return checkEventType
				} else if event.Metrics != nil {
					return metricEventType
				}
			}

			return nil
		},
	})
}
