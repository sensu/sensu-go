package middlewares

import (
	"context"
	"net/http"
	"strings"

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

		path := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(path, "/")

		// The first part of the path has to be "apis", the API prefix.
		if parts[0] != "apis" {
			return
		}

		if len(parts) >= 2 {
			info.APIGroup = parts[1]
		}

		if len(parts) >= 3 {
			info.APIVersion = parts[2]
		}

		if len(parts) >= 4 {
			// The fourth part of the path has to be "namespaces", the namespace prefix.
			if parts[3] != "namespaces" {
				return
			}
		}

		// A specific namespace is in the path.
		if len(parts) >= 5 {
			info.Namespace = parts[4]
		}

		// A specific resource type is in the path.
		if len(parts) >= 6 {
			info.Resource = parts[5]
		}

		// A specific resource name is in the path.
		if len(parts) >= 7 {
			info.ResourceName = parts[6]
		}
	})
}
