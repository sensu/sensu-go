package authorization

import (
	"net/http"

	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

// HasPermission returns true if permission is granted on the action
func HasPermission(rule types.Rule, action string) bool {
	for _, permission := range rule.Permissions {
		if permission == action {
			return true
		}
	}
	return false
}

// MatchesRuleType returns true if the rule type matches the resource
func MatchesRuleType(rule types.Rule, resource string) bool {
	return rule.Type == resource || rule.Type == types.RuleTypeAll
}

func matchesRuleEnvironment(rule types.Rule, environment string) bool {
	return rule.Environment == environment || rule.Environment == types.EnvironmentTypeAll
}

func matchesRuleOrganization(rule types.Rule, organization string) bool {
	return rule.Organization == organization || rule.Organization == types.OrganizationTypeAll
}

// isReadingRuleEnvironment returns true if the user tries to read an
// environment that is specified in the rule
func isReadingRuleEnvironment(rule types.Rule, action, resource, environment string) bool {
	if action == types.RulePermRead && resource == types.RuleTypeEnvironment && environment == rule.Environment {
		return true
	}
	return false
}

// isReadingRuleOrganization returns true if the user tries to read an
// organization that is specified in the rule
func isReadingRuleOrganization(rule types.Rule, action, resource, organization string) bool {
	if action == types.RulePermRead && resource == types.RuleTypeOrganization && organization == rule.Organization {
		return true
	}
	return false
}

// CanAccessResource will verify whether or not a user has permission to perform
// an action, for a resource, within an organization
func CanAccessResource(actor Actor, org, env, resource, action string) bool {
	// TODO: Reject irrelevant rules?
	for _, rule := range actor.Rules {
		// Verify if the user is trying to read an environment or an organization
		// that would implicitly be granted if one its rules belongs to this
		// environment or organization
		if isReadingRuleEnvironment(rule, action, resource, env) || isReadingRuleOrganization(rule, action, resource, org) {
			return true
		}
		if !MatchesRuleType(rule, resource) {
			continue
		}
		if !matchesRuleOrganization(rule, org) {
			continue
		}
		if resource != types.RuleTypeAsset && resource != types.RuleTypeOrganization && !matchesRuleEnvironment(rule, env) {
			continue
		}
		if HasPermission(rule, action) {
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
