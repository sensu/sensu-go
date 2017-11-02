package useractions

import (
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

// EventActions expose actions in which a viewer can perform.
type EventActions struct {
	Store   store.EventStore
	Policy  authorization.EventPolicy
	Context context.Context
}

// NewEventActions returns new EventActions
func NewEventActions(store store.EventStore) EventActions {
	return EventActions{
		Store:  store,
		Policy: authorization.Events,
	}
}

// WithContext returns new EventActions w/ context & policy configured.
func (a EventActions) WithContext(ctx context.Context) EventActions {
	return EventActions{
		Store:   a.Store,
		Policy:  a.Policy.WithContext(ctx),
		Context: ctx,
	}
}

// Query returns resources available to the viewer filter by given params.
func (a EventActions) Query(params QueryParams) ([]interface{}, error) {
	var results []*types.Event

	entityID := params["entity"]
	checkName := params["check"]

	// Fetch from store
	var serr error
	if entityID != "" && checkName != "" {
		var result *types.Event
		result, serr = a.Store.GetEventByEntityCheck(a.Context, entityID, checkName)
		if result != nil {
			results = append(results, result)
		}
	} else if entityID != "" {
		results, serr = a.Store.GetEventsByEntity(a.Context, entityID)
	} else {
		results, serr = a.Store.GetEvents(a.Context)
	}

	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	// Filter out those resources the viewer does not have access to view.
	resources := []interface{}{}
	for _, event := range results {
		if yes := a.Policy.CanRead(event); yes {
			resources = append(resources, event)
		}
	}

	return resources, nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (a EventActions) Find(params QueryParams) (interface{}, error) {
	entityID := params["entity"]
	checkName := params["check"]

	// Find (for events) requires both an entity and check
	if entityID == "" || checkName == "" {
		return nil, NewErrorf(InvalidArgument, "Find() requires both an entity and a check")
	}

	results, err := a.Query(params)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, NewErrorf(NotFound, "no event found")
	}

	return results[0], err
}
