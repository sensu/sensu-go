package graphqlschema

import (
	"errors"
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/graphql/globalid"
	"github.com/sensu/sensu-go/graphql/relay"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

var queryType *graphql.Object

func init() {
	queryType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.FieldsThunk(func() graphql.Fields {
			return graphql.Fields{
				"node": &graphql.Field{
					Description: "Fetches an object given its ID",
					Type:        nodeInterface,
					Args: graphql.FieldConfigArgument{
						"id": &graphql.ArgumentConfig{
							Type:        graphql.NewNonNull(graphql.ID),
							Description: "The ID of an object",
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						var id string
						if iid, ok := p.Args["id"]; ok {
							id = fmt.Sprintf("%v", iid)
						}

						// Parse given ID
						idComponents, err := globalid.Decode(id)
						if err != nil {
							return nil, err
						}

						// Lookup resolver using components of global ID
						resolver := nodeRegister.Lookup(idComponents)
						if resolver == nil {
							return nil, errors.New("unable to find GraphQL type associated with given ID")
						}

						// Lift org & env into context
						ctx := p.Context
						ctx = context.WithValue(ctx, types.OrganizationKey, idComponents.Organization())
						ctx = context.WithValue(ctx, types.EnvironmentKey, idComponents.Environment())

						// Fetch resource from store
						params := relay.NodeResolverParams{
							Context:      ctx,
							IDComponents: idComponents,
							Info:         p.Info,
						}
						record, err := resolver.Resolve(params)
						return record, err
					},
				},

				"viewer": &graphql.Field{
					Type:        viewerType,
					Description: "Describes the currently authorized viewer",
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						// TODO? User? Viewer warpper type?
						return struct{}{}, nil
					},
				},
			}
		}),
	})
}
