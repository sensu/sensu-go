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

func matchesRuleNamespace(rule types.Rule, namespace string) bool {
	return rule.Namespace == namespace || rule.Namespace == types.NamespaceTypeAll
}

// isReadingRuleNamespace returns true if the user tries to read an
// namespace that is specified in the rule
func isReadingRuleNamespace(rule types.Rule, action, resource, namespace string) bool {
	if action == types.RulePermRead && resource == types.RuleTypeNamespace && namespace == rule.Namespace {
		return true
	}
	return false
}

// CanAccessResource will verify whether or not a user has permission to perform
// an action, for a resource, within an namespace
func CanAccessResource(actor Actor, namespace, resource, action string) bool {
	// TODO: Reject irrelevant rules?
	for _, rule := range actor.Rules {
		// Verify if the user is trying to read  anamespace that would implicitly be
		// granted if one its rules belongs to this namespace
		if isReadingRuleNamespace(rule, action, resource, namespace) {
			return true
		}
		if !MatchesRuleType(rule, resource) {
			continue
		}
		if !matchesRuleNamespace(rule, namespace) {
			continue
		}
		if HasPermission(rule, action) {
			return true
		}
	}

	logrus.WithFields(logrus.Fields{
		"action":    action,
		"actor":     actor,
		"namespace": namespace,
		"resource":  resource,
	}).Info("request to resource not allowed")

	return false
}

// UnauthorizedAccessToResource will return an HTTP error that specifies that a
// user does not have access to a requested action, for a resource, within an
// namespace
func UnauthorizedAccessToResource(w http.ResponseWriter) {
	http.Error(w, "Not authorized to access the requested resource", http.StatusUnauthorized)
}
