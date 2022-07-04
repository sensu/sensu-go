package routers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/api"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// NamespacesRouter handles requests for /namespaces
type NamespacesRouter struct {
	handlers handlers.Handlers
	storev2  storev2.Interface
	auth     authorization.Authorizer
}

// NewNamespacesRouter instantiates new router for controlling check resources
func NewNamespacesRouter(auth authorization.Authorizer, store storev2.Interface) *NamespacesRouter {
	return &NamespacesRouter{
		auth: auth,
		handlers: handlers.Handlers{
			Resource: &corev2.Namespace{},
			StoreV2:  store,
		},
		storev2: store,
	}
}

// Mount the NamespacesRouter to a parent Router
func (r *NamespacesRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/{resource:namespaces}",
	}

	routes.Del(r.delete)
	routes.Get(r.handlers.GetResource)
	routes.List(r.list, corev2.NamespaceFields)
	routes.Post(r.create)
	routes.Patch(r.handlers.PatchResource)
	routes.Put(r.update)
}

func (r *NamespacesRouter) list(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	client := api.NewNamespaceClient(r.store, r.namespaceStore, r.auth, r.storev2)
	namespaces, err := client.ListNamespaces(ctx, pred)
	if err != nil {
		return nil, err
	}
	result := make([]corev2.Resource, len(namespaces))
	for i := range namespaces {
		result[i] = namespaces[i]
	}
	return result, nil
}

func (r *NamespacesRouter) create(req *http.Request) (interface{}, error) {
	ctx := req.Context()
	var ns corev2.Namespace
	if err := json.NewDecoder(req.Body).Decode(&ns); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}
	meta := ns.GetObjectMeta()
	if claims := jwt.GetClaimsFromContext(ctx); claims != nil {
		meta.CreatedBy = claims.StandardClaims.Subject
		ns.SetObjectMeta(meta)
	}
	if err := handlers.CheckMeta(&ns, mux.Vars(req), "id"); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}
	if err := ns.Validate(); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}
	client := api.NewNamespaceClient(r.store, r.namespaceStore, r.auth, r.storev2)
	if err := client.CreateNamespace(ctx, &ns); err != nil {
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

func (r *NamespacesRouter) update(req *http.Request) (interface{}, error) {
	ctx := req.Context()
	var ns corev2.Namespace
	if err := json.NewDecoder(req.Body).Decode(&ns); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}
	meta := ns.GetObjectMeta()
	if claims := jwt.GetClaimsFromContext(ctx); claims != nil {
		meta.CreatedBy = claims.StandardClaims.Subject
		ns.SetObjectMeta(meta)
	}
	if err := handlers.CheckMeta(&ns, mux.Vars(req), "id"); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}
	if err := ns.Validate(); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}
	client := api.NewNamespaceClient(r.store, r.namespaceStore, r.auth, r.storev2)
	if err := client.UpdateNamespace(ctx, &ns); err != nil {
		switch err := err.(type) {
		case *store.ErrNotValid:
			return nil, actions.NewError(actions.InvalidArgument, err)
		default:
			return nil, actions.NewError(actions.InternalErr, err)
		}
	}
	return nil, nil
}

func (r *NamespacesRouter) delete(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	name, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	client := api.NewNamespaceClient(r.store, r.namespaceStore, r.auth, r.storev2)
	if err := client.DeleteNamespace(req.Context(), name); err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return nil, actions.NewErrorf(actions.NotFound)
		default:
			return nil, actions.NewError(actions.InternalErr, err)
		}
	}

	return nil, nil
}
