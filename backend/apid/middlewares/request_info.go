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
// The general, expected format of a path is as
// follow: /apis/{group}/{version}/namespaces/{namespace}/{type}/{name}
//
// This middleware tries to fill in as many fields as it can in the
// types.RequestInfo struct that's added to the context, but there is no guarantee
// of any fields being filled in.
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

		// We need at least the API prefix, the API group and the API version to
		// build a meaningful RequestInfo.
		if len(parts) < 3 {
			return
		}

		// The first part of the path has to be "apis"
		if parts[0] != "apis" {
			return
		}

		info.APIGroup = parts[1]
		info.APIVersion = parts[2]
	})
}
