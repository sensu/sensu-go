package routers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
)

type NamespaceClient interface {
	ListNamespaces(context.Context, *store.SelectionPredicate) ([]*corev3.Namespace, error)
	CreateNamespace(context.Context, *corev3.Namespace) error
	UpdateNamespace(context.Context, *corev3.Namespace) error
	DeleteNamespace(context.Context, string) error
	FetchNamespace(context.Context, string) (*corev3.Namespace, error)
}

type NamespacePatcher interface {
	PatchResource(*http.Request) (corev3.Resource, error)
}

// NamespacesRouter handles requests for /namespaces
type NamespacesRouter struct {
	client  NamespaceClient
	patcher NamespacePatcher
}

// NewNamespacesRouter instantiates new router for controlling check resources
func NewNamespacesRouter(client NamespaceClient, patcher NamespacePatcher) *NamespacesRouter {
	return &NamespacesRouter{
		client:  client,
		patcher: patcher,
	}
}

// Mount the NamespacesRouter to a parent Router
func (r *NamespacesRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/{resource:namespaces}",
	}

	routes.Del(r.delete)
	routes.Get(r.get)
	routes.List(r.list, corev2.NamespaceFields)
	routes.Post(r.create)
	routes.Patch(r.patcher.PatchResource)
	routes.Put(r.update)
}

func (r *NamespacesRouter) get(req *http.Request) (corev3.Resource, error) {
	params := mux.Vars(req)
	name, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}
	ns, err := r.client.FetchNamespace(req.Context(), name)
	if err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return nil, actions.NewErrorf(actions.NotFound)
		case *store.ErrNotValid:
			return nil, actions.NewError(actions.InvalidArgument, err)
		default:
			return nil, actions.NewError(actions.InternalErr, err)
		}
	}
	return ns, nil
}

func (r *NamespacesRouter) list(ctx context.Context, pred *store.SelectionPredicate) ([]corev3.Resource, error) {
	namespaces, err := r.client.ListNamespaces(ctx, pred)
	if err != nil {
		return nil, err
	}
	result := make([]corev3.Resource, len(namespaces))
	for i := range namespaces {
		result[i] = namespaces[i]
	}
	return result, nil
}

func (r *NamespacesRouter) create(req *http.Request) (corev3.Resource, error) {
	ctx := req.Context()
	var ns corev3.Namespace
	if err := json.NewDecoder(req.Body).Decode(&ns); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}
	meta := ns.GetMetadata()
	if claims := jwt.GetClaimsFromContext(ctx); claims != nil {
		meta.CreatedBy = claims.StandardClaims.Subject
		ns.Metadata = meta
	}
	if err := handlers.CheckMeta(&ns, mux.Vars(req), "id"); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}
	if err := ns.Validate(); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}
	if err := r.client.CreateNamespace(ctx, &ns); err != nil {
		switch err := err.(type) {
		case *store.ErrAlreadyExists:
			return nil, actions.NewErrorf(actions.AlreadyExistsErr)
		case *store.ErrNotValid:
			return nil, actions.NewError(actions.InvalidArgument, err)
		default:
			return nil, actions.NewError(actions.InternalErr, err)
		}
	}
	return nil, nil
}

func (r *NamespacesRouter) update(req *http.Request) (corev3.Resource, error) {
	ctx := req.Context()
	var ns corev3.Namespace
	if err := json.NewDecoder(req.Body).Decode(&ns); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}
	meta := ns.GetMetadata()
	if claims := jwt.GetClaimsFromContext(ctx); claims != nil {
		meta.CreatedBy = claims.StandardClaims.Subject
		ns.Metadata = meta
	}
	if err := handlers.CheckMeta(&ns, mux.Vars(req), "id"); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}
	if err := ns.Validate(); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}
	if err := r.client.UpdateNamespace(ctx, &ns); err != nil {
		switch err := err.(type) {
		case *store.ErrNotValid:
			return nil, actions.NewError(actions.InvalidArgument, err)
		default:
			return nil, actions.NewError(actions.InternalErr, err)
		}
	}
	return nil, nil
}

func (r *NamespacesRouter) delete(req *http.Request) (corev3.Resource, error) {
	params := mux.Vars(req)
	name, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	if err := r.client.DeleteNamespace(req.Context(), name); err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return nil, actions.NewErrorf(actions.NotFound)
		default:
			return nil, actions.NewError(actions.InternalErr, err)
		}
	}

	return nil, nil
}
