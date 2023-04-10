package routers

import (
	"context"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/apid/request"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// SilencedRouter handles requests for /users
type SilencedRouter struct {
	controller silencedController
	store      storev2.Interface
}

// silencedController represents the controller needs of the SilencedRouter.
type silencedController interface {
	Create(ctx context.Context, entry *corev2.Silenced) error
	CreateOrReplace(ctx context.Context, entry *corev2.Silenced) error
	List(ctx context.Context, sub, check string) ([]*corev2.Silenced, error)
	Get(ctx context.Context, name string) (*corev2.Silenced, error)
}

// NewSilencedRouter instantiates new router for controlling user resources
func NewSilencedRouter(store storev2.Interface) *SilencedRouter {
	return &SilencedRouter{
		controller: actions.NewSilencedController(store),
		store:      store,
	}
}

// Mount the SilencedRouter to a parent Router
func (r *SilencedRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:silenced}",
	}

	handlers := handlers.NewHandlers[*corev2.Silenced](r.store)

	routes.Del(handlers.DeleteResource)
	routes.Get(r.get)
	routes.Post(r.create)
	routes.Put(r.createOrReplace)
	routes.List(r.listr, corev3.SilencedFields)
	routes.ListAllNamespaces(r.listr, "/{resource:silenced}", corev3.SilencedFields)

	// Custom routes for listing by subscription and checks for a specific
	// namespace, in addition to all namespaces for checks.
	routes.Router.HandleFunc("/{resource:silenced}/checks/{check}", listHandler(r.list)).Methods(http.MethodGet)
	routes.Router.HandleFunc(routes.PathPrefix+"/subscriptions/{subscription}", listHandler(r.list)).Methods(http.MethodGet)
	routes.Router.HandleFunc(routes.PathPrefix+"/checks/{check}", listHandler(r.list)).Methods(http.MethodGet)
}

func (r *SilencedRouter) get(req *http.Request) (corev3.Resource, error) {
	id, err := url.PathUnescape(mux.Vars(req)["id"])
	if err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}
	return r.controller.Get(req.Context(), id)
}

func (r *SilencedRouter) create(req *http.Request) (corev3.Resource, error) {
	entry, err := request.Resource[*corev2.Silenced](req)
	if err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	if err := handlers.CheckMeta(entry, mux.Vars(req), "id"); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	err = r.controller.Create(req.Context(), entry)
	return nil, err
}

func (r *SilencedRouter) createOrReplace(req *http.Request) (corev3.Resource, error) {
	entry, err := request.Resource[*corev2.Silenced](req)
	if err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	if err := handlers.CheckMeta(entry, mux.Vars(req), "id"); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	err = r.controller.CreateOrReplace(req.Context(), entry)
	return nil, err
}

func (r *SilencedRouter) listr(ctx context.Context, pred *store.SelectionPredicate) ([]corev3.Resource, error) {
	entries, err := r.controller.List(ctx, "", "")
	if err != nil {
		return nil, err
	}
	result := make([]corev3.Resource, 0, len(entries))
	for _, e := range entries {
		result = append(result, e)
	}
	return result, nil
}

func (r *SilencedRouter) list(w http.ResponseWriter, req *http.Request) ([]corev3.Resource, error) {
	params := mux.Vars(req)
	entries, err := r.controller.List(req.Context(), params["subscription"], params["check"])
	if err != nil {
		return nil, err
	}
	result := make([]corev3.Resource, 0, len(entries))
	for _, resource := range entries {
		result = append(result, resource)
	}
	return result, nil
}
