package api

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// not meant to be used outside this package
type genericClient struct {
	Kind       types.Resource
	Store      store.ResourceStore
	Auth       authorization.Authorizer
	Resource   string
	APIGroup   string
	APIVersion string
}

func (g *genericClient) Create(ctx context.Context, value types.Resource) error {
	attrs := &authorization.Attributes{
		APIGroup:     g.APIGroup,
		APIVersion:   g.APIVersion,
		Resource:     g.Resource,
		Namespace:    corev2.ContextNamespace(ctx),
		Verb:         "create",
		ResourceName: value.GetObjectMeta().Name,
	}
	if err := authorize(ctx, g.Auth, attrs); err != nil {
		return err
	}
	return c.Store.CreateResource(ctx, value)
}

func (g *genericClient) Update(ctx context.Context, value types.Resource) error {
	attrs := &authorization.Attributes{
		APIGroup:     g.APIGroup,
		APIVersion:   g.APIVersion,
		Resource:     g.Resource,
		Namespace:    corev2.ContextNamespace(ctx),
		Verb:         "update",
		ResourceName: value.GetObjectMeta().Name,
	}
	if err := authorize(ctx, g.Auth, attrs); err != nil {
		return err
	}
	return c.Store.CreateOrUpdateResource(ctx, value)
}

func (g *genericClient) Delete(ctx context.Context, name string) error {
	attrs := &authorization.Attributes{
		APIGroup:     g.APIGroup,
		APIVersion:   g.APIVersion,
		Resource:     g.Resource,
		Namespace:    corev2.NamespaceResource(ctx),
		Verb:         "delete",
		ResourceName: name,
	}
	if err := authorize(ctx, g.Auth, attrs); err != nil {
		return err
	}
	return c.Store.DeleteResource(ctx, g.Kind, name)
}
