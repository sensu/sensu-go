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
	eventsCtrl actions.EventController
}

func newEnvImpl(store store.Store) *envImpl {
	return &envImpl{
		orgCtrl:    actions.NewOrganizationsController(store),
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

// Events implements response to request for 'events' field.
func (r *envImpl) Events(p schema.EnvironmentEventsFieldResolverParams) (interface{}, error) {
	env := p.Source.(types.Environment)
	ctx := types.SetContextFromResource(p.Context, &env)
	records, err := r.eventsCtrl.Query(ctx, "", "")
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

	if p.Args.OrderBy == schema.EventsListOrders.SEVERITY {
		sort.Slice(filteredEvents, func(i, j int) bool {
			first, second := filteredEvents[i], filteredEvents[j]

			// Sort events with the same exit status by timestamp
			if first.Check.Status == second.Check.Status {
				return first.Timestamp > second.Timestamp
			}

			// We want the order of importance to go 1, 2, 3, 0 so we shift status by one.
			// (Critical, Warning, Unknown, & OK.)
			firstStatus := (first.Check.Status + 3) % 4
			secondStatus := (second.Check.Status + 3) % 4
			return firstStatus < secondStatus
		})
	} else {
		sort.Slice(filteredEvents, func(i, j int) bool {
			first, second := filteredEvents[i], filteredEvents[j]

			if p.Args.OrderBy == schema.EventsListOrders.NEWEST {
				return first.Timestamp > second.Timestamp
			}
			return first.Timestamp < second.Timestamp
		})

	}

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
