package graphql

import (
	"context"

	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var _ schema.QueryFieldResolvers = (*queryImpl)(nil)

type eventFetcher interface {
	Find(ctx context.Context, entity, check string) (*types.Event, error)
}

//
// Implement QueryFieldResolvers
//

type queryImpl struct {
	store           store.Store
	eventController eventFetcher
	nodeResolver    *nodeResolver
}

func newQueryImpl(store store.Store, resolver *nodeResolver) *queryImpl {
	return &queryImpl{
		store:           store,
		eventController: actions.NewEventController(store, nil),
		nodeResolver:    resolver,
	}
}

// Viewer implements response to request for 'viewer' field.
func (r *queryImpl) Viewer(p graphql.ResolveParams) (interface{}, error) {
	return struct{}{}, nil
}

// Environment implements response to request for 'environment' field.
func (r *queryImpl) Environment(p schema.QueryEnvironmentFieldResolverParams) (interface{}, error) {
	env := types.Environment{
		Name:         p.Args.Environment,
		Organization: p.Args.Organization,
	}
	return &env, nil
}

// Event implements response to request for 'event' field.
func (r *queryImpl) Event(p schema.QueryEventFieldResolverParams) (interface{}, error) {
	ctx := types.SetContextFromResource(p.Context, p.Args.Ns)
	event, err := r.eventController.Find(ctx, p.Args.Entity, p.Args.Check)
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
