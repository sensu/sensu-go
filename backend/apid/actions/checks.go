package actions

import (
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

// CheckMutator exposes actions in which a viewer can perform.
type CheckMutator interface {
	Create(context.Context, types.CheckConfig) error
	Update(context.Context, types.CheckConfig) error
}

// updateFields refers to fields a viewer may update
var updateFields = []string{
	"Command",
	"Handlers",
	"HighFlapThreshold",
	"LowFlapThreshold",
	"Interval",
	"Publish",
	"RuntimeAssets",
	"Subscriptions",
}

// CheckController exposes actions in which a viewer can perform.
type CheckController struct {
	Store  store.CheckConfigStore
	Policy authorization.CheckPolicy
}

// NewCheckController returns new CheckController
func NewCheckController(store store.CheckConfigStore) CheckController {
	return CheckController{
		Store:  store,
		Policy: authorization.Checks,
	}
}

// Query returns resources available to the viewer filter by given params.
func (a CheckController) Query(ctx context.Context, params QueryParams) ([]interface{}, error) {
	abilities := a.Policy.WithContext(ctx)
	if yes := abilities.CanList(); !yes {
		return nil, NewErrorf(PermissionDenied, "cannot list resources")
	}

	// Fetch from store
	results, serr := a.Store.GetCheckConfigs(ctx)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	// Filter out those resources the viewer does not have access to view.
	resources := []interface{}{}
	for _, result := range results {
		if yes := abilities.CanRead(result); yes {
			resources = append(resources, result)
		}
	}

	return resources, nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (a CheckController) Find(ctx context.Context, params QueryParams) (interface{}, error) {
	// Validate params
	if id := params["id"]; id == "" {
		return nil, NewErrorf(InternalErr, "'id' param missing")
	}

	// Fetch from store
	result, serr := a.Store.GetCheckConfigByName(ctx, params["id"])
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	// Verify user has permission to view
	abilities := a.Policy.WithContext(ctx)
	if result != nil && abilities.CanRead(result) {
		return result, nil
	}

	return nil, NewErrorf(NotFound)
}

// Create instatiates, validates and persists new resource if viewer has access.
func (a CheckController) Create(ctx context.Context, newCheck types.CheckConfig) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &newCheck)
	abilities := a.Policy.WithContext(ctx)

	// Check for existing
	if e, err := a.Store.GetCheckConfigByName(ctx, newCheck.Name); err != nil {
		return NewError(InternalErr, err)
	} else if e != nil {
		return NewErrorf(AlreadyExistsErr)
	}

	// Verify viewer can make change
	if yes := abilities.CanCreate(&newCheck); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Validate
	if err := newCheck.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := a.Store.UpdateCheckConfig(ctx, &newCheck); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Update validates and persists changes to a resource if viewer has access.
func (a CheckController) Update(ctx context.Context, given types.CheckConfig) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &given)
	abilities := a.Policy.WithContext(ctx)

	// Find existing check
	check, err := a.Store.GetCheckConfigByName(ctx, given.Name)
	if err != nil {
		return NewError(InternalErr, err)
	} else if check == nil {
		return NewErrorf(NotFound)
	}

	// Verify viewer can make change
	if yes := abilities.CanUpdate(check); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Copy
	copyFields(check, &given, updateFields...)

	// Validate
	if err := check.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist Changes
	if serr := a.Store.UpdateCheckConfig(ctx, check); serr != nil {
		return NewError(InternalErr, serr)
	}

	return nil
}

// Destroy removes a resource if viewer has access.
func (a CheckController) Destroy(ctx context.Context, params QueryParams) error {
	abilities := a.Policy.WithContext(ctx)

	// Verify user has permission
	if yes := abilities.CanDelete(); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Fetch from store
	result, serr := a.Store.GetCheckConfigByName(ctx, params["id"])
	if serr != nil {
		return NewError(InternalErr, serr)
	} else if result == nil {
		return NewErrorf(NotFound)
	}

	// Remove from store
	if err := a.Store.DeleteCheckConfigByName(ctx, result.Name); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}
