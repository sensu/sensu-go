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

// Context holds the namespace the action is associated with and the user
// making said action.
type Context struct {
	Actor     Actor
	Namespace string
}

// ExtractValueFromContext extracts authorization details from a context
func ExtractValueFromContext(ctx context.Context) Context {
	context := Context{}

	if namespace, ok := ctx.Value(types.NamespaceKey).(string); ok {
		context.Namespace = namespace
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
	return canPerformOn(
		policy,
		policy.Context().Namespace,
		action,
	)
}

func canPerformOn(policy Policy, namespace, action string) bool {
	return CanAccessResource(
		policy.Context().Actor,
		namespace,
		policy.Resource(),
		action,
	)
}
