package useractions

import (
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

// EventController expose actions in which a viewer can perform.
type EventController struct {
	Store  store.EventStore
	Policy authorization.EventPolicy
}

// NewEventController returns new EventController
func NewEventController(store store.EventStore) EventController {
	return EventController{
		Store:  store,
		Policy: authorization.Events,
	}
}

// Query returns resources available to the viewer filter by given params.
func (a EventController) Query(ctx context.Context, params QueryParams) ([]interface{}, error) {
	var results []*types.Event

	entityID := params["entity"]
	checkName := params["check"]

	// Fetch from store
	var serr error
	if entityID != "" && checkName != "" {
		var result *types.Event
		result, serr = a.Store.GetEventByEntityCheck(ctx, entityID, checkName)
		if result != nil {
			results = append(results, result)
		}
	} else if entityID != "" {
		results, serr = a.Store.GetEventsByEntity(ctx, entityID)
	} else {
		results, serr = a.Store.GetEvents(ctx)
	}

	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	abilities := a.Policy.WithContext(ctx)

	// Filter out those resources the viewer does not have access to view.
	resources := []interface{}{}
	for _, event := range results {
		if yes := abilities.CanRead(event); yes {
			resources = append(resources, event)
		}
	}

	return resources, nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (a EventController) Find(ctx context.Context, params QueryParams) (interface{}, error) {
	// Find (for events) requires both an entity and check
	if params["entity"] == "" || params["check"] == "" {
		return nil, NewErrorf(InvalidArgument, "Find() requires both an entity and a check")
	}

	results, err := a.Query(ctx, params)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, NewErrorf(NotFound, "no event found")
	}

	return results[0], err
}
