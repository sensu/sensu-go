package api

import (
	"context"
	"errors"
	"fmt"
	"path"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// GenericClient is a generic API client that uses the ResourceStore.
type GenericClient struct {
	Kind       corev2.Resource
	Store      store.ResourceStore
	Auth       authorization.Authorizer
	APIGroup   string
	APIVersion string
}

func (g GenericClient) validateConfig() error {
	if g.Kind == nil {
		return errors.New("nil Kind")
	}
	if g.Store == nil {
		return errors.New("nil store")
	}
	if g.Auth == nil {
		return errors.New("nil auth")
	}
	if g.APIGroup == "" || g.APIVersion == "" {
		return errors.New("empty api group/version")
	}
	return nil
}

// Create creates a resource, if authorized
func (g *GenericClient) Create(ctx context.Context, value corev2.Resource) error {
	if err := g.validateConfig(); err != nil {
		return err
	}
	if err := value.Validate(); err != nil {
		return err
	}
	attrs := &authorization.Attributes{
		APIGroup:     g.APIGroup,
		APIVersion:   g.APIVersion,
		Resource:     path.Base(g.Kind.StorePrefix()),
		Namespace:    corev2.ContextNamespace(ctx),
		Verb:         "create",
		ResourceName: value.GetObjectMeta().Name,
	}
	if err := authorize(ctx, g.Auth, attrs); err != nil {
		return err
	}
	return g.Store.CreateResource(ctx, value)
}

func (g *GenericClient) resource() string {
	return path.Base(g.Kind.StorePrefix())
}

// SetTypeMeta sets the type of values that the client expects to be dealing
// with. The TypeMeta must match the type of objects that are passed to the
// CRUD methods.
func (g *GenericClient) SetTypeMeta(meta corev2.TypeMeta) error {
	if meta.APIVersion == "" {
		meta.APIVersion = "core/v2"
	}
	g.APIGroup, g.APIVersion = path.Split(meta.APIVersion)
	kind, err := types.ResolveType(meta.APIVersion, meta.Type)
	if err != nil {
		return fmt.Errorf("error (SetTypeMeta): %s", err)
	}
	g.Kind = kind
	return nil
}

// Update creates or updates a resource, if authorized
func (g *GenericClient) Update(ctx context.Context, value corev2.Resource) error {
	if err := g.validateConfig(); err != nil {
		return err
	}
	if err := value.Validate(); err != nil {
		return err
	}
	attrs := &authorization.Attributes{
		APIGroup:     g.APIGroup,
		APIVersion:   g.APIVersion,
		Resource:     g.resource(),
		Namespace:    corev2.ContextNamespace(ctx),
		Verb:         "update",
		ResourceName: value.GetObjectMeta().Name,
	}
	if err := authorize(ctx, g.Auth, attrs); err != nil {
		return err
	}
	return g.Store.CreateOrUpdateResource(ctx, value)
}

// Delete deletes a resource, if authorized
func (g *GenericClient) Delete(ctx context.Context, name string) error {
	if err := g.validateConfig(); err != nil {
		return err
	}
	attrs := &authorization.Attributes{
		APIGroup:     g.APIGroup,
		APIVersion:   g.APIVersion,
		Resource:     g.resource(),
		Namespace:    corev2.ContextNamespace(ctx),
		Verb:         "delete",
		ResourceName: name,
	}
	if err := authorize(ctx, g.Auth, attrs); err != nil {
		return err
	}
	return g.Store.DeleteResource(ctx, g.Kind.StorePrefix(), name)
}

// Get gets a resource, if authorized
func (g *GenericClient) Get(ctx context.Context, name string, val corev2.Resource) error {
	if err := g.validateConfig(); err != nil {
		return err
	}
	attrs := &authorization.Attributes{
		APIGroup:     g.APIGroup,
		APIVersion:   g.APIVersion,
		Resource:     g.resource(),
		Namespace:    corev2.ContextNamespace(ctx),
		Verb:         "get",
		ResourceName: name,
	}
	if err := authorize(ctx, g.Auth, attrs); err != nil {
		return err
	}
	return g.Store.GetResource(ctx, name, val)
}

// List lists all resources within a namespace, according to a selection
// predicate, if authorized
func (g *GenericClient) List(ctx context.Context, resources interface{}, pred *store.SelectionPredicate) error {
	if err := g.validateConfig(); err != nil {
		return err
	}
	attrs := &authorization.Attributes{
		APIGroup:   g.APIGroup,
		APIVersion: g.APIVersion,
		Resource:   g.resource(),
		Namespace:  corev2.ContextNamespace(ctx),
		Verb:       "list",
	}
	if err := authorize(ctx, g.Auth, attrs); err != nil {
		return err
	}
	return g.Store.ListResources(ctx, g.Kind.StorePrefix(), resources, pred)
}
