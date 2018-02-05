package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
)

var _ schema.QueryFieldResolvers = (*queryImpl)(nil)

//
// Implement QueryFieldResolvers
//

type queryImpl struct {
	store        store.Store
	nodeResolver *nodeResolver
}

func newQueryImpl(store store.Store, resolver *nodeResolver) *queryImpl {
	return &queryImpl{
		store:        store,
		nodeResolver: resolver,
	}
}

// Viewer implements response to request for 'viewer' field.
func (r *queryImpl) Viewer(p graphql.ResolveParams) (interface{}, error) {
	return struct{}{}, nil
}

// Node implements response to request for 'node' field.
func (r *queryImpl) Node(p schema.QueryNodeFieldResolverParams) (interface{}, error) {
	resolver := r.nodeResolver
	id := p.Args.ID.(string)
	return resolver.Find(p.Context, id, p.Info)
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
