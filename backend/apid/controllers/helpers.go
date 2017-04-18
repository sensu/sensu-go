package controllers

import (
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
)

func processRequest(c Controller, req *http.Request) *httptest.ResponseRecorder {
	router := mux.NewRouter()
	c.Register(router)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	return res
}
