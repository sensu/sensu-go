package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/authentication"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// UsersController defines the fields required by UsersController.
type UsersController struct {
	Authentication authentication.Provider
	Store          store.Store
}

// Register should define an association between HTTP routes and their
// respective handlers defined within this Controller.
func (c *UsersController) Register(r *mux.Router) {
	r.HandleFunc("/users", c.many).Methods(http.MethodGet)
	r.HandleFunc("/users", c.updateUser).Methods(http.MethodPut)
}

// many handles GET requests to /users
func (c *UsersController) many(w http.ResponseWriter, r *http.Request) {
	users, err := c.Store.GetUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	err = c.Authentication.CreateUser(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	return
}
