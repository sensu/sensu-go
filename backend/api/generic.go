package api

import (
	"context"
	"errors"
	"fmt"
	"path"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/types"
)

type RBACVerb string

const (
	VerbGet    RBACVerb = "get"
	VerbList   RBACVerb = "list"
	VerbCreate RBACVerb = "create"
	VerbUpdate RBACVerb = "update"
	VerbDelete RBACVerb = "delete"
)

// GenericClient is a generic API client that uses the ResourceStore.
type GenericClient struct {
	Kind       corev2.Resource
	Store      store.ResourceStore
	StoreV2    storev2.Interface
	Auth       authorization.Authorizer
	APIGroup   string
	APIVersion string
}

func (g GenericClient) validateConfig() error {
	if g.Kind == nil {
		return errors.New("nil Kind")
	}
	_, v3Resource := g.Kind.(*corev3.V2ResourceProxy)
	if g.Store == nil && (g.StoreV2 == nil && v3Resource) {
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

func (g *GenericClient) createResource(ctx context.Context, value corev2.Resource) error {
	if value, ok := value.(*corev3.V2ResourceProxy); ok {
		resource := value.Resource
		req := storev2.NewResourceRequestFromResource(ctx, resource)
		req.Namespace = corev2.ContextNamespace(ctx)
		wrapper, err := storev2.WrapResource(resource)
		if err != nil {
			return err
		}
		return g.StoreV2.CreateIfNotExists(req, wrapper)
	}
	return g.Store.CreateResource(ctx, value)
}

// Create creates a resource, if authorized
func (g *GenericClient) Create(ctx context.Context, value corev2.Resource) error {
	if err := g.validateConfig(); err != nil {
		return err
	}
	if err := value.Validate(); err != nil {
		return err
	}
	if err := g.Authorize(ctx, "create", value.GetObjectMeta().Name); err != nil {
		return err
	}
	setCreatedBy(ctx, value)
	return g.createResource(ctx, value)
}

// SetTypeMeta sets the type of values that the client expects to be dealing
// with. The TypeMeta must match the type of objects that are passed to the
// CRUD methods.
func (g *GenericClient) SetTypeMeta(meta corev2.TypeMeta) error {
	if meta.APIVersion == "" {
		meta.APIVersion = "core/v2"
	}
	g.APIGroup = path.Dir(meta.APIVersion)
	g.APIVersion = path.Base(meta.APIVersion)
	kind, err := types.ResolveRaw(meta.APIVersion, meta.Type)
	if err != nil {
		return fmt.Errorf("error (SetTypeMeta): %s", err)
	}
	switch kind := kind.(type) {
	case corev2.Resource:
		g.Kind = kind
	case corev3.Resource:
		g.Kind = corev3.V3ToV2Resource(kind)
	default:
		return fmt.Errorf("%T is not a sensu resource", kind)
	}
	return nil
}

func (g *GenericClient) updateResource(ctx context.Context, value corev2.Resource) error {
	if value, ok := value.(*corev3.V2ResourceProxy); ok {
		resource := value.Resource
		req := storev2.NewResourceRequestFromResource(ctx, resource)
		req.Namespace = corev2.ContextNamespace(ctx)
		wrapper, err := storev2.WrapResource(resource)
		if err != nil {
			return err
		}
		return g.StoreV2.CreateOrUpdate(req, wrapper)
	}
	return g.Store.CreateOrUpdateResource(ctx, value)
}

// Update creates or updates a resource, if authorized
func (g *GenericClient) Update(ctx context.Context, value corev2.Resource) error {
	if err := g.validateConfig(); err != nil {
		return err
	}
	if err := value.Validate(); err != nil {
		return err
	}
	if err := g.Authorize(ctx, "update", value.GetObjectMeta().Name); err != nil {
		return err
	}
	setCreatedBy(ctx, value)
	return g.updateResource(ctx, value)
}

func (g *GenericClient) deleteResource(ctx context.Context, name string) error {
	if _, ok := g.Kind.(*corev3.V2ResourceProxy); ok {
		req := storev2.ResourceRequest{
			Namespace: corev2.ContextNamespace(ctx),
			Name:      name,
			StoreName: g.Kind.StorePrefix(),
			Context:   ctx,
		}
		return g.StoreV2.Delete(req)
	}
	return g.Store.DeleteResource(ctx, g.Kind.StorePrefix(), name)
}

// Delete deletes a resource, if authorized
func (g *GenericClient) Delete(ctx context.Context, name string) error {
	if err := g.validateConfig(); err != nil {
		return err
	}
	if err := g.Authorize(ctx, "delete", name); err != nil {
		return err
	}
	return g.deleteResource(ctx, name)
}

func (g *GenericClient) getResource(ctx context.Context, name string, value corev2.Resource) error {
	if value, ok := value.(*corev3.V2ResourceProxy); ok {
		req := storev2.ResourceRequest{
			Namespace: corev2.ContextNamespace(ctx),
			Name:      name,
			StoreName: value.StorePrefix(),
			Context:   ctx,
		}
		wrapper, err := g.StoreV2.Get(req)
		if err != nil {
			return err
		}
		return wrapper.UnwrapInto(value.Resource)
	}
	return g.Store.GetResource(ctx, name, value)
}

// Get gets a resource, if authorized
func (g *GenericClient) Get(ctx context.Context, name string, val corev2.Resource) error {
	if err := g.validateConfig(); err != nil {
		return err
	}
	if err := g.Authorize(ctx, "get", name); err != nil {
		return err
	}
	return g.getResource(ctx, name, val)
}

func (g *GenericClient) list(ctx context.Context, resources interface{}, pred *store.SelectionPredicate) error {
	if _, ok := g.Kind.(*corev3.V2ResourceProxy); ok {
		req := storev2.ResourceRequest{
			Namespace: corev2.ContextNamespace(ctx),
			StoreName: g.Kind.StorePrefix(),
			Context:   ctx,
		}
		list, err := g.StoreV2.List(req, pred)
		if err != nil {
			return err
		}
		if resourceList, ok := resources.(*[]corev2.Resource); ok {
			values, err := list.Unwrap()
			if err != nil {
				return err
			}
			v2values := make([]corev2.Resource, len(values))
			for i := range values {
				v2values[i] = corev3.V3ToV2Resource(values[i])
			}
			*resourceList = v2values
			return nil
		}
		return list.UnwrapInto(resources)
	}
	return g.Store.ListResources(ctx, g.Kind.StorePrefix(), resources, pred)
}

// List lists all resources within a namespace, according to a selection
// predicate, if authorized
func (g *GenericClient) List(ctx context.Context, resources interface{}, pred *store.SelectionPredicate) error {
	if err := g.validateConfig(); err != nil {
		return err
	}
	if err := g.Authorize(ctx, "list", ""); err != nil {
		return err
	}
	return g.list(ctx, resources, pred)
}

// Authorize tests whether or not the current user can perform an action.
// Returns nil if action is allow and otherwise an auth error.
func (g *GenericClient) Authorize(ctx context.Context, verb RBACVerb, name string) error {
	attrs := &authorization.Attributes{
		APIGroup:     g.APIGroup,
		APIVersion:   g.APIVersion,
		Namespace:    corev2.ContextNamespace(ctx),
		Resource:     g.Kind.RBACName(),
		ResourceName: name,
		Verb:         string(verb),
	}
	if err := authorize(ctx, g.Auth, attrs); err != nil {
		return err
	}
	return nil
}
