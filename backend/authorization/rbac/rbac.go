package rbac

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sirupsen/logrus"
)

type ErrRoleNotFound struct {
	Role    string
	Cluster bool
}

func (e ErrRoleNotFound) Type() string {
	if e.Cluster {
		return "cluster role"
	}
	return "role"
}

func (e ErrRoleNotFound) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Type(), e.Role)
}

// Authorizer implements an authorizer interface using Role-Based Acccess
// Control (RBAC)
type Authorizer struct {
	Store storev2.Interface
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
	namespace := corev2.ContextNamespace(ctx)
	var empty = corev2.Rule{}
	crbStore := storev2.Of[*corev2.ClusterRoleBinding](a.Store)
	clusterRoleBindings, err := crbStore.List(ctx, storev2.ID{}, nil)
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
		rules, err := a.getRoleReferenceRules(ctx, namespace, binding.RoleRef)
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

	if attrs.Namespace == "" && attrs.Resource != (&corev2.Namespace{}).RBACName() {
		return
	}

	rbStore := storev2.Of[*corev2.RoleBinding](a.Store)
	roleBindings, err := rbStore.List(ctx, storev2.ID{Namespace: namespace}, nil)
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

		ctx = store.NamespaceContext(ctx, binding.Namespace)

		// Get the RoleRef that matched our user
		rules, err := a.getRoleReferenceRules(ctx, binding.Namespace, binding.RoleRef)
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
			case ErrRoleNotFound:
				// The role binding specified a role that does not exist
				logger.WithError(err).Error("rbac configuration error")
				visitErr = err
				return false
			default:
				if ctx.Err() == nil {
					logger.WithError(err).Warning("could not retrieve the ClusterRoleBindings or RoleBindings")
				}
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
		logger.Debug("unauthorized request")
	}

	return authorized, visitErr
}

func (a *Authorizer) getRoleReferenceRules(ctx context.Context, namespace string, roleRef corev2.RoleRef) ([]corev2.Rule, error) {
	switch roleRef.Type {
	case "Role":
		rStore := storev2.Of[*corev2.Role](a.Store)

		role, err := rStore.Get(ctx, storev2.ID{Namespace: namespace, Name: roleRef.Name})
		if _, ok := err.(*store.ErrNotFound); ok {
			return nil, ErrRoleNotFound{Role: roleRef.Name, Cluster: false}
		} else if err != nil {
			return nil, fmt.Errorf("could not retrieve the role %s: %s", roleRef.Name, err)
		}
		return role.Rules, nil

	case "ClusterRole":
		crStore := storev2.Of[*corev2.ClusterRole](a.Store)
		clusterRole, err := crStore.Get(ctx, storev2.ID{Namespace: "", Name: roleRef.Name})
		if _, ok := err.(*store.ErrNotFound); ok {
			return nil, ErrRoleNotFound{Role: roleRef.Name, Cluster: true}
		} else if err != nil {
			return nil, fmt.Errorf("could not retrieve the ClusterRole %s: %s", roleRef.Name, err.Error())
		}
		return clusterRole.Rules, nil

	default:
		return nil, fmt.Errorf("unsupported role reference type: %s", roleRef.Type)
	}
}

// matchesUser returns whether any of the subjects matches the specified user
func matchesUser(user corev2.User, subjects []corev2.Subject) bool {
	for _, subject := range subjects {
		switch subject.Type {
		case corev2.UserType:
			if user.Username == subject.Name {
				return true
			}

		case corev2.GroupType:
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
func ruleAllows(attrs *authorization.Attributes, rule corev2.Rule) (bool, string) {
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
