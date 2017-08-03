package controllers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/types"
)

func processRequest(c Controller, req *http.Request) *httptest.ResponseRecorder {
	router := mux.NewRouter()
	c.Register(router)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	return res
}

func newRequest(meth, url string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(meth, url, body)
	req = requestWithDefaultContext(req)

	return req
}

func requestWithDefaultContext(req *http.Request) *http.Request {
	req = requestWithEnvironment(req, "default")
	req = requestWithOrganization(req, "default")
	req = requestWithFullAccess(req)

	return req
}

func requestWithEnvironment(r *http.Request, environment string) *http.Request {
	context := context.WithValue(
		r.Context(),
		types.EnvironmentKey,
		environment,
	)

	return r.WithContext(context)
}

func requestWithOrganization(r *http.Request, organization string) *http.Request {
	context := context.WithValue(
		r.Context(),
		types.OrganizationKey,
		organization,
	)

	return r.WithContext(context)
}

func requestWithFullAccess(r *http.Request) *http.Request {
<<<<<<< HEAD
	userRules := []types.Rule{*types.FixtureRule("*")}
	actor := authorization.Actor{Name: "sensu", Rules: userRules}
=======
	userRoles := []*types.Role{types.FixtureRole("test", "*", "*")}
>>>>>>> Add environments to RBAC
	context := context.WithValue(
		r.Context(),
		types.AuthorizationActorKey,
		actor,
	)

	return r.WithContext(context)
}

func requestWithNoAccess(r *http.Request) *http.Request {
	context := context.WithValue(
		r.Context(),
		types.AuthorizationActorKey,
		authorization.Actor{},
	)

	return r.WithContext(context)
}
