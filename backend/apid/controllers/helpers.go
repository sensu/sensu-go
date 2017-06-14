package controllers

import (
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
)

// organization returns any organization provided as a URL query value
func organization(r *http.Request) string {
	org := r.URL.Query().Get("org")
	if org == "" {
		return "default"
	}
	return org
}

func processRequest(c Controller, req *http.Request) *httptest.ResponseRecorder {
	router := mux.NewRouter()
	c.Register(router)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	return res
}
