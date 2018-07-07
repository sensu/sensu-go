package authorization

import (
	"net/http"

	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

func hasPermission(rule types.Rule, action string) bool {
	for _, permission := range rule.Permissions {
		if permission == action {
			return true
		}
	}
	return false
}

func matchesRuleType(rule types.Rule, resource string) bool {
	return rule.Type == resource || rule.Type == types.RuleTypeAll
}

func matchesRuleEnvironment(rule types.Rule, environment string) bool {
	return rule.Environment == environment || rule.Environment == types.EnvironmentTypeAll
}

func matchesRuleOrganization(rule types.Rule, organization string) bool {
	return rule.Organization == organization || rule.Organization == types.OrganizationTypeAll
}

// CanAccessResource will verify whether or not a user has permission to perform
// an action, for a resource, within an organization
func CanAccessResource(actor Actor, org, env, resource, action string) bool {
	// TODO: Reject irrelevant rules?
	for _, rule := range actor.Rules {
		if !matchesRuleType(rule, resource) {
			continue
		}
		if !matchesRuleOrganization(rule, org) {
			continue
		}
		if resource != types.RuleTypeAsset && resource != types.RuleTypeOrganization && !matchesRuleEnvironment(rule, env) {
			continue
		}
		if hasPermission(rule, action) {
			return true
		}
	}

	logrus.WithFields(logrus.Fields{
		"action":   action,
		"actor":    actor,
		"env":      env,
		"org":      org,
		"resource": resource,
	}).Info("request to resource not allowed")

	return false
}

// UnauthorizedAccessToResource will return an HTTP error that specifies that a
// user does not have access to a requested action, for a resource, within an
// organization
func UnauthorizedAccessToResource(w http.ResponseWriter) {
	http.Error(w, "Not authorized to access the requested resource", http.StatusUnauthorized)
}
