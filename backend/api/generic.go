package api

import (
	"context"
	"errors"
	"fmt"
	"path"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
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
	Kind       corev3.Resource
	Store      storev2.Interface
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

func (g *GenericClient) createResource(ctx context.Context, value corev3.Resource) error {
	req := storev2.NewResourceRequestFromResource(value)
	if gr, ok := g.Kind.(corev3.GlobalResource); !ok || !gr.IsGlobalResource() {
		req.Namespace = corev2.ContextNamespace(ctx)
	}
	wrapper, err := storev2.WrapResource(value)
	if err != nil {
		return err
	}
	return g.Store.CreateIfNotExists(ctx, req, wrapper)
}

// Create creates a resource, if authorized
func (g *GenericClient) Create(ctx context.Context, value corev3.Resource) error {
	if err := g.validateConfig(); err != nil {
		return err
	}
	if err := value.Validate(); err != nil {
		return err
	}
	if err := g.Authorize(ctx, "create", value.GetMetadata().Name); err != nil {
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
	case corev3.Resource:
		g.Kind = kind
	default:
		return fmt.Errorf("%T is not a sensu resource", kind)
	}
	return nil
}

func (g *GenericClient) updateResource(ctx context.Context, value corev3.Resource) error {
	req := storev2.NewResourceRequestFromResource(value)
	if gr, ok := g.Kind.(corev3.GlobalResource); !ok || !gr.IsGlobalResource() {
		req.Namespace = corev2.ContextNamespace(ctx)
	}
	wrapper, err := storev2.WrapResource(value)
	if err != nil {
		return err
	}
	return g.Store.CreateOrUpdate(ctx, req, wrapper)
}

// Update creates or updates a resource, if authorized
func (g *GenericClient) Update(ctx context.Context, value corev3.Resource) error {
	if err := g.validateConfig(); err != nil {
		return err
	}
	if err := value.Validate(); err != nil {
		return err
	}
	if err := g.Authorize(ctx, "update", value.GetMetadata().Name); err != nil {
		return err
	}
	setCreatedBy(ctx, value)
	return g.updateResource(ctx, value)
}

func (g *GenericClient) deleteResource(ctx context.Context, name string) error {
	req := storev2.NewResourceRequestFromResource(g.Kind)
	if gr, ok := g.Kind.(corev3.GlobalResource); !ok || !gr.IsGlobalResource() {
		req.Namespace = corev2.ContextNamespace(ctx)
	}
	req.Name = name
	return g.Store.Delete(ctx, req)
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

func (g *GenericClient) getResource(ctx context.Context, name string, value corev3.Resource) error {
	req := storev2.NewResourceRequestFromResource(g.Kind)
	if gr, ok := g.Kind.(corev3.GlobalResource); !ok || !gr.IsGlobalResource() {
		req.Namespace = corev2.ContextNamespace(ctx)
	}
	req.Name = name
	wrapper, err := g.Store.Get(ctx, req)
	if err != nil {
		return err
	}
	if err := wrapper.UnwrapInto(value); err != nil {
		return err
	}
	if redacter, ok := value.(corev3.Redacter); ok {
		redacter.ProduceRedacted()
	}
	return err
}

// Get gets a resource, if authorized
func (g *GenericClient) Get(ctx context.Context, name string, val corev3.Resource) error {
	if err := g.validateConfig(); err != nil {
		return err
	}
	if err := g.Authorize(ctx, "get", name); err != nil {
		return err
	}
	return g.getResource(ctx, name, val)
}

func (g *GenericClient) list(ctx context.Context, resources interface{}, pred *store.SelectionPredicate) error {
	req := storev2.NewResourceRequestFromResource(g.Kind)
	if gr, ok := g.Kind.(corev3.GlobalResource); !ok || !gr.IsGlobalResource() {
		req.Namespace = corev2.ContextNamespace(ctx)
	}
	if pred != nil && pred.Ordering == "NAME" {
		req.SortOrder = storev2.SortAscend
		if pred.Descending {
			req.SortOrder = storev2.SortDescend
		}
	}
	list, err := g.Store.List(ctx, req, pred)
	if err != nil {
		return err
	}

	if err := list.UnwrapInto(resources); err != nil {
		return err
	}
	if v3Resources, ok := resources.(*[]corev3.Resource); ok {
		for i, resource := range *v3Resources {
			if redacter, ok := resource.(corev3.Redacter); ok {
				(*v3Resources)[i] = redacter.ProduceRedacted()
			}
		}
	}
	return nil
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
