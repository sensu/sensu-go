package api

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/store"
	stringsutil "github.com/sensu/sensu-go/util/strings"
	"github.com/sirupsen/logrus"
)

type ruleVisitor interface {
	VisitRulesFor(ctx context.Context, attrs *authorization.Attributes, fn rbac.RuleVisitFunc)
}

// NamespaceClient is an API client for namespaces.
type NamespaceClient struct {
	client GenericClient
	auth   authorization.Authorizer
}

// NewNamespaceClient creates a new NamespaceClient, given a store and authorizer.
func NewNamespaceClient(store store.ResourceStore, auth authorization.Authorizer) *NamespaceClient {
	return &NamespaceClient{
		client: GenericClient{
			Kind:       &corev2.Namespace{},
			Store:      store,
			Auth:       auth,
			APIGroup:   "core",
			APIVersion: "v2",
		},
		auth: auth,
	}
}

// ListNamespaces fetches a list of the namespace resources that are authorized
// by the supplied credentials. This may include implicit access via resources
// that are in a namespace that the credentials are authorized to get.
func (a *NamespaceClient) ListNamespaces(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Namespace, error) {
	var resources, namespaces []*corev2.Namespace

	visitor, ok := a.auth.(ruleVisitor)
	if !ok {
		if err := a.client.List(ctx, &namespaces, pred); err != nil {
			return nil, err
		}
		return namespaces, nil
	}
	if err := a.client.Store.ListResources(ctx, a.client.Kind.StorePrefix(), &resources, pred); err != nil {
		return nil, err
	}
	namespaceMap := make(map[string]*corev2.Namespace, len(resources))
	for _, namespace := range resources {
		namespaceMap[namespace.Name] = namespace
	}

	attrs := &authorization.Attributes{
		APIGroup:   a.client.APIGroup,
		APIVersion: a.client.APIVersion,
		Resource:   a.client.Kind.RBACName(),
		Namespace:  corev2.ContextNamespace(ctx),
		Verb:       "list",
	}
	if err := addAuthUser(ctx, attrs); err != nil {
		return nil, err
	}
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

	var funcErr error

	visitor.VisitRulesFor(ctx, attrs, func(binding rbac.RoleBinding, rule corev2.Rule, err error) (cont bool) {
		if err != nil {
			funcErr = err
			return false
		}
		if len(namespaceMap) == 0 {
			return false
		}
		if !rule.VerbMatches("get") {
			return true
		}

		// Explicit access to namespaces can only be granted via a
		// ClusterRoleBinding
		if rule.ResourceMatches(corev2.NamespacesResource) && binding.GetObjectMeta().Namespace == "" {
			// If this rule applies to namespaces, determine if all resources of type "namespace" are allowed
			if len(rule.ResourceNames) == 0 {
				// All resources of type "namespace" are allowed
				logger.Debugf("all namespaces explicitly authorized by the binding %s", binding.GetObjectMeta().Name)
				namespaces = resources
				return false
			}

			// If this rule applies to namespaces, and only certain namespaces are
			// specified, determine if it matches this current namespace
			for name, namespace := range namespaceMap {
				if rule.ResourceNameMatches(name) {
					logger.Debugf("namespace %s explicitly authorized by the binding %s", namespace.Name, binding.GetObjectMeta().Name)
					namespaces = append(namespaces, namespace)
					delete(namespaceMap, name)
				}
			}

			return true
		}

		// Determine if this ClusterRoleBinding provides implicit access to
		// namespaced resources
		if binding.GetObjectMeta().Namespace == "" {
			for _, resource := range rule.Resources {
				if stringsutil.InArray(resource, corev2.CommonCoreResources) {
					// All resources of type "namespace" are allowed
					logger.Debugf("all namespaces implicitly authorized by the binding %s", binding.GetObjectMeta().Name)
					namespaces = resources
					return false
				}
			}
		}

		// Determine if this RoleBinding matches the namespace
		bindingNamespace := binding.GetObjectMeta().Namespace
		if namespace, ok := namespaceMap[bindingNamespace]; ok {
			logger.Debugf("namespace %s implicitly authorized by the binding %s", namespace.Name, binding.GetObjectMeta().Name)
			namespaces = append(namespaces, namespace)
			delete(namespaceMap, bindingNamespace)
		}

		return true
	})

	if funcErr != nil {
		return nil, fmt.Errorf("error listing namespaces: %s", funcErr)
	}

	if len(namespaces) == 0 {
		logger.Debug("unauthorized request")
		return nil, authorization.ErrUnauthorized
	}

	return namespaces, nil
}

// FetchNamespace fetches a namespace resource from the backend, if authorized.
func (a *NamespaceClient) FetchNamespace(ctx context.Context, name string) (*corev2.Namespace, error) {
	var namespace corev2.Namespace
	visitor, ok := a.auth.(ruleVisitor)
	if !ok {
		if err := a.client.Get(ctx, name, &namespace); err != nil {
			return nil, err
		}
		return &namespace, nil
	}
	if err := a.client.Store.GetResource(ctx, name, &namespace); err != nil {
		return nil, err
	}

	attrs := &authorization.Attributes{
		APIGroup:     a.client.APIGroup,
		APIVersion:   a.client.APIVersion,
		Resource:     a.client.Kind.RBACName(),
		ResourceName: name,
		Namespace:    name,
		Verb:         "get",
	}
	if err := addAuthUser(ctx, attrs); err != nil {
		return nil, err
	}
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

	var funcErr error
	var authorized bool

	visitor.VisitRulesFor(ctx, attrs, func(binding rbac.RoleBinding, rule corev2.Rule, err error) (cont bool) {
		if err != nil {
			funcErr = err
			logger.WithError(err).Warning("could not retrieve the ClusterRoleBindings or RoleBindings")
			return false
		}

		// If the rule verb doesn't match "get", ignore this rule and continue
		if !rule.VerbMatches("get") {
			return true
		}

		// Explicit access to namespaces can only be granted via a
		// ClusterRoleBinding
		if rule.ResourceMatches(corev2.NamespacesResource) && binding.GetObjectMeta().Namespace == "" {
			// If this rule applies to namespaces, determine if all resources of type "namespace" are allowed
			if len(rule.ResourceNames) == 0 {
				logger.Debugf("request authorized by the binding %s", binding.GetObjectMeta().Name)
				authorized = true
				return false
			}

			// If this rule applies to namespaces, and only certain namespaces are
			// specified, determine if it matches this current namespace
			if rule.ResourceNameMatches(name) {
				logger.Debugf("request authorized by the binding %s", binding.GetObjectMeta().Name)
				authorized = true
				return false
			}
		}

		// Determine if this ClusterRoleBinding provides implicit access to
		// namespaced resources
		if binding.GetObjectMeta().Namespace == "" {
			for _, resource := range rule.Resources {
				if stringsutil.InArray(resource, corev2.CommonCoreResources) {
					logger.Debugf("request authorized by the binding %s", binding.GetObjectMeta().Name)
					authorized = true
					return false
				}
			}
		}

		// Determine if this RoleBinding matches the namespace
		if binding.GetObjectMeta().Namespace == name {
			logger.Debugf("request authorized by the binding %s", binding.GetObjectMeta().Name)
			authorized = true
			return false
		}

		return true
	})

	if funcErr != nil {
		return nil, fmt.Errorf("error getting namespace: %s", funcErr)
	}

	if !authorized {
		logger.Debug("unauthorized request")
		return nil, authorization.ErrUnauthorized
	}

	return &namespace, nil
}

// CreateNamespace creates a namespace resource, if authorized.
func (a *NamespaceClient) CreateNamespace(ctx context.Context, namespace *corev2.Namespace) error {
	if err := a.client.Create(ctx, namespace); err != nil {
		return err
	}
	return nil
}

// UpdateNamespace updates a namespace resource, if authorized.
func (a *NamespaceClient) UpdateNamespace(ctx context.Context, namespace *corev2.Namespace) error {
	if err := a.client.Update(ctx, namespace); err != nil {
		return err
	}
	return nil
}
