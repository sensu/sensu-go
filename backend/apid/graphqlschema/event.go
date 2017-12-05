package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var eventType *graphql.Union

//
// NOTE:
//
//   Since (at this time) a union's types cannot be wrapped in a thunk, we
//   need to make sure that the checkEventType & metricEventType have been
//   instantiated before the EventType is.
//
func init() {
	initCheckEventType()
	initMetricEventType()
	initEventType()
}

func initEventType() {
	eventType = graphql.NewUnion(graphql.UnionConfig{
		Name:        "Event",
		Description: "Describes check result or metric collected.",
		Types: []*graphql.Object{
			checkEventType,
			metricEventType,
		},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
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
