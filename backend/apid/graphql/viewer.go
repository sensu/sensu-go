package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/relay"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/eval"
)

var _ schema.ViewerFieldResolvers = (*viewerImpl)(nil)

//
// Implement QueryFieldResolvers
//

type viewerImpl struct {
	checksCtrl actions.CheckController
	entityCtrl actions.EntityController
	eventsCtrl actions.EventController
	usersCtrl  actions.UserController
	orgsCtrl   actions.OrganizationsController
}

func newViewerImpl(store QueueStore, bus messaging.MessageBus) *viewerImpl {
	return &viewerImpl{
		checksCtrl: actions.NewCheckController(store),
		entityCtrl: actions.NewEntityController(store),
		eventsCtrl: actions.NewEventController(store, bus),
		usersCtrl:  actions.NewUserController(store),
		orgsCtrl:   actions.NewOrganizationsController(store),
	}
}

// Entities implements response to request for 'entities' field.
func (r *viewerImpl) Entities(p schema.ViewerEntitiesFieldResolverParams) (interface{}, error) {
	records, err := r.entityCtrl.Query(p.Context)
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

// Checks implements response to request for 'checks' field.
func (r *viewerImpl) Checks(p schema.ViewerChecksFieldResolverParams) (interface{}, error) {
	records, err := r.checksCtrl.Query(p.Context)
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
func (r *viewerImpl) Events(p schema.ViewerEventsFieldResolverParams) (interface{}, error) {
	records, err := r.eventsCtrl.Query(p.Context, "", "")
	if err != nil {
		return nil, err
	}

	var filteredEvents []*types.Event
	filter := p.Args.Filter
	if len(filter) > 0 {
		for _, event := range records {
			args := map[string]interface{}{"event": event}
			if matched, err := eval.Evaluate(filter, args); err != nil {
				return nil, err
			} else if matched {
				filteredEvents = append(filteredEvents, event)
			}
		}
	} else {
		filteredEvents = records
	}

	info := relay.NewArrayConnectionInfo(
		0, len(filteredEvents),
		p.Args.First, p.Args.Last, p.Args.Before, p.Args.After,
	)

	edges := make([]*relay.Edge, len(filteredEvents))
	for i, r := range filteredEvents[info.Begin:info.End] {
		edges[i] = relay.NewArrayConnectionEdge(r, i)
	}
	return relay.NewArrayConnection(edges, info), nil
}

// Organizations implements response to request for 'organizations' field.
func (r *viewerImpl) Organizations(p graphql.ResolveParams) (interface{}, error) {
	return r.orgsCtrl.Query(p.Context)
}

// User implements response to request for 'user' field.
func (r *viewerImpl) User(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	actor := ctx.Value(types.AuthorizationActorKey).(authorization.Actor)
	return r.usersCtrl.Find(ctx, actor.Name)
}
