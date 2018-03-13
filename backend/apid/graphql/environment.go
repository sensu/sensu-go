package graphql

import (
	"sort"

	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/relay"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/eval"
)

var _ schema.EnvironmentFieldResolvers = (*envImpl)(nil)

//
// Implement EnvironmentFieldResolvers
//

type envImpl struct {
	orgCtrl    actions.OrganizationsController
	checksCtrl actions.CheckController
	entityCtrl actions.EntityController
	eventsCtrl actions.EventController
}

func newEnvImpl(store store.Store, getter types.QueueGetter) *envImpl {
	return &envImpl{
		orgCtrl:    actions.NewOrganizationsController(store),
		checksCtrl: actions.NewCheckController(store, getter),
		entityCtrl: actions.NewEntityController(store),
		eventsCtrl: actions.NewEventController(store, nil),
	}
}

// ID implements response to request for 'id' field.
func (r *envImpl) ID(p graphql.ResolveParams) (interface{}, error) {
	return globalid.EnvironmentTranslator.EncodeToString(p.Source), nil
}

// Name implements response to request for 'name' field.
func (r *envImpl) Name(p graphql.ResolveParams) (string, error) {
	org := p.Source.(*types.Environment)
	return org.Name, nil
}

// Description implements response to request for 'description' field.
func (r *envImpl) Description(p graphql.ResolveParams) (string, error) {
	org := p.Source.(*types.Environment)
	return org.Description, nil
}

// Organization implements response to request for 'organization' field.
func (r *envImpl) Organization(p graphql.ResolveParams) (interface{}, error) {
	env := p.Source.(*types.Environment)
	org, err := r.orgCtrl.Find(p.Context, env.Name)
	return handleControllerResults(org, err)
}

// Checks implements response to request for 'checks' field.
func (r *envImpl) Checks(p schema.EnvironmentChecksFieldResolverParams) (interface{}, error) {
	env := p.Source.(types.Environment)
	ctx := types.SetContextFromResource(p.Context, &env)
	records, err := r.checksCtrl.Query(ctx)
	if err != nil {
		return nil, err
	}

	// pagination
	info := relay.NewArrayConnectionInfo(
		0, len(records),
		p.Args.First, p.Args.Last, p.Args.Before, p.Args.After,
	)
	edges := make([]*relay.Edge, info.End-info.Begin)
	for i, r := range records[info.Begin:info.End] {
		edges[i] = relay.NewArrayConnectionEdge(r, i)
	}
	return relay.NewArrayConnection(edges, info), nil
}

// Entities implements response to request for 'entities' field.
func (r *envImpl) Entities(p schema.EnvironmentEntitiesFieldResolverParams) (interface{}, error) {
	env := p.Source.(types.Environment)
	ctx := types.SetContextFromResource(p.Context, &env)
	records, err := r.entityCtrl.Query(ctx)
	if err != nil {
		return nil, err
	}

	info := relay.NewArrayConnectionInfo(
		0, len(records),
		p.Args.First, p.Args.Last, p.Args.Before, p.Args.After,
	)

	edges := make([]*relay.Edge, info.End-info.Begin)
	for i, r := range records[info.Begin:info.End] {
		edges[i] = relay.NewArrayConnectionEdge(r, i)
	}
	return relay.NewArrayConnection(edges, info), nil
}

// Events implements response to request for 'events' field.
func (r *envImpl) Events(p schema.EnvironmentEventsFieldResolverParams) (interface{}, error) {
	env := p.Source.(types.Environment)
	ctx := types.SetContextFromResource(p.Context, &env)
	records, err := r.eventsCtrl.Query(ctx, "", "")
	if err != nil {
		return nil, err
	}

	// apply filters
	var filteredEvents []*types.Event
	filter := p.Args.Filter
	if len(filter) > 0 {
		predicate, err := eval.NewPredicate(filter)
		if err != nil {
			return nil, err
		}

		for _, event := range records {
			if matched, err := predicate.Eval(event); err != nil {
				return nil, err
			} else if matched {
				filteredEvents = append(filteredEvents, event)
			}
		}
	} else {
		filteredEvents = records
	}

	// sort records
	if p.Args.OrderBy == schema.EventsListOrders.SEVERITY {
		sort.Sort(types.EventsBySeverity(filteredEvents))
	} else {
		sort.Sort(types.EventsByTimestamp(
			filteredEvents,
			p.Args.OrderBy == schema.EventsListOrders.NEWEST,
		))
	}

	// pagination
	info := relay.NewArrayConnectionInfo(
		0, len(filteredEvents),
		p.Args.First, p.Args.Last, p.Args.Before, p.Args.After,
	)
	edges := make([]*relay.Edge, info.End-info.Begin)
	for i, r := range filteredEvents[info.Begin:info.End] {
		edges[i] = relay.NewArrayConnectionEdge(r, i)
	}
	return relay.NewArrayConnection(edges, info), nil
}
