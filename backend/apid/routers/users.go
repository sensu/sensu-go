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
	routes := resourceRoute{router: parent, pathPrefix: "/rbac/users"}
	routes.index(r.list)
	routes.show(r.find)
	routes.create(r.create)
	routes.update(r.update)
	routes.destroy(r.destroy)

	// Custom
	routes.path("{id}/reinstate", r.reinstate).Methods(http.MethodPut)
	routes.path("{id}/roles/{role}", r.addRole).Methods(http.MethodPut)
	routes.path("{id}/roles/{role}", r.removeRole).Methods(http.MethodDelete)

	// TODO: Remove?
	routes.path("{id}/password", r.updatePassword).Methods(http.MethodPut)
}

func (r *UsersRouter) list(req *http.Request) (interface{}, error) {
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
	if err := unmarshalBody(req, &cfg); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), cfg)

	// Obfustace users password
	cfg.Password = ""
	return cfg, err
}

func (r *UsersRouter) update(req *http.Request) (interface{}, error) {
	cfg := types.User{}
	if err := unmarshalBody(req, &cfg); err != nil {
		return nil, err
	}

	err := r.controller.Update(req.Context(), cfg)

	// Obfustace users password
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
	if err := unmarshalBody(req, &params); err != nil {
		return nil, err
	}

	vars := mux.Vars(req)
	id, err := url.PathUnescape(vars["id"])
	if err != nil {
		return nil, err
	}
	cfg := types.User{Username: id, Password: params["password"]}

	err = r.controller.Update(req.Context(), cfg)
	return nil, err
}

func (r *UsersRouter) addRole(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	role, err := url.PathUnescape(params["role"])
	if err != nil {
		return nil, err
	}
	err = r.controller.AddRole(req.Context(), id, role)
	return nil, err
}

func (r *UsersRouter) removeRole(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	role, err := url.PathUnescape(params["role"])
	if err != nil {
		return nil, err
	}
	err = r.controller.RemoveRole(req.Context(), id, role)
	return nil, err
}
