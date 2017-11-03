package useractions

import (
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

// CheckMutator exposes actions in which a viewer can perform.
type CheckMutator interface {
	Create(types.CheckConfig) error
	Update(types.CheckConfig) error
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

// CheckActions exposes actions in which a viewer can perform.
type CheckActions struct {
	Store   store.CheckConfigStore
	Policy  authorization.CheckPolicy
	Context context.Context
}

// NewCheckActions returns new CheckActions
func NewCheckActions(ctx context.Context, store store.CheckConfigStore) CheckActions {
	return CheckActions{Store: store}.WithContext(ctx)
}

// WithContext returns new CheckActions w/ context & policy configured.
func (a CheckActions) WithContext(ctx context.Context) CheckActions {
	if ctx != nil {
		a.Policy = a.Policy.WithContext(ctx)
		a.Context = ctx
	}
	return a
}

// Query returns resources available to the viewer filter by given params.
func (a CheckActions) Query(params QueryParams) ([]interface{}, error) {
	if yes := a.Policy.CanList(); !yes {
		return nil, NewErrorf(PermissionDenied, "cannot list resources")
	}

	// Fetch from store
	results, serr := a.Store.GetCheckConfigs(a.Context)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	// Filter out those resources the viewer does not have access to view.
	resources := []interface{}{}
	for _, result := range results {
		if yes := a.Policy.CanRead(result); yes {
			resources = append(resources, result)
		}
	}

	return resources, nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (a CheckActions) Find(params QueryParams) (interface{}, error) {
	// Validate params
	if id := params["id"]; id == "" {
		return nil, NewErrorf(InternalErr, "'id' param missing")
	}

	// Fetch from store
	result, serr := a.Store.GetCheckConfigByName(a.Context, params["id"])
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	// Verify user has permission to view
	if result != nil && a.Policy.CanRead(result) {
		return result, nil
	}

	return nil, NewErrorf(NotFound, "not found")
}

// Create instatiates, validates and persists new resource if viewer has access.
func (a CheckActions) Create(newCheck types.CheckConfig) error {
	// Adjust context
	ctx := addOrgEnvToContext(a.Context, &newCheck)

	// Check for existing
	if e, err := a.Store.GetCheckConfigByName(ctx, newCheck.Name); err != nil {
		return NewError(InternalErr, err)
	} else if e != nil {
		return NewErrorf(AlreadyExistsErr, "already exists")
	}

	// Verify viewer can make change
	if a.Policy.CanCreate(&newCheck) {
		return NewErrorf(PermissionDenied, "denied")
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
func (a CheckActions) Update(given types.CheckConfig) error {
	// Adjust context
	ctx := addOrgEnvToContext(a.Context, &given)

	// Find existing check
	check, err := a.Store.GetCheckConfigByName(ctx, given.Name)
	if err != nil {
		return NewError(InternalErr, err)
	} else if check == nil {
		return NewErrorf(NotFound, "not found")
	}

	// Verify viewer can make change
	if a.Policy.CanUpdate(check) {
		return NewErrorf(PermissionDenied, "denied")
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
func (a CheckActions) Destroy(params QueryParams) error {
	// Verify user has permission
	if a.Policy.CanDelete() {
		return NewErrorf(PermissionDenied, "denied")
	}

	// Fetch from store
	result, serr := a.Store.GetCheckConfigByName(a.Context, params["id"])
	if serr != nil {
		return NewError(InternalErr, serr)
	} else if result == nil {
		return NewErrorf(NotFound, "not found")
	}

	// Remove from store
	if err := a.Store.DeleteCheckConfigByName(a.Context, result.Name); err != nil {
		return NewError(InternalErr, serr)
	}

	return nil
}
