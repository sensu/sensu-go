package api

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	stringsutil "github.com/sensu/sensu-go/util/strings"
	"github.com/sirupsen/logrus"
)

const pipelineRoleName = "system:pipeline"

type ruleVisitor interface {
	VisitRulesFor(ctx context.Context, attrs *authorization.Attributes, fn rbac.RuleVisitFunc)
}

// NamespaceClient is an API client for namespaces.
type NamespaceClient struct {
	client        GenericClient
	roleClient    GenericClient
	bindingClient GenericClient
	store         storev2.Interface
	auth          authorization.Authorizer
}

func namespaceRequest(ctx context.Context, name string) storev2.ResourceRequest {
	return storev2.NewResourceRequestFromResource(&corev3.Namespace{
		Metadata: &corev2.ObjectMeta{
			Name: name,
		},
	})
}

// NewNamespaceClient creates a new NamespaceClient, given a store and authorizer.
func NewNamespaceClient(store storev2.Interface, auth authorization.Authorizer) *NamespaceClient {
	return &NamespaceClient{
		client: GenericClient{
			Kind:       &corev3.Namespace{},
			Store:      store,
			Auth:       auth,
			APIGroup:   "core",
			APIVersion: "v3",
		},
		roleClient: GenericClient{
			Kind:       &corev2.Role{},
			Store:      store,
			Auth:       auth,
			APIGroup:   "core",
			APIVersion: "v2",
		},
		bindingClient: GenericClient{
			Kind:       &corev2.RoleBinding{},
			Store:      store,
			Auth:       auth,
			APIGroup:   "core",
			APIVersion: "v2",
		},
		store: store,
		auth:  auth,
	}
}

// ListNamespaces fetches a list of the namespace resources that are authorized
// by the supplied credentials. This may include implicit access via resources
// that are in a namespace that the credentials are authorized to get.
func (a *NamespaceClient) ListNamespaces(ctx context.Context, pred *store.SelectionPredicate) ([]*corev3.Namespace, error) {
	var resources, namespaces []*corev3.Namespace

	visitor, ok := a.auth.(ruleVisitor)
	if !ok {
		if err := a.client.List(ctx, &namespaces, pred); err != nil {
			return nil, err
		}
		return namespaces, nil
	}
	req := namespaceRequest(ctx, "")
	list, err := a.client.Store.List(ctx, req, pred)
	if err != nil {
		return nil, err
	}
	if err := list.UnwrapInto(&resources); err != nil {
		return nil, err
	}
	namespaceMap := make(map[string]*corev3.Namespace, len(resources))
	for _, namespace := range resources {
		namespaceMap[namespace.Metadata.Name] = namespace
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
					logger.Debugf("namespace %s explicitly authorized by the binding %s", namespace.Metadata.Name, binding.GetObjectMeta().Name)
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
			logger.Debugf("namespace %s implicitly authorized by the binding %s", namespace.Metadata.Name, binding.GetObjectMeta().Name)
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
func (a *NamespaceClient) FetchNamespace(ctx context.Context, name string) (*corev3.Namespace, error) {
	visitor, ok := a.auth.(ruleVisitor)
	if !ok {
		var ns corev3.Namespace
		if err := a.client.Get(ctx, name, &ns); err != nil {
			return nil, err
		}
		return &ns, nil
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

	req := namespaceRequest(ctx, name)
	wrapper, err := a.store.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	var ns corev3.Namespace
	if err := wrapper.UnwrapInto(&ns); err != nil {
		return nil, err
	}

	return &ns, nil
}

func (a *NamespaceClient) createRoleAndBinding(ctx context.Context, namespace string) error {
	role := &corev2.Role{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: namespace,
			Name:      pipelineRoleName,
		},
		Rules: []corev2.Rule{
			{
				Verbs:     []string{"get", "list"},
				Resources: []string{new(corev2.Event).RBACName()},
			},
		},
	}
	binding := &corev2.RoleBinding{
		Subjects: []corev2.Subject{
			{
				Type: corev2.GroupType,
				Name: pipelineRoleName,
			},
		},
		RoleRef: corev2.RoleRef{
			Name: pipelineRoleName,
			Type: "Role",
		},
		ObjectMeta: corev2.ObjectMeta{
			Name:      pipelineRoleName,
			Namespace: namespace,
		},
	}
	if err := a.roleClient.Update(ctx, role); err != nil {
		return err
	}
	return a.bindingClient.Update(ctx, binding)
}

// CreateNamespace creates a namespace resource, if authorized.
func (a *NamespaceClient) CreateNamespace(ctx context.Context, namespace *corev3.Namespace) error {
	if err := a.client.Create(ctx, namespace); err != nil {
		return err
	}
	if err := a.createResourceTemplates(ctx, namespace.Metadata.Name); err != nil {
		return err
	}
	return a.createRoleAndBinding(ctx, namespace.Metadata.Name)
}

func (a *NamespaceClient) createResourceTemplates(ctx context.Context, namespace string) error {
	req := storev2.NewResourceRequestFromResource(new(corev3.ResourceTemplate))
	list, err := a.store.List(ctx, req, nil)
	if err != nil {
		return err
	}
	var templates []*corev3.ResourceTemplate
	if err := list.UnwrapInto(&templates); err != nil {
		return err
	}
	for _, template := range templates {
		meta := &corev2.ObjectMeta{
			Namespace: namespace,
		}
		resource, err := template.Execute(meta)
		if err != nil {
			return err
		}
		req := storev2.NewResourceRequestFromResource(resource)
		wrapper, err := storev2.WrapResource(resource)
		if err != nil {
			return err
		}
		if err := a.store.CreateOrUpdate(ctx, req, wrapper); err != nil {
			return err
		}
	}
	return nil
}

// UpdateNamespace updates a namespace resource, if authorized.
func (a *NamespaceClient) UpdateNamespace(ctx context.Context, namespace *corev3.Namespace) error {
	if err := a.client.Update(ctx, namespace); err != nil {
		return err
	}
	if err := a.createResourceTemplates(ctx, namespace.Metadata.Name); err != nil {
		return err
	}
	return a.createRoleAndBinding(ctx, namespace.Metadata.Name)
}

// DeleteNamespace deletes a namespace.
func (a *NamespaceClient) DeleteNamespace(ctx context.Context, name string) error {
	// Inject the namespace into the context so we can target the namespaced
	// resources
	namespacedCtx := context.WithValue(ctx, corev2.NamespaceKey, name)

	// The generic client takes care of authorization for us, so if we
	// bypass it as we're doing here, we must not forget to deal with
	// authorization ourselves.
	attrs := namespaceDeleteAttributes(ctx, name)
	if err := authorize(ctx, a.auth, attrs); err != nil {
		return err
	}

	// We don't use the generic client and store here because there is some
	// special logic that applies to namespace deletion, namely the fact that we
	// don't want to delete namespace objects as if they were independent
	// objects: we want to make sure that a namespace is logically "empty"
	// before we remove it for good.
	req := storev2.NewResourceRequestFromResource(&corev3.Namespace{Metadata: &corev2.ObjectMeta{Name: name}})
	if err := a.store.Delete(ctx, req); err != nil {
		return err
	}

	if err := a.roleClient.Delete(namespacedCtx, pipelineRoleName); err != nil {
		logger.Warnf("could not delete implicit %s role in namespace %s: %s", pipelineRoleName, name, err)
	}

	if err := a.bindingClient.Delete(namespacedCtx, pipelineRoleName); err != nil {
		logger.Warnf("could not delete implicit %s binding in namespace %s: %s", pipelineRoleName, name, err)
	}

	return nil
}

func namespaceDeleteAttributes(ctx context.Context, name string) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:     "core",
		APIVersion:   "v2",
		Resource:     "namespaces",
		Verb:         "delete",
		ResourceName: name,
	}
}
