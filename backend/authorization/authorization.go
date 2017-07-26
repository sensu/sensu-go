package authorization

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/types"
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

// CanAccessResource will verify whether or not a user has permission to perform
// an action, for a resource, within an organization
func CanAccessResource(actor Actor, resource, action string) bool {
	// TODO (JP): flatten rules prior to storing in actor?
	for _, role := range actor.Roles {
		for _, rule := range role.Rules {
			if !matchesRuleType(rule, resource) {
				continue
			}
			if !matchesRuleOrganization(rule, actor.Organization) {
				continue
			}
			if hasPermission(rule, action) {
				return true
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"actor":    actor,
		"resource": resource,
		"action":   action,
	}).Debug("request to resource not allowed")

	return false
}

// UnauthorizedAccessToResource will return an HTTP error that specifies that a
// user does not have access to a requested action, for a resource, within an
// organization
func UnauthorizedAccessToResource(w http.ResponseWriter) {
	http.Error(w, "Not authorized to access the requested resource", http.StatusUnauthorized)
}
