package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// UsersController defines the fields required by UsersController.
type UsersController struct {
	Store store.Store
}

// Register should define an association between HTTP routes and their
// respective handlers defined within this Controller.
func (c *UsersController) Register(r *mux.Router) {
	r.HandleFunc("/rbac/users", c.many).Methods(http.MethodGet)
	r.HandleFunc("/rbac/users", c.updateUser).Methods(http.MethodPut)
	r.HandleFunc("/rbac/users/{username}", c.single).Methods(http.MethodGet)
	r.HandleFunc("/rbac/users/{username}", c.deleteUser).Methods(http.MethodDelete)
}

// deleteUser handles DELETE requests to /users/:username
func (c *UsersController) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := types.User{Username: vars["username"]}

	abilities := authorization.Users.WithContext(r.Context())
	if !abilities.CanDelete(&user) {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	if err := c.Store.DeleteUserByName(user.Username); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}

// many handles GET requests to /users
func (c *UsersController) many(w http.ResponseWriter, r *http.Request) {
	abilities := authorization.Users.WithContext(r.Context())
	if !abilities.CanList() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	users, err := c.Store.GetUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reject those resources the viewer is unauthorized to view
	rejectUsers(&users, abilities.CanRead)

	// Obfustace users password for security
	for i := range users {
		users[i].Password = ""
	}

	usersBytes, err := json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(usersBytes))
}

// single handles requests to /users/:username
func (c *UsersController) single(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	var (
		user *types.User
		err  error
	)

	user, err = c.Store.GetUser(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.NotFound(w, r)
		return
	}

	abilities := authorization.Users.WithContext(r.Context())
	if !abilities.CanRead(user) {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	// Obfustace user password for security
	user.Password = ""

	userBytes, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(userBytes))

}

func (c *UsersController) updateUser(w http.ResponseWriter, r *http.Request) {
	var user types.User

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(bodyBytes, &user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Differeniate between create & update
	abilities := authorization.Users.WithContext(r.Context())
	if !abilities.CanUpdate(&user) {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	err = user.Validate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = user.ValidatePassword()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = validateRoles(c.Store, user.Roles)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.Store.CreateUser(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	return
}

func validateRoles(store store.Store, givenRoles []string) error {
	storedRoles, err := store.GetRoles()
	if err != nil {
		return err
	}

	for _, givenRole := range givenRoles {
		if present := hasRole(storedRoles, givenRole); !present {
			return fmt.Errorf("given role '%s' is not valid", givenRole)
		}
	}

	return nil
}

func hasRole(roles []*types.Role, roleName string) bool {
	for _, role := range roles {
		if roleName == role.Name {
			return true
		}
	}
	return false
}

func rejectUsers(users *[]*types.User, predicate func(*types.User) bool) {
	newUsers := make([]*types.User, 0, len(*users))
	for _, user := range *users {
		if predicate(user) {
			newUsers = append(newUsers, user)
		}
	}
	*users = newUsers
}
