package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var _ schema.QueryFieldResolvers = (*queryImpl)(nil)

//
// Implement QueryFieldResolvers
//

type queryImpl struct {
	eventFinder  eventFinder
	entityFinder entityFinder
	checkFinder  checkFinder
	nsFinder     namespaceFinder

	nodeResolver *nodeResolver
}

func newQueryImpl(store store.Store, resolver *nodeResolver, queue types.QueueGetter) *queryImpl {
	return &queryImpl{
		eventFinder:  actions.NewEventController(store, nil),
		entityFinder: actions.NewEntityController(store),
		checkFinder:  actions.NewCheckController(store, queue),
		nsFinder:     actions.NewNamespacesController(store),
		nodeResolver: resolver,
	}
}

// Viewer implements response to request for 'viewer' field.
func (r *queryImpl) Viewer(p graphql.ResolveParams) (interface{}, error) {
	return struct{}{}, nil
}

// Environment implements response to request for 'namespace' field.
func (r *queryImpl) Namespace(p schema.QueryNamespaceFieldResolverParams) (interface{}, error) {
	env, err := r.nsFinder.Find(p.Context, p.Args.Name)
	return handleControllerResults(env, err)
}

// Event implements response to request for 'event' field.
func (r *queryImpl) Event(p schema.QueryEventFieldResolverParams) (interface{}, error) {
	ctx := contextWithNamespace(p.Context, p.Args.Namespace)
	event, err := r.eventFinder.Find(ctx, p.Args.Entity, p.Args.Check)
	return handleControllerResults(event, err)
}

// Entity implements response to request for 'entity' field.
func (r *queryImpl) Entity(p schema.QueryEntityFieldResolverParams) (interface{}, error) {
	ctx := contextWithNamespace(p.Context, p.Args.Namespace)
	entity, err := r.entityFinder.Find(ctx, p.Args.Name)
	return handleControllerResults(entity, err)
}

// Check implements response to request for 'check' field.
func (r *queryImpl) Check(p schema.QueryCheckFieldResolverParams) (interface{}, error) {
	ctx := contextWithNamespace(p.Context, p.Args.Namespace)
	check, err := r.checkFinder.Find(ctx, p.Args.Name)
	return handleControllerResults(check, err)
}

// Node implements response to request for 'node' field.
func (r *queryImpl) Node(p schema.QueryNodeFieldResolverParams) (interface{}, error) {
	resolver := r.nodeResolver
	return resolver.Find(p.Context, p.Args.ID, p.Info)
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
