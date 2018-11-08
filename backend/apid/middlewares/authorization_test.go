package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"

	"github.com/gorilla/mux"
	sensuJWT "github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/seeds"
	"github.com/sensu/sensu-go/backend/store/etcd/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestAuthorization(t *testing.T) {
	// If required, uncomment the following comment to enable the tracing logs
	//logrus.SetLevel(logrus.TraceLevel)

	// Prepare the store
	store, err := testutil.NewStoreInstance()
	if err != nil {
		t.Fatal("Couldn't initialize the store: ", err)
	}
	if err := seeds.SeedInitialData(store); err != nil {
		t.Fatal("Couldn't seed the backend: ", err)
	}

	// TODO (Simon): seed additional roles etc.

	// Prepare our user claims
	clusterAdminClaims := types.Claims{
		StandardClaims: jwt.StandardClaims{Subject: "admin"},
		Groups:         []string{"cluster-admins"},
	}

	cases := []struct {
		description         string
		method              string
		url                 string
		namespace           string
		claims              types.Claims   // JWT claims inserted into the context
		attibutesMiddleware HTTPMiddleware // Legacy or Kubernetes-like routes
		expectedCode        int
	}{
		//
		// cluster-admin ClusterRole & ClusterRoleBinding
		//
		{
			description:         "cluster-admin can list users",
			method:              "GET",
			url:                 "/rbac/users",
			namespace:           "default",
			claims:              clusterAdminClaims,
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        200,
		},
		{
			description:         "cluster-admin can create users",
			method:              "POST",
			url:                 "/rbac/users",
			namespace:           "default",
			claims:              clusterAdminClaims,
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        200,
		},
		{
			description:         "cluster-admin can list ClusterRoles",
			method:              "GET",
			url:                 "/apis/rbac/v2/clusterroles",
			namespace:           "default",
			claims:              clusterAdminClaims,
			attibutesMiddleware: AuthorizationAttributes{},
			expectedCode:        200,
		},
		{
			description:         "cluster-admin can create ClusterRoles",
			method:              "POST",
			url:                 "/apis/rbac/v2/clusterroles",
			namespace:           "default",
			claims:              clusterAdminClaims,
			attibutesMiddleware: AuthorizationAttributes{},
			expectedCode:        200,
		},
		{
			description:         "cluster-admin can access checks in default namespace",
			method:              "GET",
			url:                 "/checks/check-cpu",
			namespace:           "default",
			claims:              clusterAdminClaims,
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        200,
		},
		{
			description:         "cluster-admin can access checks in any namespace",
			method:              "GET",
			url:                 "/checks/check-cpu",
			namespace:           "acme",
			claims:              clusterAdminClaims,
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        200,
		},
	}
	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			// testHandler is a catch-all handler that returns 200 OK
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				return
			})

			// Prepare our HTTP server
			w := httptest.NewRecorder()

			// Prepare the request
			r, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Fatal("Couldn't create request: ", err)
			}

			// Inject the namespace into the path variable
			r = mux.SetURLVars(r, map[string]string{"namespace": tt.namespace})

			// Inject the claims into the request context
			ctx := sensuJWT.SetClaimsIntoContext(r, &tt.claims)

			// Prepare our middlewares
			namespaceMiddleware := Namespace{}
			attributesMiddleware := tt.attibutesMiddleware
			authorizationMiddleware := Authorization{Authorizer: &rbac.Authorizer{Store: store}}

			// Chain our middlewares
			handler := authorizationMiddleware.Then(testHandler)
			handler = attributesMiddleware.Then(handler)
			handler = namespaceMiddleware.Then(handler)

			// Serve the request
			handler.ServeHTTP(w, r.WithContext(ctx))
			assert.Equal(t, tt.expectedCode, w.Code)
			if w.Code >= 200 {
				t.Logf("Response body: %s", w.Body.String())
			}
		})
	}
}
