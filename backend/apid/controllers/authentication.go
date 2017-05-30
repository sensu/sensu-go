package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
)

// AuthenticationController handles authentication related requests
type AuthenticationController struct{}

// Register the EventsController with a mux.Router.
func (a *AuthenticationController) Register(r *mux.Router) {
	r.HandleFunc("/auth", a.login).Methods(http.MethodGet)
}

func (a *AuthenticationController) login(w http.ResponseWriter, r *http.Request) {
	// Add this point the user/password have already been validated by the
	// authentication middleware and the signed JWT has been added to the response
	// body. We just need to set the response code appropriately
	w.WriteHeader(http.StatusOK)
	return
}
