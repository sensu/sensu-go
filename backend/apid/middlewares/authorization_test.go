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
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func seedStore(t *testing.T, store store.Store) {
	t.Helper()

	if err := seeds.SeedInitialData(store); err != nil {
		t.Fatal("Could not seed the backend: ", err)
	}

	// Add custom resources for the tests
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

	editors := &types.RoleBinding{
		Name:      "edit",
		Namespace: "default",
		RoleRef: types.RoleRef{
			Kind: "ClusterRole",
			Name: "edit",
		},
		Subjects: []types.Subject{
			types.Subject{
				Kind: "Group",
				Name: "editors",
			},
		},
	}
	if err := store.CreateRoleBinding(context.Background(), editors); err != nil {
		t.Fatal("Could not add the edit RoleBinding")
	}

	viewers := &types.RoleBinding{
		Name:      "view",
		Namespace: "default",
		RoleRef: types.RoleRef{
			Kind: "ClusterRole",
			Name: "view",
		},
		Subjects: []types.Subject{
			types.Subject{
				Kind: "Group",
				Name: "viewers",
			},
		},
	}
	if err := store.CreateRoleBinding(context.Background(), viewers); err != nil {
		t.Fatal("Could not add the view RoleBinding")
	}

	fooViewerRole := &types.Role{
		Name:      "foo-viewer",
		Namespace: "default",
		Rules: []types.Rule{
			types.Rule{
				Verbs:         []string{"get"},
				Resources:     []string{"checks"},
				ResourceNames: []string{"foo"},
			},
		},
	}
	if err := store.CreateRole(context.Background(), fooViewerRole); err != nil {
		t.Fatal("Could not add the foo-viewer RoleBinding")
	}

	fooViewerRoleBinding := &types.RoleBinding{
		Name:      "foo-viewer",
		Namespace: "default",
		RoleRef: types.RoleRef{
			Kind: "Role",
			Name: "foo-viewer",
		},
		Subjects: []types.Subject{
			types.Subject{
				Kind: "Group",
				Name: "foo-viewers",
			},
		},
	}
	if err := store.CreateRoleBinding(context.Background(), fooViewerRoleBinding); err != nil {
		t.Fatal("Could not add the foo-viewer RoleBinding")
	}
}

func TestAuthorization(t *testing.T) {
	// If required, uncomment the following comment to enable the tracing logs
	logrus.SetLevel(logrus.TraceLevel)

	// Prepare the store
	// Use the default seeds
	store, err := testutil.NewStoreInstance()
	if err != nil {
		t.Fatal("Could not initialize the store: ", err)
	}
	seedStore(t, store)

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
			description:         "admins can't access resource outside of their namespaces",
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
		//
		// The editors group should grant read/write access to most objects in the
		// RoleBinding's namespace, expected for Roles or RoleBindings RoleBinding:
		// edit ClusterRole: edit
		//
		{
			description:         "editors can't list ClusterRoles",
			method:              "GET",
			url:                 "/apis/rbac/v2/clusterroles",
			group:               "editors",
			attibutesMiddleware: AuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "editors can't create namespaces",
			method:              "POST",
			url:                 "/rbac/namespaces",
			group:               "editors",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "editors can't access resource outside of their namespaces",
			method:              "GET",
			url:                 "/checks/check-cpu?namespace=acme",
			group:               "editors",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "editors can't create RoleBindings within their namespace",
			method:              "POST",
			url:                 "/apis/rbac/v2/namespaces/default/rolebindings",
			group:               "editors",
			attibutesMiddleware: AuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "editors can list namespaces",
			method:              "GET",
			url:                 "/rbac/namespaces",
			group:               "editors",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        200,
		},
		{
			description:         "editors can access resource within their namespace",
			method:              "GET",
			url:                 "/checks/check-cpu?namespace=default",
			group:               "editors",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        200,
		},
		//
		// The viewers group only grant read access to most objects in the
		// RoleBinding's namespace, expected for Roles or RoleBindings
		// RoleBinding: view
		// ClusterRole: view
		//
		{
			description:         "viewers can't list ClusterRoles",
			method:              "GET",
			url:                 "/apis/rbac/v2/clusterroles",
			group:               "viewers",
			attibutesMiddleware: AuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "viewers can't create namespaces",
			method:              "POST",
			url:                 "/rbac/namespaces",
			group:               "viewers",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "viewers can't access resource outside of their namespaces",
			method:              "GET",
			url:                 "/checks/check-cpu?namespace=acme",
			group:               "viewers",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "viewers can't create RoleBindings within their namespace",
			method:              "POST",
			url:                 "/apis/rbac/v2/namespaces/default/rolebindings",
			group:               "viewers",
			attibutesMiddleware: AuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "viewers can't create resources within their namespace",
			method:              "PUT",
			url:                 "/checks/check-cpu?namespace=default",
			group:               "viewers",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "viewers can list namespaces",
			method:              "GET",
			url:                 "/rbac/namespaces",
			group:               "viewers",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        200,
		},
		{
			description:         "viewers can access resource within their namespace",
			method:              "GET",
			url:                 "/checks/check-cpu?namespace=default",
			group:               "viewers",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        200,
		},
		//
		// The system:agents group only grant read/write access to events
		// RoleBinding: system:agent
		// ClusterRole: system:agent
		//
		{
			description:         "system:agents can't list ClusterRoles",
			method:              "GET",
			url:                 "/apis/rbac/v2/clusterroles",
			group:               "system:agents",
			attibutesMiddleware: AuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "system:agents can't create namespaces",
			method:              "POST",
			url:                 "/rbac/namespaces",
			group:               "system:agents",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "system:agents can't access resource outside of their namespaces",
			method:              "GET",
			url:                 "/checks/check-cpu?namespace=acme",
			group:               "system:agents",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "system:agents can't create RoleBindings within their namespace",
			method:              "POST",
			url:                 "/apis/rbac/v2/namespaces/default/rolebindings",
			group:               "system:agents",
			attibutesMiddleware: AuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "system:agents can't create any resources within their namespace",
			method:              "PUT",
			url:                 "/checks/check-cpu?namespace=default",
			group:               "system:agents",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "system:agents can't list namespaces",
			method:              "GET",
			url:                 "/rbac/namespaces",
			group:               "system:agents",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "system:agents can create events",
			method:              "POST",
			url:                 "/events",
			group:               "system:agents",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        200,
		},
		//
		// The foo-viewers group only grant read access to a check named 'foo'
		// RoleBinding: foo-viewer
		// ClusterRole: foo-viewer
		//
		{
			description:         "foo-viewers can't get an event",
			method:              "GET",
			url:                 "/events/foo/bar",
			group:               "foo-viewers",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "foo-viewers can't list checks",
			method:              "GET",
			url:                 "/checks",
			group:               "foo-viewers",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "foo-viewers can't update the foo check",
			method:              "PUT",
			url:                 "/checks/foo",
			group:               "foo-viewers",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "foo-viewers can view the foo check",
			method:              "GET",
			url:                 "/checks/foo",
			group:               "foo-viewers",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        200,
		},
		//
		// The system:users only grant the user access to view itself and update its
		// password
		// RoleBinding: system:user
		// ClusterRole: system:user
		//
		{
			description:         "system:users can't view another user",
			method:              "GET",
			url:                 "/rbac/users/bar",
			group:               "system:users",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "system:users can't modify another user password",
			method:              "PUT",
			url:                 "/rbac/users/bar/password",
			group:               "system:users",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        403,
		},
		{
			description:         "system:users can view themselves",
			method:              "GET",
			url:                 "/rbac/users/foo",
			group:               "system:users",
			attibutesMiddleware: LegacyAuthorizationAttributes{},
			expectedCode:        200,
		},
		{
			description:         "system:users can modify their own user password",
			method:              "PUT",
			url:                 "/rbac/users/foo/password",
			group:               "system:users",
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
			router.PathPrefix("/{prefix:rbac}/{resource}/{id}/{subresource}").Handler(testHandler)
			router.PathPrefix("/{prefix:rbac}/{resource}/{id}").Handler(testHandler)
			router.PathPrefix("/{kind}/{id}").Handler(testHandler)
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
