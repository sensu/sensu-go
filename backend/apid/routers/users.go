package routers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
)

// UserController represents the controller needs of the UsersRouter.
type UserController interface {
	List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error)
	Get(ctx context.Context, name string) (*corev2.User, error)
	Create(ctx context.Context, user *corev2.User) error
	CreateOrReplace(ctx context.Context, user *corev2.User) error
	Disable(ctx context.Context, name string) error
	Enable(ctx context.Context, name string) error
	AddGroup(ctx context.Context, name string, group string) error
	RemoveGroup(ctx context.Context, name string, group string) error
	RemoveAllGroups(ctx context.Context, name string) error
	AuthenticateUser(ctx context.Context, username, password string) (*corev2.User, error)
}

// UsersRouter handles requests for /users
type UsersRouter struct {
	controller UserController
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
	routes.List(r.controller.List, corev2.UserFields)
	routes.Get(r.get)
	routes.Post(r.create)
	routes.Del(r.disable)
	routes.Put(r.createOrReplace)

	// Custom
	routes.Path("{id}/{subresource:reinstate}", r.reinstate).Methods(http.MethodPut)
	routes.Path("{id}/{subresource:groups}", r.removeAllGroups).Methods(http.MethodDelete)
	routes.Path("{id}/{subresource:groups}/{user-group-name}", r.addGroup).Methods(http.MethodPut)
	routes.Path("{id}/{subresource:groups}/{user-group-name}", r.removeGroup).Methods(http.MethodDelete)

	// Password change & reset
	routes.Path("{id}/{subresource:password}", r.updatePassword).Methods(http.MethodPut)
	routes.Path("{id}/{subresource:reset_password}", r.resetPassword).Methods(http.MethodPut)
}

func (r *UsersRouter) get(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	user, err := r.controller.Get(req.Context(), id)

	// Obfuscate users password
	if user != nil {
		user.Password = ""
	}
	return user, err
}

func (r *UsersRouter) create(req *http.Request) (interface{}, error) {
	user := &corev2.User{}
	if err := UnmarshalBody(req, user); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	err := r.controller.Create(req.Context(), user)
	return nil, err
}

func (r *UsersRouter) createOrReplace(req *http.Request) (interface{}, error) {
	user := &corev2.User{}
	if err := UnmarshalBody(req, user); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	vars := mux.Vars(req)
	if user.Username != vars["id"] {
		return nil, actions.NewError(actions.InvalidArgument,
			fmt.Errorf(
				"the username (%s) does not match the username on the request (%s)",
				user.Username,
				vars["id"],
			))
	}

	err := r.controller.CreateOrReplace(req.Context(), user)
	return nil, err
}

func (r *UsersRouter) disable(req *http.Request) (interface{}, error) {
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

// updatePassword updates a user password by requiring the current password
func (r *UsersRouter) updatePassword(req *http.Request) (interface{}, error) {
	params := map[string]string{}
	if err := UnmarshalBody(req, &params); err != nil {
		return nil, err
	}

	vars := mux.Vars(req)
	username, err := url.PathUnescape(vars["id"])
	if err != nil {
		return nil, err
	}
	password := params["password"]

	user, err := r.controller.AuthenticateUser(req.Context(), username, password)
	if err != nil {
		return nil, err
	}

	// Remove any old password hash and set the new password hash. The controller
	// will set the resulting hash in both fields before storing it.
	user.Password = ""
	user.PasswordHash = params["password_hash"]
	err = r.controller.CreateOrReplace(req.Context(), user)
	return nil, err
}

// resetPassword updates a user password without any kind of verification
func (r *UsersRouter) resetPassword(req *http.Request) (interface{}, error) {
	params := map[string]string{}
	if err := UnmarshalBody(req, &params); err != nil {
		return nil, err
	}

	vars := mux.Vars(req)
	username, err := url.PathUnescape(vars["id"])
	if err != nil {
		return nil, err
	}

	user, err := r.controller.Get(req.Context(), username)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = params["password_hash"]
	err = r.controller.CreateOrReplace(req.Context(), user)
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
