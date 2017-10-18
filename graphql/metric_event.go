package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/graphql/globalid"
	"github.com/sensu/sensu-go/graphql/relay"
	"github.com/sensu/sensu-go/types"
)

var metricEventType *graphql.Object
var metricEventConnection *relay.ConnectionDefinitions

func init() {
	initNodeInterface()
	initMetricEventType()
	initMetricEventConnection()

	nodeResolver := newMetricEventNodeResolver()
	nodeRegister.RegisterResolver(nodeResolver)
}

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
						idComponents := globalid.EventTranslator.Encode(p.Source)
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
}

func initMetricEventConnection() {
	if metricEventConnection != nil {
		return
	}

	metricEventConnection = relay.NewConnectionDefinition(relay.ConnectionConfig{
		Name:     "MetricEvent",
		NodeType: metricEventType,
	})
}

func newMetricEventNodeResolver() relay.NodeResolver {
	return relay.NodeResolver{
		Object:     metricEventType,
		Translator: globalid.EventTranslator,
		Resolve: func(p relay.NodeResolverParams) (interface{}, error) {
			// components := p.IDComponents.(globalid.EventComponents)
			// store := p.Context.Value(types.StoreKey).(store.EventStore)

			// TODO: Implement along side metrics
			return nil, nil
		},
		IsKindOf: func(components globalid.Components) bool {
			return components.ResourceType() == "check"
		},
	}
}
