package actions

import (
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

// OrganizationsController defines the fields required for this controller.
type OrganizationsController struct {
	Store  store.Store
	Policy authorization.OrganizationPolicy
}

// NewOrganizationsController returns new OrganizationsController
func NewOrganizationsController(store store.OrganizationStore) OrganizationsController {
	return OrganizationsController{
		Store:  store,
		Policy: authorization.Organizations,
	}
}

// Query returns resources available to the viewer filter by given params.
func (a OrganizationsController) Query(ctx context.Context, params QueryParams) ([]*types.Organization, error) {
	// Fetch from store
	results, serr := a.Store.GetOrganizations(ctx)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	// Filter out those resources the viewer does not have access to view.
	abilities := a.Policy.WithContext(ctx)
	for i := 0; i < len(results); i++ {
		if !abilities.CanRead(results[i]) {
			results = append(results[:i], results[i+1:]...)
			i--
		}
	}

	return results, nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (a OrganizationsController) Find(ctx context.Context, name string) (*types.Organization, error) {
	// Fetch from store
	result, serr := a.Store.GetOrganizationByName(ctx, name)
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
func (a OrganizationsController) Create(ctx context.Context, newOrg types.Organization) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &newOrg)
	abilities := a.Policy.WithContext(ctx)

	// Check for existing
	if e, err := a.Store.GetCheckConfigByName(ctx, newOrg.Name); err != nil {
		return NewError(InternalErr, err)
	} else if e != nil {
		return NewErrorf(AlreadyExistsErr)
	}

	// Verify viewer can make change
	if yes := abilities.CanCreate(&newOrg); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Validate
	if err := newOrg.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := a.Store.UpdateOrganization(ctx, &newOrg); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Note - removed addition of org to context and call to copyFields, make sure
// we don't need those
// Update validates and persists changes to a resource if viewer has access.
func (a OrganizationsController) Update(ctx context.Context, given types.Organization) error {
	abilities := a.Policy.WithContext(ctx)

	// Find existing organization
	org, err := a.Store.GetOrganizationByName(ctx, given.Name)
	if err != nil {
		return NewError(InternalErr, err)
	} else if org == nil {
		return NewErrorf(NotFound)
	}

	// Verify viewer can make change
	if yes := abilities.CanUpdate(org); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Validate
	if err := org.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist Changes
	if serr := a.Store.UpdateOrganization(ctx, org); serr != nil {
		return NewError(InternalErr, serr)
	}

	return nil
}

// Destroy removes a resource if viewer has access.
func (a OrganizationsController) Destroy(ctx context.Context, name string) error {
	abilities := a.Policy.WithContext(ctx)

	// Verify user has permission
	if yes := abilities.CanDelete(); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Fetch from store
	result, serr := a.Store.GetOrganizationByName(ctx, name)
	if serr != nil {
		return NewError(InternalErr, serr)
	} else if result == nil {
		return NewErrorf(NotFound)
	}

	// Remove from store
	if err := a.Store.DeleteOrganizationByName(ctx, result.Name); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}
