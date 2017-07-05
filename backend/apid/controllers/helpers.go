package controllers

import (
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/middlewares"
)

const (
	defaultOrganization = "default"
)

// organization returns any organization provided in the request context
func organization(r *http.Request) string {
	if value := context.Get(r, middlewares.OrganizationKey); value != nil {
		org, ok := value.(string)
		if ok {
			return org
		}
	}
	return defaultOrganization
}

func processRequest(c Controller, req *http.Request) *httptest.ResponseRecorder {
	router := mux.NewRouter()
	c.Register(router)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	return res
}
