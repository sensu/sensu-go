package authorization

import (
	"context"
	"net/http"

	"github.com/sensu/sensu-go/types"
)

const (
	ContextRoleKey = "roles"
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

func UnauthorizedAccessToResource(w http.ResponseWriter) {
	http.Error(w, "Not authorized to access the requested resource", http.StatusUnauthorized)
}
