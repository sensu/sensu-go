package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/graphql/globalid"
	"github.com/sensu/sensu-go/graphql/relay"
)

var nodeInterface *graphql.Interface
var nodeRegister = relay.NodeRegister{}

func init() {
	nodeInterface = graphql.NewInterface(graphql.InterfaceConfig{
		Name:        "Node",
		Description: "An object with an ID",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.ID),
				Description: "The id of the object",
			},
		},
		//
		// TODO:
		//
		// I believe calls to this closure are relatively frequent and this
		// process is convoluted and a far cry from optimal. Likely place to focus
		// for future optimizations.
		//
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			if translator, err := globalid.ReverseLookup(p.Value); err != nil {
				components := translator.Encode(p.Value)
				resolver := nodeRegister.Lookup(components)
				return resolver.Object
			}
			return nil
		},
	})
}
