package api

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
)

type NamespaceClient struct {
	client genericClient
	auth   authorization.Authorizer
}

func NewNamespaceClient(store store.ResourceStore, auth authorization.Authorizer) *NamespaceClient {
	return &NamespaceClient{
		client: genericClient{
			Kind:       &corev2.Namespace{},
			Store:      store,
			Auth:       auth,
			Resource:   "namespaces",
			APIGroup:   "core",
			APIVersion: "v2",
		},
		auth: auth,
	}
}

// ListNamespaces fetches a list of namespace resources
func (a *NamespaceClient) ListNamespaces(ctx context.Context) ([]*corev2.Namespace, error) {
	pred := &store.SelectionPredicate{
		Continue: corev2.PageContinueFromContext(ctx),
		Limit:    int64(corev2.PageSizeFromContext(ctx)),
	}
	slice := []*corev2.Namespace{}
	if err := a.client.List(ctx, &slice, pred); err != nil {
		return nil, fmt.Errorf("couldn't list namespaces: %s", err)
	}
	return slice, nil
}

// FetchNamespace fetches a namespace resource from the backend
func (a *NamespaceClient) FetchNamespace(ctx context.Context, name string) (*corev2.Namespace, error) {
	var namespace corev2.Namespace
	if err := a.client.Get(ctx, name, &namespace); err != nil {
		return nil, fmt.Errorf("couldn't get namespace: %s", err)
	}
	return &namespace, nil
}

// CreateNamespace creates a namespace resource
func (a *NamespaceClient) CreateNamespace(ctx context.Context, namespace *corev2.Namespace) error {
	if err := a.client.Create(ctx, namespace); err != nil {
		return fmt.Errorf("couldn't create namespace: %s", err)
	}
	return nil
}

// UpdateNamespace updates a namespace resource
func (a *NamespaceClient) UpdateNamespace(ctx context.Context, namespace *corev2.Namespace) error {
	if err := a.client.Update(ctx, namespace); err != nil {
		return fmt.Errorf("couldn't update namespace: %s", err)
	}
	return nil
}
