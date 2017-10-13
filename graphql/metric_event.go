package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/graphql/globalid"
	"github.com/sensu/sensu-go/graphql/relay"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

var metricEventType *graphql.Object

func initMetricEventType() {
	if metricEventType != nil {
		return
	}

	metricEventType = graphql.NewObject(graphql.ObjectConfig{
		Name: "MetricEvent",
		Interfaces: graphql.InterfacesThunk(func() []*graphql.Interface {
			return []*graphql.Interface{
				nodeInterface,
			}
		}),
		Fields: graphql.FieldsThunk(func() graphql.Fields {
			return graphql.Fields{
				"id": &graphql.Field{
					Name:        "id",
					Description: "The ID of an object",
					Type:        graphql.NewNonNull(graphql.ID),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						idComponents := globalid.EventResource.Encode(p.Source)
						return idComponents.String(), nil
					},
				},
				"timestamp": &graphql.Field{
					Type:        timeScalar,
					Description: "Time in-which event occurred.",
				},
			}
		}),
		IsTypeOf: func(p graphql.IsTypeOfParams) bool {
			if e, ok := p.Value.(*types.Event); ok {
				return e.Metrics != nil
			}
			return false
		},
	})

	nodeRegister.RegisterResolver(relay.NodeResolver{
		Object:     checkEventType,
		Translator: globalid.EventResource,
		Resolve: func(ctx context.Context, c globalid.Components) (interface{}, error) {
			// components := c.(globalid.EventComponents)
			// store := ctx.Value(types.StoreKey).(store.EventStore)

			// TODO: Implement along side metrics
			return nil, nil
		},
		IsKindOf: func(components globalid.Components) bool {
			return components.ResourceType() == "check"
		},
	})
}
