package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Actor describes an entity who can perform actions within the system that are
// bound by access controls.
type Actor struct {
	Name  string
	Rules []types.Rule
}

// Context holds the organization the action is associated with and the user
// making said action.
type Context struct {
	Actor        Actor
	Organization string
}

// ExtractValueFromContext extracts authorization details from a context
func ExtractValueFromContext(ctx context.Context) Context {
	context := Context{}

	if organization, ok := ctx.Value(types.OrganizationKey).(string); ok {
		context.Organization = organization
	}

	if actor, ok := ctx.Value(types.AuthorizationActorKey).(Actor); ok {
		context.Actor = actor
	}

	return context
}

// Policy ...
type Policy interface { // TODO: rename to ...?
	Resource() string
	Context() Context
}

func canPerform(policy Policy, action string) bool {
	return canPerformOn(policy, policy.Context().Organization, action)
}

func canPerformOn(policy Policy, organization, action string) bool {
	return CanAccessResource(
		policy.Context().Actor,
		organization,
		policy.Resource(),
		action,
	)
}
