package graphqlschema

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql/globalid"
	"github.com/sensu/sensu-go/graphql/relay"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

var checkConfigType *graphql.Object
var checkEventType *graphql.Object
var checkEventConnection *relay.ConnectionDefinitions

func init() {
	checkConfigType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Check",
		Description: "The `Check` object type represents  the specification of a check",
		Interfaces: []*graphql.Interface{
			nodeInterface,
			multitenantInterface,
		},
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Name:        "id",
				Description: "The ID of an object",
				Type:        graphql.NewNonNull(graphql.ID),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					idComponents := globalid.CheckResource.Encode(p.Source)
					return idComponents.String(), nil
				},
			},
			"name":          &graphql.Field{Type: graphql.String},
			"interval":      &graphql.Field{Type: graphql.Int},
			"subscriptions": &graphql.Field{Type: graphql.NewList(graphql.String)},
			"command":       &graphql.Field{Type: graphql.String},
			"handlers":      &graphql.Field{Type: graphql.NewList(graphql.String)},
			"runtimeAssets": &graphql.Field{Type: graphql.NewList(graphql.String)},
			"environment":   &graphql.Field{Type: graphql.String},
			"organization":  &graphql.Field{Type: graphql.String},
		},
	})

	nodeRegister.RegisterResolver(relay.NodeResolver{
		Object:     checkConfigType,
		Translator: globalid.CheckResource,
		Resolve: func(ctx context.Context, c globalid.Components) (interface{}, error) {
			components := c.(globalid.NamedComponents)
			store := ctx.Value(types.StoreKey).(store.CheckConfigStore)

			// TODO: Filter out unauthorized results
			record, err := store.GetCheckConfigByName(ctx, components.Name())
			return record, err
		},
	})

	checkEventType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "CheckEvent",
		Description: "A check result",
		Interfaces: []*graphql.Interface{
			nodeInterface,
		},
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Name:        "id",
				Description: "The ID of an object",
				Type:        graphql.NewNonNull(graphql.ID),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					idComponents := globalid.EventResource.Encode(p.Source)
					return idComponents.String(), nil
				},
			},
			"timestamp": &graphql.Field{Type: timeScalar},
			"entity":    &graphql.Field{Type: entityType},
			"output":    AliasField(graphql.String, "Check", "Output"),
			"status":    AliasField(graphql.Int, "Check", "Status"),
			"issued":    AliasField(timeScalar, "Check", "Issued"),
			"executed":  AliasField(timeScalar, "Check", "Executed"),
			"config":    AliasField(checkConfigType, "Check", "Config"),
		},
		IsTypeOf: func(p graphql.IsTypeOfParams) bool {
			if e, ok := p.Value.(*types.Event); ok {
				return e.Check != nil
			}
			return false
		},
	})

	nodeRegister.RegisterResolver(relay.NodeResolver{
		Object:     checkEventType,
		Translator: globalid.EventResource,
		Resolve: func(ctx context.Context, c globalid.Components) (interface{}, error) {
			components := c.(globalid.EventComponents)
			store := ctx.Value(types.StoreKey).(store.EventStore)

			// TODO: Why does GetEventByEntityCheck only return a single event?!
			events, err := store.GetEventsByEntity(ctx, components.EntityName())
			if err != nil {
				return nil, err
			}

			// TODO: Filter out unauthorized results
			for _, event := range events {
				if event.Timestamp == components.Timestamp() &&
					event.Check.Config.Name == components.CheckName() {
					return event, nil
				}
			}

			return nil, errors.New("event not found")
		},
	})

	checkEventConnection = relay.NewConnectionDefinition(relay.ConnectionConfig{
		Name:     "Check",
		NodeType: checkConfigType,
	})
}
