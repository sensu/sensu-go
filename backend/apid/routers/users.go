package routers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// UsersRouter handles requests for /users
type UsersRouter struct {
	controller actions.UserController
}

// NewUsersRouter instantiates new router for controlling user resources
func NewUsersRouter(store store.Store) *UsersRouter {
	return &UsersRouter{
		controller: actions.NewUserController(store),
	}
}

// Mount the UsersRouter to a parent Router
func (r *UsersRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/{resource:users}",
	}
	routes.List(r.list)
	routes.Get(r.find)
	routes.Post(r.create)
	routes.Del(r.destroy)
	routes.Put(r.createOrReplace)

	// Custom
	routes.Path("{id}/{subresource:reinstate}", r.reinstate).Methods(http.MethodPut)
	routes.Path("{id}/{subresource:groups}", r.removeAllGroups).Methods(http.MethodDelete)
	routes.Path("{id}/{subresource:groups}/{user-group-name}", r.addGroup).Methods(http.MethodPut)
	routes.Path("{id}/{subresource:groups}/{user-group-name}", r.removeGroup).Methods(http.MethodDelete)

	// TODO: Remove?
	routes.Path("{id}/{subresource:password}", r.updatePassword).Methods(http.MethodPut)
}

func (r *UsersRouter) list(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	records, err := r.controller.Query(req.Context())

	// Obfustace users password
	for i := range records {
		records[i].Password = ""
	}

	return records, err
}

func (r *UsersRouter) find(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	record, err := r.controller.Find(req.Context(), id)

	// Obfustace users password
	record.Password = ""
	return record, err
}

func (r *UsersRouter) create(req *http.Request) (interface{}, error) {
	cfg := types.User{}
	if err := UnmarshalBody(req, &cfg); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), cfg)

	// Hide user's password
	cfg.Password = ""
	return cfg, err
}

func (r *UsersRouter) createOrReplace(req *http.Request) (interface{}, error) {
	cfg := types.User{}
	if err := UnmarshalBody(req, &cfg); err != nil {
		return nil, err
	}

	err := r.controller.CreateOrReplace(req.Context(), cfg)

	// Hide user's password
	cfg.Password = ""
	return cfg, err
}

func (r *UsersRouter) destroy(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	err = r.controller.Disable(req.Context(), id)
	return nil, err
}

func (r *UsersRouter) reinstate(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	err = r.controller.Enable(req.Context(), id)
	return nil, err
}

func (r *UsersRouter) updatePassword(req *http.Request) (interface{}, error) {
	params := map[string]string{}
	if err := UnmarshalBody(req, &params); err != nil {
		return nil, err
	}

	vars := mux.Vars(req)
	id, err := url.PathUnescape(vars["id"])
	if err != nil {
		return nil, err
	}

	record, err := r.controller.Find(req.Context(), id)
	if err != nil {
		return nil, err
	}

	record.Password = params["password"]
	err = r.controller.CreateOrReplace(req.Context(), *record)
	return nil, err
}

func (r *UsersRouter) addGroup(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}

	group, err := url.PathUnescape(params["user-group-name"])
	if err != nil {
		return nil, err
	}

	err = r.controller.AddGroup(req.Context(), id, group)
	return nil, err
}

func (r *UsersRouter) removeGroup(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}

	group, err := url.PathUnescape(params["user-group-name"])
	if err != nil {
		return nil, err
	}

	err = r.controller.RemoveGroup(req.Context(), id, group)
	return nil, err
}

func (r *UsersRouter) removeAllGroups(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}

	err = r.controller.RemoveAllGroups(req.Context(), id)
	return nil, err
}
