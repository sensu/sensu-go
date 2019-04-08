package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var _ schema.QueryFieldResolvers = (*queryImpl)(nil)

//
// Implement QueryFieldResolvers
//

type queryImpl struct {
	nodeResolver *nodeResolver
	factory      ClientFactory
}

// Viewer implements response to request for 'viewer' field.
func (r *queryImpl) Viewer(p graphql.ResolveParams) (interface{}, error) {
	return struct{}{}, nil
}

// Environment implements response to request for 'namespace' field.
func (r *queryImpl) Namespace(p schema.QueryNamespaceFieldResolverParams) (interface{}, error) {
	client := r.factory.NewWithContext(p.Context)
	res, err := client.FetchNamespace(p.Args.Name)
	return handleFetchResult(res, err)
}

// Event implements response to request for 'event' field.
func (r *queryImpl) Event(p schema.QueryEventFieldResolverParams) (interface{}, error) {
	ctx := contextWithNamespace(p.Context, p.Args.Namespace)
	client := r.factory.NewWithContext(ctx)
	res, err := client.FetchEvent(p.Args.Entity, p.Args.Check)
	return handleFetchResult(res, err)
}

// Entity implements response to request for 'entity' field.
func (r *queryImpl) Entity(p schema.QueryEntityFieldResolverParams) (interface{}, error) {
	ctx := contextWithNamespace(p.Context, p.Args.Namespace)
	client := r.factory.NewWithContext(ctx)
	res, err := client.FetchEntity(p.Args.Name)
	return handleFetchResult(res, err)
}

// Check implements response to request for 'check' field.
func (r *queryImpl) Check(p schema.QueryCheckFieldResolverParams) (interface{}, error) {
	ctx := contextWithNamespace(p.Context, p.Args.Namespace)
	client := r.factory.NewWithContext(ctx)
	res, err := client.FetchCheck(p.Args.Name)
	return handleFetchResult(res, err)
}

// Node implements response to request for 'node' field.
func (r *queryImpl) Node(p schema.QueryNodeFieldResolverParams) (interface{}, error) {
	resolver := r.nodeResolver
	return resolver.Find(p.Context, p.Args.ID, p.Info)
}

// WrappedNode implements response to request for 'wrappedNode' field.
func (r *queryImpl) WrappedNode(p schema.QueryWrappedNodeFieldResolverParams) (interface{}, error) {
	resolver := r.nodeResolver
	res, err := resolver.Find(p.Context, p.Args.ID, p.Info)
	if rres, ok := res.(types.Resource); ok {
		return types.WrapResource(rres), err
	}
	return nil, err
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
