package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"

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
	logrus.SetLevel(logrus.TraceLevel)

	// Prepare the store
	store, err := testutil.NewStoreInstance()
	if err != nil {
		t.Fatal("Could not initialize the store: ", err)
	}
	if err := seeds.SeedInitialData(store); err != nil {
		t.Fatal("Could not seed the backend: ", err)
	}

	// Add a ClusterRoleBinding for the ClusteRole admin and assign the
	// local-admins group
	localAdmins := &types.ClusterRoleBinding{
		Name: "admin",
		RoleRef: types.RoleRef{
			Kind: "ClusterRole",
			Name: "admin",
		},
		Subjects: []types.Subject{
			types.Subject{
				Kind: "Group",
				Name: "local-admins",
			},
		},
	}
	if err := store.CreateClusterRoleBinding(context.Background(), localAdmins); err != nil {
		t.Fatal("Could not add the admin ClusterRoleBinding")
	}

	admins := &types.RoleBinding{
		Name:      "admin",
		Namespace: "default",
		RoleRef: types.RoleRef{
			Kind: "ClusterRole",
			Name: "admin",
		},
		Subjects: []types.Subject{
			types.Subject{
				Kind: "Group",
				Name: "admins",
			},
		},
	}
	if err := store.CreateRoleBinding(context.Background(), admins); err != nil {
		t.Fatal("Could not add the admin RoleBinding")
	}

	cases := []struct {
		description         string
		method              string
		url                 string
		group               string         // Group the user belongs to
		attibutesMiddleware HTTPMiddleware // Legacy or Kubernetes-like routes
		expectedCode        int
	}{
		//
		// The cluster-admins group should grant all permissions on every resource
		// ClusterRoleBinding: cluster-admin (default)
		// ClusterRole: cluster-admin (default)
		//
		{
			description:         "cluster-admins can list users",
			method:              "GET",
			url:                 "/rbac/users",
			group:               "cluster-admins",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        200,
		},
		{
			description:         "cluster-admins can create users",
			method:              "POST",
			url:                 "/rbac/users",
			group:               "cluster-admins",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        200,
		},
		{
			description:         "cluster-admins can list ClusterRoles",
			method:              "GET",
			url:                 "/apis/rbac/v2/clusterroles",
			group:               "cluster-admins",
			attibutesMiddleware: AuthorizationAttributes{},
			expectedCode:        200,
		},
		{
			description:         "cluster-admins can create ClusterRoles",
			method:              "POST",
			url:                 "/apis/rbac/v2/clusterroles",
			group:               "cluster-admins",
			attibutesMiddleware: AuthorizationAttributes{},
			expectedCode:        200,
		},
		{
			description:         "cluster-admins can access checks in default namespace",
			method:              "GET",
			url:                 "/checks/check-cpu",
			group:               "cluster-admins",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        200,
		},
		{
			description:         "cluster-admins can access checks of any namespace",
			method:              "GET",
			url:                 "/checks/check-cpu?namespace=acme",
			group:               "cluster-admins",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        200,
		},
		//
		// The local-admins group should grant all permissions on all resources
		// contained within any namespace
		// ClusterRoleBinding: admin
		// ClusterRole: admin
		//
		{
			description:         "local-admins can't list ClusterRoles",
			method:              "GET",
			url:                 "/apis/rbac/v2/clusterroles",
			group:               "local-admins",
			attibutesMiddleware: AuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "local-admins can't create namespaces",
			method:              "POST",
			url:                 "/rbac/namespaces",
			group:               "local-admins",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "local-admins can list namespaces",
			method:              "GET",
			url:                 "/rbac/namespaces",
			group:               "local-admins",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        200,
		},
		{
			description:         "local-admins can access resource of any namespaces",
			method:              "GET",
			url:                 "/checks/check-cpu?namespace=acme",
			group:               "local-admins",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        200,
		},
		{
			description:         "local-admins can create RoleBindings",
			method:              "POST",
			url:                 "/apis/rbac/v2/namespaces/acme/rolebindings",
			group:               "local-admins",
			attibutesMiddleware: AuthorizationAttributes{},
			expectedCode:        200,
		},
		//
		// The admins group should grant all permissions on every resource within
		// the RoleBinding's namespace
		// RoleBinding: admin
		// ClusterRole: admin
		//
		{
			description:         "admins can't list ClusterRoles",
			method:              "GET",
			url:                 "/apis/rbac/v2/clusterroles",
			group:               "admins",
			attibutesMiddleware: AuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "admins can't create namespaces",
			method:              "POST",
			url:                 "/rbac/namespaces",
			group:               "admins",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "admins can't access resource of any namespaces",
			method:              "GET",
			url:                 "/checks/check-cpu?namespace=acme",
			group:               "admins",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "admins can list namespaces",
			method:              "GET",
			url:                 "/rbac/namespaces",
			group:               "admins",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        200,
		},
		{
			description:         "admins can create RoleBindings within their namespace",
			method:              "POST",
			url:                 "/apis/rbac/v2/namespaces/default/rolebindings",
			group:               "admins",
			attibutesMiddleware: AuthorizationAttributes{},
			expectedCode:        200,
		},
		{
			description:         "admins can access resource within their namespace",
			method:              "GET",
			url:                 "/checks/check-cpu?namespace=default",
			group:               "admins",
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

			// Inject the claims into the request context
			claims := types.Claims{
				StandardClaims: jwt.StandardClaims{Subject: "foo"},
				Groups:         []string{tt.group},
			}
			ctx := sensuJWT.SetClaimsIntoContext(r, &claims)

			// Prepare our middlewares
			namespaceMiddleware := Namespace{}
			attributesMiddleware := tt.attibutesMiddleware
			authorizationMiddleware := Authorization{Authorizer: &rbac.Authorizer{Store: store}}

			// Prepare the router
			router := mux.NewRouter()
			router.PathPrefix("/apis/{group}/{version}/namespaces/{namespace}/{kind}/{name}").Handler(testHandler)
			router.PathPrefix("/apis/{group}/{version}/namespaces/{namespace}/{kind}").Handler(testHandler)
			router.PathPrefix("/apis/{group}/{version}/{kind}/{name}").Handler(testHandler)
			router.PathPrefix("/apis/{group}/{version}/{kind}").Handler(testHandler)
			router.PathPrefix("/").Handler(testHandler) // catch all for legacy routes
			router.Use(namespaceMiddleware.Then, attributesMiddleware.Then, authorizationMiddleware.Then)

			// Serve the request
			router.ServeHTTP(w, r.WithContext(ctx))
			assert.Equal(t, tt.expectedCode, w.Code)
			if w.Body.Len() != 0 {
				t.Logf("Response body: %s", w.Body.String())
			}
		})
	}
}
