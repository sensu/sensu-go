package rbac

import (
	"context"
	"fmt"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// Authorizer implements an authorizer interface using Role-Based Acccess
// Control (RBAC)
type Authorizer struct {
	Store store.Store
}

// RoleBinding implements the RoleBinding interface.
type RoleBinding interface {
	GetSubjects() []corev2.Subject
	GetRoleRef() corev2.RoleRef
	GetObjectMeta() corev2.ObjectMeta
}

// RuleVisitFunc is a function to help visit matching rules.
type RuleVisitFunc func(RoleBinding, corev2.Rule, error) (terminate bool)

// VisitRulesFor visits all of the matching rules for a given Attributes.
// It applies a visitor function that can elect to either continue visiting
// rules, or stop visiting rules.
//
// It is up to the visitor function to make a useful decision about the
// information it is given. For an example, see the Authorize method.
func (a *Authorizer) VisitRulesFor(ctx context.Context, attrs *authorization.Attributes, visitor RuleVisitFunc) {
	var empty = corev2.Rule{}
	clusterRoleBindings, err := a.Store.ListClusterRoleBindings(ctx, &store.SelectionPredicate{})
	if err != nil {
		if !visitor(nil, empty, err) {
			return
		}
	}
	for _, binding := range clusterRoleBindings {
		// Verify if this cluster role binding matches our user
		if !matchesUser(attrs.User, binding.Subjects) {
			continue
		}

		// Get the RoleRef that matched our user
		rules, err := a.getRoleReferencerules(ctx, binding.RoleRef)
		if err != nil {
			if !visitor(binding, empty, err) {
				return
			}
		}
		for _, rule := range rules {
			if !visitor(binding, rule, nil) {
				return
			}
		}
	}

	if len(attrs.Namespace) == 0 {
		return
	}

	roleBindings, err := a.Store.ListRoleBindings(ctx, &store.SelectionPredicate{})
	if err != nil {
		if !visitor(nil, empty, err) {
			return
		}
	}

	for _, binding := range roleBindings {
		// Verify if this role binding matches our user
		if !matchesUser(attrs.User, binding.Subjects) {
			continue
		}

		// Get the RoleRef that matched our user
		rules, err := a.getRoleReferencerules(ctx, binding.RoleRef)
		if err != nil {
			if !visitor(nil, empty, err) {
				return
			}
		}

		// Visit the rules
		for _, rule := range rules {
			if !visitor(binding, rule, nil) {
				return
			}
		}
	}
}

// Authorize determines if a request is authorized based on its attributes
func (a *Authorizer) Authorize(ctx context.Context, attrs *authorization.Attributes) (bool, error) {
	if attrs != nil {
		logger = logger.WithFields(logrus.Fields{
			"zz_request": map[string]string{
				"apiGroup":     attrs.APIGroup,
				"apiVersion":   attrs.APIVersion,
				"namespace":    attrs.Namespace,
				"resource":     attrs.Resource,
				"resourceName": attrs.ResourceName,
				"username":     attrs.User.Username,
				"verb":         attrs.Verb,
			},
		})
	}

	var (
		authorized bool
		visitErr   error
	)

	a.VisitRulesFor(ctx, attrs, func(binding RoleBinding, rule corev2.Rule, err error) bool {
		if err != nil {
			switch err := err.(type) {
			case *store.ErrNotFound:
				// No ClusterRoleBindings founds, let's continue with the RoleBindings
				logger.WithError(err).Debug("no bindings found")
			default:
				logger.WithError(err).Warning("could not retrieve the ClusterRoleBindings or RoleBindings")
				visitErr = err
				return false
			}
		}

		allowed, reason := ruleAllows(attrs, rule)
		if allowed {
			roleRef := binding.GetRoleRef()
			name := roleRef.GetName()
			logger.Debugf("request authorized by the binding %s", name)
			authorized = true
			return false
		}
		logger.Tracef("%s by rule %+v", reason, rule)

		return true
	})

	if !authorized {
		logger.Debugf("unauthorized request")
	}

	return authorized, visitErr
}

func (a *Authorizer) getRoleReferencerules(ctx context.Context, roleRef types.RoleRef) ([]types.Rule, error) {
	switch roleRef.Type {
	case "Role":
		role, err := a.Store.GetRole(ctx, roleRef.Name)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve the Role %s: %s", roleRef.Name, err.Error())
		} else if role == nil {
			return nil, fmt.Errorf("the Role %s is invalid", roleRef.Name)
		}
		return role.Rules, nil

	case "ClusterRole":
		clusterRole, err := a.Store.GetClusterRole(ctx, roleRef.Name)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve the ClusterRole %s: %s", roleRef.Name, err.Error())
		} else if clusterRole == nil {
			return nil, fmt.Errorf("the ClusterRole %s is invalid", roleRef.Name)
		}
		return clusterRole.Rules, nil

	default:
		return nil, fmt.Errorf("unsupported role reference type: %s", roleRef.Type)
	}
}

// matchesUser returns whether any of the subjects matches the specified user
func matchesUser(user types.User, subjects []types.Subject) bool {
	for _, subject := range subjects {
		switch subject.Type {
		case types.UserType:
			if user.Username == subject.Name {
				return true
			}

		case types.GroupType:
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
// attributes and if not, the reason why
func ruleAllows(attrs *authorization.Attributes, rule types.Rule) (bool, string) {
	if matches := rule.VerbMatches(attrs.Verb); !matches {
		return false, "forbidden verb"
	}

	if matches := rule.ResourceMatches(attrs.Resource); !matches {
		return false, "forbidden resource"
	}

	if matches := rule.ResourceNameMatches(attrs.ResourceName); !matches {
		return false, "forbidden resource name"
	}

	return true, ""
}
