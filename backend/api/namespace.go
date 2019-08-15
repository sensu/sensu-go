package api

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
)

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

// ListNamespaces fetches a list of namespace resources, if authorized.
func (a *NamespaceClient) ListNamespaces(ctx context.Context) ([]*corev2.Namespace, error) {
	pred := &store.SelectionPredicate{
		Continue: corev2.PageContinueFromContext(ctx),
		Limit:    int64(corev2.PageSizeFromContext(ctx)),
	}
	slice := []*corev2.Namespace{}
	if err := a.client.List(ctx, &slice, pred); err != nil {
		return nil, err
	}
	return slice, nil
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
