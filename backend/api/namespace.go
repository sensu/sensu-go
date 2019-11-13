package api

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/store"
)

type ruleVisitor interface {
	VisitRulesFor(ctx context.Context, attrs *authorization.Attributes, fn rbac.RuleVisitFunc)
}

// NamespaceClient is an API client for namespaces.
type NamespaceClient struct {
	client GenericClient
	auth   authorization.Authorizer
}

// NewnamespaceClient creates a new NamespaceClient, given a store and authorizer.
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
func (a *NamespaceClient) ListNamespaces(ctx context.Context) ([]*corev2.Namespace, error) {
	var resources, namespaces []*corev2.Namespace
	pred := &store.SelectionPredicate{
		Continue: corev2.PageContinueFromContext(ctx),
		Limit:    int64(corev2.PageSizeFromContext(ctx)),
	}
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
		if !rule.ResourceMatches(corev2.NamespacesResource) {
			// Find namespaces with implicit access
			ns := binding.GetObjectMeta().Namespace
			if namespace, ok := namespaceMap[ns]; ok {
				namespaces = append(namespaces, namespace)
				delete(namespaceMap, ns)
			}
			return true
		}
		if len(rule.ResourceNames) == 0 {
			// All resources of type "namespace" are allowed
			namespaces = resources
			return false
		}
		for name, namespace := range namespaceMap {
			if rule.ResourceNameMatches(name) {
				namespaces = append(namespaces, namespace)
				delete(namespaceMap, name)
			}
		}
		return true
	})

	if funcErr != nil {
		return nil, fmt.Errorf("error listing namespaces: %s", funcErr)
	}

	return namespaces, nil
}

// FetchNamespace fetches a namespace resource from the backend, if authorized.
func (a *NamespaceClient) FetchNamespace(ctx context.Context, name string) (*corev2.Namespace, error) {
	var namespace corev2.Namespace
	if err := a.client.Get(ctx, name, &namespace); err != nil {
		return nil, err
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
