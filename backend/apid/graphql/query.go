package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
)

//
// Implement QueryFieldResolvers
//

type queryImpl struct {
	store        store.Store
	nodeResolver *nodeResolver
}

// Checks implements response to request for 'checks' field.
func (r *queryImpl) Checks(p graphql.ResolveParams) (interface{}, error) {
	// TODO: Remove me
	checks, err := r.store.GetCheckConfigs(p.Context)
	return checks, err
}

// Node implements response to request for 'node' field.
func (r *queryImpl) Node(p schema.QueryNodeFieldResolverParams) (interface{}, error) {
	resolver := r.nodeResolver
	id := p.Args.ID.(string)
	return resolver.Find(p.Context, id, p.Info)
}

// IsTypeOf is used to determine if a given value is associated with the type
func (r *queryImpl) IsTypeOf(_ interface{}, _ graphql.IsTypeOfParams) bool {
	return false
}

//
// Implement Node interface
//

type nodeImpl struct {
	nodeResolver *nodeResolver
}

func (impl *nodeImpl) ResolveType(i interface{}, _ graphql.ResolveTypeParams) *graphql.Type {
	resolver := impl.nodeResolver
	return resolver.FindType(i)
}
