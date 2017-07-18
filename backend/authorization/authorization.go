package authorization

import (
	"context"
	"net/http"

	"github.com/sensu/sensu-go/types"
)

// Define the key type to avoid collisions in context
type key int

const (
	// ContextRoleKey is the key name used to store a user's roles within
	// the context of a request
	ContextRoleKey key = iota
)

func hasPermission(rule types.Rule, action string) bool {
	for _, permission := range rule.Permissions {
		if permission == action {
			return true
		}
	}
	return false
}

// TODO (JK): this function may end up becoming more complex if
// we decide to use "*" as more than a way of saying "all resources"
func matchesRuleType(rule types.Rule, resource string) bool {
	return rule.Type == resource || rule.Type == "*"
}

// TODO (JK): this function may end up becoming more complex if
// we decide to use "*" as more than a way of saying "all organizations"
func matchesRuleOrganization(rule types.Rule, organization string) bool {
	return rule.Organization == organization || rule.Organization == "*"
}

// ContextCanAccessResource will verify whether or not a user has permission to
// an action, for a resource, within an organization
func ContextCanAccessResource(ctx context.Context, resource, action string) bool {
	organization := ctx.Value(types.OrganizationKey).(string)
	roles := ctx.Value(ContextRoleKey).([]types.Role)
	for _, role := range roles {
		for _, rule := range role.Rules {
			if !matchesRuleType(rule, resource) {
				continue
			}
			if !matchesRuleOrganization(rule, organization) {
				continue
			}
			if hasPermission(rule, action) {
				return true
			}
		}
	}
	return false
}

// UnauthorizedAccessToResource will return an HTTP error that specifies that a
// user does not have access to a requested action, for a resource, within an
// organization
func UnauthorizedAccessToResource(w http.ResponseWriter) {
	http.Error(w, "Not authorized to access the requested resource", http.StatusUnauthorized)
}
