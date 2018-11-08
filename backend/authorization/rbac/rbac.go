package rbac

import (
	"context"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// Authorizer implements an authorizer interface using Role-Based Acccess
// Control (RBAC)
type Authorizer struct {
	Store store.Store
}

// Authorize determines if a request is authorized based on its attributes
func (a *Authorizer) Authorize(attrs *authorization.Attributes) (bool, error) {
	ctx := context.Background()

	// Get cluster roles binding
	clusterRoleBindings, err := a.Store.ListClusterRoleBindings(ctx)
	if err != nil {
		return false, err
	}

	// Inspect each cluster role binding
	for _, clusterRoleBinding := range clusterRoleBindings {
		// Verify if this cluster role binding matches our user
		if !matchesUser(attrs.User, clusterRoleBinding.Subjects) {
			continue
		}

		// Get the cluster role that matched our user
		clusterRole, err := a.Store.GetClusterRole(ctx, clusterRoleBinding.RoleRef.Name)
		if err != nil {
			return false, err
		} else if clusterRole == nil {
			continue
		}

		// Loop through the cluster role rules
		for _, rule := range clusterRole.Rules {
			// Verify if this rule applies to our request
			if ruleAllows(attrs, rule) {
				return true, nil
			}
		}
	}

	// None of the cluster roles authorized our request. Let's try with roles
	// First, make sure we have a namespace
	if len(attrs.Namespace) > 0 {
		// Get roles binding
		roleBindings, err := a.Store.ListRoleBindings(ctx)
		if err != nil {
			return false, err
		}

		// Inspect each role binding
		for _, roleBinding := range roleBindings {
			// Verify if this role binding matches our user
			if !matchesUser(attrs.User, roleBinding.Subjects) {
				continue
			}
			// Get the role that matched our user
			role, err := a.Store.GetRole(ctx, roleBinding.RoleRef.Name)
			if err != nil {
				return false, err
			} else if role == nil {
				continue
			}

			// Loop through the role rules
			for _, rule := range role.Rules {
				// Verify if this rule applies to our request
				if ruleAllows(attrs, rule) {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// matchesUser returns whether any of the subjects matches the specified user
func matchesUser(user types.User, subjects []types.Subject) bool {
	for _, subject := range subjects {
		switch subject.Kind {
		case types.UserKind:
			if user.Username == subject.Name {
				return true
			}

		case types.GroupKind:
			for _, group := range user.Groups {
				if group == subject.Name {
					return true
				}
			}
		}
	}

	return false
}

// ruleAllows returns whether the specified rule allows the request based on its
// attributes
func ruleAllows(attrs *authorization.Attributes, rule types.Rule) bool {
	return rule.VerbMatches(attrs.Verb) &&
		rule.ResourceMatches(attrs.Resource) &&
		rule.ResourceNameMatches(attrs.ResourceName)
}
