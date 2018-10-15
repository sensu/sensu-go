package middlewares

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/types"
)

// RequestInfo is an HTTP middleware that populates a context for the request,
// to be used by other middlewares. You probably want this middleware to be
// executed early in the middleware stack.
type RequestInfo struct{}

// Then is the RequestInfo middleware's main logic, heavily inspired by
// Kubernetes' WithRequestInfo.
//
// The general, expected format of a path one of the following:
// - /apis/{group}/{version}/namespaces, to create or list namespaces
// - /apis/{group}/{version}/namespaces/{namespace}, to retrieve, update or
//   delete namespace {namespace}
// - /apis/{group}/{version}/namespaces/{namespace}/{type}, to create or list
//   objects of type {type} in namespace {namespace}
// - /apis/{group}/{version}/namespaces/{namespace}/{type}/{name}, to retrieve,
//   update or delete object {name} of type {type} in namespace {namespace}
//
// This middleware tries to fill in as many fields as it can in the
// types.RequestInfo struct added to the context, but there is no guarantee of
// any fields being filled in.
func (i RequestInfo) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		info := types.RequestInfo{}

		router := mux.NewRouter().PathPrefix("/apis/{group}/{version}/namespaces/").Subrouter()
		router.Path("").Methods("GET", "POST")
		router.Path("/{namespace}").Methods("GET", "PUT", "DELETE")
		router.Path("/{namespace}/{type}").Methods("GET", "POST")
		router.Path("/{namespace}/{type}/{name}").Methods("GET", "PUT", "DELETE")

		ctx = context.WithValue(ctx, types.RequestInfoKey, &info)
		defer next.ServeHTTP(w, r.WithContext(ctx))

		match := mux.RouteMatch{}
		if router.Match(r, &match) {
			switch r.Method {
			case "POST":
				info.Verb = "create"
			case "GET", "HEAD":
				info.Verb = "get"
			case "PUT":
				info.Verb = "update"
			case "DELETE":
				info.Verb = "delete"
			default:
				info.Verb = ""
			}

			info.APIGroup = match.Vars["group"]
			info.APIVersion = match.Vars["version"]
			info.Namespace = match.Vars["namespace"]
			info.Resource = match.Vars["type"]
			info.ResourceName = match.Vars["name"]
		}
	})
}
