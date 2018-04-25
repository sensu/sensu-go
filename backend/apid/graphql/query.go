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
	eventCtrl       eventFinder
	environmentCtrl environmentFinder
	nodeResolver    *nodeResolver
}

func newQueryImpl(store store.Store, resolver *nodeResolver) *queryImpl {
	return &queryImpl{
		eventCtrl:       actions.NewEventController(store, nil),
		environmentCtrl: actions.NewEnvironmentController(store),
		nodeResolver:    resolver,
	}
}

// Viewer implements response to request for 'viewer' field.
func (r *queryImpl) Viewer(p graphql.ResolveParams) (interface{}, error) {
	return struct{}{}, nil
}

// Environment implements response to request for 'environment' field.
func (r *queryImpl) Environment(p schema.QueryEnvironmentFieldResolverParams) (interface{}, error) {
	env, err := r.environmentCtrl.Find(p.Context, p.Args.Organization, p.Args.Environment)
	return handleControllerResults(env, err)
}

// Event implements response to request for 'event' field.
func (r *queryImpl) Event(p schema.QueryEventFieldResolverParams) (interface{}, error) {
	ctx := types.SetContextFromResource(p.Context, p.Args.Ns)
	event, err := r.eventCtrl.Find(ctx, p.Args.Entity, p.Args.Check)
	return handleControllerResults(event, err)
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
