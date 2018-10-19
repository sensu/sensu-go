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
// - /apis/{group}/{version}/namespaces/{name}
// - /apis/{group}/{version}/namespaces/{namespace}/{kind}
// - /apis/{group}/{version}/namespaces/{namespace}/{kind}/{name}
//
// This middleware tries to fill in as many fields as it can in the
// types.RequestInfo struct added to the context, but there is no guarantee of
// any fields being filled in.
func (i RequestInfo) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		info := types.RequestInfo{}

		ctx = context.WithValue(ctx, types.RequestInfoKey, &info)
		defer next.ServeHTTP(w, r.WithContext(ctx))

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

		vars := mux.Vars(r)
		info.APIGroup = vars["group"]
		info.APIVersion = vars["version"]
		info.Namespace = vars["namespace"]
		info.Resource = vars["kind"]
		info.ResourceName = vars["name"]

		if info.Verb == "get" && info.ResourceName == "" {
			info.Verb = "list"
		}
	})
}
