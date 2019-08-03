package api

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
)

// not meant to be used outside this package for now
type genericClient struct {
	Kind       corev2.Resource
	Store      store.ResourceStore
	Auth       authorization.Authorizer
	Resource   string
	APIGroup   string
	APIVersion string
}

func (g *genericClient) Create(ctx context.Context, value corev2.Resource) error {
	if err := value.Validate(); err != nil {
		return err
	}
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
	return g.Store.CreateResource(ctx, value)
}

func (g *genericClient) Update(ctx context.Context, value corev2.Resource) error {
	if err := value.Validate(); err != nil {
		return err
	}
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
	return g.Store.CreateOrUpdateResource(ctx, value)
}

func (g *genericClient) Delete(ctx context.Context, name string) error {
	attrs := &authorization.Attributes{
		APIGroup:     g.APIGroup,
		APIVersion:   g.APIVersion,
		Resource:     g.Resource,
		Namespace:    corev2.ContextNamespace(ctx),
		Verb:         "delete",
		ResourceName: name,
	}
	if err := authorize(ctx, g.Auth, attrs); err != nil {
		return err
	}
	return g.Store.DeleteResource(ctx, g.Kind.StorePrefix(), name)
}

func (g *genericClient) Get(ctx context.Context, name string, val corev2.Resource) error {
	attrs := &authorization.Attributes{
		APIGroup:     g.APIGroup,
		APIVersion:   g.APIVersion,
		Resource:     g.Resource,
		Namespace:    corev2.ContextNamespace(ctx),
		Verb:         "get",
		ResourceName: name,
	}
	if err := authorize(ctx, g.Auth, attrs); err != nil {
		return err
	}
	return g.Store.GetResource(ctx, name, val)
}

func (g *genericClient) List(ctx context.Context, resources interface{}, pred *store.SelectionPredicate) error {
	attrs := &authorization.Attributes{
		APIGroup:   g.APIGroup,
		APIVersion: g.APIVersion,
		Resource:   g.Resource,
		Namespace:  corev2.ContextNamespace(ctx),
		Verb:       "list",
	}
	if err := authorize(ctx, g.Auth, attrs); err != nil {
		return err
	}
	return g.Store.ListResources(ctx, g.Kind.StorePrefix(), resources, pred)
}
