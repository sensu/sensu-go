package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/middlewares"
	sensuJWT "github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/seeds"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/postgres"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func seedStore(t *testing.T, store storev2.Interface, nsStore storev2.NamespaceStore) {
	t.Helper()

	if err := seeds.SeedInitialDataWithContext(context.Background(), store, nsStore); err != nil {
		t.Fatalf("Could not seed the backend: %s", err)
	}

	// Add custom resources for the tests
	// Add a ClusterRoleBinding for the ClusteRole admin and assign the
	// local-admins group
	localAdmins := &corev2.ClusterRoleBinding{
		ObjectMeta: corev2.NewObjectMeta("admin", ""),
		RoleRef: corev2.RoleRef{
			Type: "ClusterRole",
			Name: "admin",
		},
		Subjects: []corev2.Subject{
			{
				Type: "Group",
				Name: "local-admins",
			},
		},
	}

	admins := &corev2.RoleBinding{
		ObjectMeta: corev2.NewObjectMeta("admin", "default"),
		RoleRef: corev2.RoleRef{
			Type: "ClusterRole",
			Name: "admin",
		},
		Subjects: []corev2.Subject{
			{
				Type: "Group",
				Name: "admins",
			},
		},
	}

	editors := &corev2.RoleBinding{
		ObjectMeta: corev2.NewObjectMeta("edit", "default"),
		RoleRef: corev2.RoleRef{
			Type: "ClusterRole",
			Name: "edit",
		},
		Subjects: []corev2.Subject{
			{
				Type: "Group",
				Name: "editors",
			},
		},
	}

	viewers := &corev2.RoleBinding{
		ObjectMeta: corev2.NewObjectMeta("view", "default"),
		RoleRef: corev2.RoleRef{
			Type: "ClusterRole",
			Name: "view",
		},
		Subjects: []corev2.Subject{
			{
				Type: "Group",
				Name: "viewers",
			},
		},
	}

	fooViewerRole := &corev2.Role{
		ObjectMeta: corev2.NewObjectMeta("foo-viewer", "default"),
		Rules: []corev2.Rule{
			{
				Verbs:         []string{"get"},
				Resources:     []string{"checks"},
				ResourceNames: []string{"foo"},
			},
		},
	}

	fooViewerRoleBinding := &corev2.RoleBinding{
		ObjectMeta: corev2.NewObjectMeta("foo-viewer", "default"),
		RoleRef: corev2.RoleRef{
			Type: "Role",
			Name: "foo-viewer",
		},
		Subjects: []corev2.Subject{
			{
				Type: "Group",
				Name: "foo-viewers",
			},
		},
	}

	rwRoleBinding := &corev2.RoleBinding{
		ObjectMeta: corev2.NewObjectMeta("rw", "default"),
		RoleRef: corev2.RoleRef{
			Type: "Role",
			Name: "rw",
		},
		Subjects: []corev2.Subject{
			{
				Type: "Group",
				Name: "rw",
			},
		},
	}

	rwRole := &corev2.Role{
		ObjectMeta: corev2.NewObjectMeta("rw", "default"),
		Rules: []corev2.Rule{
			{
				Verbs:     []string{"get", "list", "create", "update", "delete"},
				Resources: []string{"*"},
			},
		},
	}

	resources := []corev3.Resource{
		localAdmins,
		admins,
		editors,
		viewers,
		fooViewerRole,
		fooViewerRoleBinding,
		rwRoleBinding,
		rwRole,
	}

	for i, resource := range resources {
		req := storev2.NewResourceRequestFromResource(resource)
		wrapper, err := storev2.WrapResource(resource)
		if err != nil {
			t.Fatalf("error wrapping resource %d: %s", i, err)
		}
		if err := store.CreateIfNotExists(context.Background(), req, wrapper); err != nil {
			t.Fatalf("error creating resource %d: %s", i, err)
		}
	}
}

func TestAuthorization(t *testing.T) {
	t.Skip("skipping until further refactoring is complete")

	// If required, uncomment the following comment to enable the tracing logs
	logrus.SetLevel(logrus.TraceLevel)

	cases := []struct {
		description          string
		method               string
		url                  string
		group                string                     // Group the user belongs to
		attributesMiddleware middlewares.HTTPMiddleware // Legacy or Kubernetes-like routes
		expectedCode         int
	}{
		//
		// The cluster-admins group should grant all permissions on every resource
		// ClusterRoleBinding: cluster-admin (default)
		// ClusterRole: cluster-admin (default)
		//
		{
			description:          "cluster-admins can list users",
			method:               "GET",
			url:                  "/api/core/v2/users",
			group:                "cluster-admins",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "cluster-admins can create users",
			method:               "POST",
			url:                  "/api/core/v2/users",
			group:                "cluster-admins",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "cluster-admins can list ClusterRoles",
			method:               "GET",
			url:                  "/api/core/v2/clusterroles",
			group:                "cluster-admins",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "cluster-admins can create ClusterRoles",
			method:               "POST",
			url:                  "/api/core/v2/clusterroles",
			group:                "cluster-admins",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "cluster-admins can access checks in default namespace",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/default/checks/check-cpu",
			group:                "cluster-admins",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "cluster-admins can access checks of any namespace",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/acme/checks/check-cpu",
			group:                "cluster-admins",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         200,
		},
		//
		// The local-admins group should grant all permissions on all resources
		// contained within any namespace
		// ClusterRoleBinding: admin
		// ClusterRole: admin
		//
		{
			description:          "local-admins can't list ClusterRoles",
			method:               "GET",
			url:                  "/api/core/v2/clusterroles",
			group:                "local-admins",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "local-admins can't create namespaces",
			method:               "POST",
			url:                  "/api/core/v2/namespaces",
			group:                "local-admins",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "local-admins can list namespaces",
			method:               "GET",
			url:                  "/api/core/v2/namespaces",
			group:                "local-admins",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "local-admins can access resource of any namespaces",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/acme/checks/check-cpu",
			group:                "local-admins",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "local-admins can create RoleBindings",
			method:               "POST",
			url:                  "/api/core/v2/namespaces/acme/rolebindings",
			group:                "local-admins",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         200,
		},
		//
		// The admins group should grant all permissions on every resource within
		// the RoleBinding's namespace
		// RoleBinding: admin
		// ClusterRole: admin
		//
		{
			description:          "admins can't list ClusterRoles",
			method:               "GET",
			url:                  "/api/core/v2/clusterroles",
			group:                "admins",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "admins can't create namespaces",
			method:               "POST",
			url:                  "/api/core/v2/namespaces",
			group:                "admins",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "admins can't access resource outside of their namespaces",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/acme/checks/check-cpu",
			group:                "admins",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "admins can create RoleBindings within their namespace",
			method:               "POST",
			url:                  "/api/core/v2/namespaces/default/rolebindings",
			group:                "admins",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "admins can access resource within their namespace",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/default/checks/check-cpu",
			group:                "admins",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         200,
		},
		//
		// The editors group should grant read/write access to most objects in the
		// RoleBinding's namespace, expected for Roles or RoleBindings RoleBinding:
		// edit ClusterRole: edit
		//
		{
			description:          "editors can't list ClusterRoles",
			method:               "GET",
			url:                  "/api/core/v2/clusterroles",
			group:                "editors",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "editors can't create namespaces",
			method:               "POST",
			url:                  "/api/core/v2/namespaces",
			group:                "editors",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "editors can't access resource outside of their namespaces",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/acme/checks/check-cpu",
			group:                "editors",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "editors can't create RoleBindings within their namespace",
			method:               "POST",
			url:                  "/api/core/v2/namespaces/default/rolebindings",
			group:                "editors",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "editors can access resource within their namespace",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/default/checks/check-cpu",
			group:                "editors",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         200,
		},
		//
		// The viewers group only grant read access to most objects in the
		// RoleBinding's namespace, expected for Roles or RoleBindings
		// RoleBinding: view
		// ClusterRole: view
		//
		{
			description:          "viewers can't list ClusterRoles",
			method:               "GET",
			url:                  "/api/core/v2/clusterroles",
			group:                "viewers",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "viewers can't create namespaces",
			method:               "POST",
			url:                  "/api/core/v2/namespaces",
			group:                "viewers",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "viewers can't access resource outside of their namespaces",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/acme/checks/check-cpu",
			group:                "viewers",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "viewers can't create RoleBindings within their namespace",
			method:               "POST",
			url:                  "/api/core/v2/namespaces/default/rolebindings",
			group:                "viewers",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "viewers can't create resources within their namespace",
			method:               "PUT",
			url:                  "/api/core/v2/namespaces/default/checks/check-cpu",
			group:                "viewers",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "viewers can access resource within their namespace",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/default/checks/check-cpu",
			group:                "viewers",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         200,
		},
		//
		// The system:agents group only grant read/write access to events
		// RoleBinding: system:agent
		// ClusterRole: system:agent
		//
		{
			description:          "system:agents can't list ClusterRoles",
			method:               "GET",
			url:                  "/api/core/v2/clusterroles",
			group:                "system:agents",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "system:agents can't create namespaces",
			method:               "POST",
			url:                  "/api/core/v2/namespaces",
			group:                "system:agents",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "system:agents can't access resource outside of their namespaces",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/acme/checks/check-cpu",
			group:                "system:agents",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "system:agents can't create RoleBindings within their namespace",
			method:               "POST",
			url:                  "/api/core/v2/namespaces/default/rolebindings",
			group:                "system:agents",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "system:agents can't create any resources within their namespace",
			method:               "PUT",
			url:                  "/api/core/v2/namespaces/default/checks/check-cpu",
			group:                "system:agents",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "system:agents can list namespaces",
			method:               "GET",
			url:                  "/api/core/v2/namespaces",
			group:                "system:agents",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "system:agents can create events",
			method:               "POST",
			url:                  "/api/core/v2/namespaces/default/events",
			group:                "system:agents",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         200,
		},
		//
		// The foo-viewers group only grant read access to a check named 'foo'
		// RoleBinding: foo-viewer
		// ClusterRole: foo-viewer
		//
		{
			description:          "foo-viewers can't get an event",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/default/events/foo/bar",
			group:                "foo-viewers",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "foo-viewers can't list checks",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/default/checks",
			group:                "foo-viewers",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "foo-viewers can't update the foo check",
			method:               "PUT",
			url:                  "/api/core/v2/namespaces/default/checks/foo",
			group:                "foo-viewers",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "foo-viewers can view the foo check",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/default/checks/foo",
			group:                "foo-viewers",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         200,
		},
		//
		// The system:users only grant the user access to view itself and update its
		// password
		// RoleBinding: system:user
		// ClusterRole: system:user
		//
		{
			description:          "system:users can't view another user",
			method:               "GET",
			url:                  "/api/core/v2/users/bar",
			group:                "system:users",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "system:users can't modify another user password",
			method:               "PUT",
			url:                  "/api/core/v2/users/bar/password",
			group:                "system:users",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "system:users can view themselves",
			method:               "GET",
			url:                  "/api/core/v2/users/foo",
			group:                "system:users",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "system:users can modify their own user password",
			method:               "PUT",
			url:                  "/api/core/v2/users/foo/password",
			group:                "system:users",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         200,
		},
		//
		// A user with explicit permission on all verbs is able to PATCH resources
		//
		{
			description:          "the update verb grants permission to PATCH",
			method:               "PATCH",
			url:                  "/api/core/v2/namespaces/default/checks/foo",
			group:                "rw",
			attributesMiddleware: middlewares.AuthorizationAttributes{},
			expectedCode:         200,
		},
	}
	postgres.WithPostgres(t, func(ctx context.Context, configDB *pgxpool.Pool, dsn string) {
		postgres.WithPostgres(t, func(ctx context.Context, stateDB *pgxpool.Pool, dsn string) {
			configStore := postgres.NewConfigStore(configDB, stateDB)
			namespaceStore := postgres.NewNamespaceStore(stateDB)
			seedStore(t, configStore, namespaceStore)

			for _, tt := range cases {
				t.Run(tt.description, func(t *testing.T) {
					// testHandler is a catch-all handler that returns 200 OK
					testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

					// Prepare our HTTP server
					w := httptest.NewRecorder()

					// Prepare the request
					r, err := http.NewRequest(tt.method, tt.url, nil)
					if err != nil {
						t.Fatal("Couldn't create request: ", err)
					}

					// Inject the claims into the request context
					claims := corev2.Claims{
						StandardClaims: jwt.StandardClaims{Subject: "foo"},
						Groups:         []string{tt.group},
					}
					ctx := sensuJWT.SetClaimsIntoContext(r, &claims)

					// Prepare our middlewares
					namespaceMiddleware := middlewares.Namespace{}
					attributesMiddleware := tt.attributesMiddleware
					authorizationMiddleware := middlewares.Authorization{Authorizer: &rbac.Authorizer{Store: configStore}}

					// Prepare the router
					router := mux.NewRouter()
					router.PathPrefix("/api/{group}/{version}/{resource:users}/{id}/{subresource}").Handler(testHandler)
					router.PathPrefix("/api/{group}/{version}/namespaces/{namespace}/{resource}/{id}").Handler(testHandler)
					router.PathPrefix("/api/{group}/{version}/namespaces/{namespace}/{resource}").Handler(testHandler)
					router.PathPrefix("/api/{group}/{version}/{resource}/{id}").Handler(testHandler)
					router.PathPrefix("/api/{group}/{version}/{resource}").Handler(testHandler)
					router.PathPrefix("/api/{group}/{version}/{resource:namespaces}/{id}").Handler(testHandler)
					router.PathPrefix("/api/{group}/{version}/{resource:namespaces}").Handler(testHandler)
					router.PathPrefix("/api/{group}/{version}/{resource:users}/{id}").Handler(testHandler)
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
		})
	})

}

func getFaultyRoleBinding() *corev2.RoleBinding {
	return &corev2.RoleBinding{
		Subjects: []corev2.Subject{
			{
				Type: "Group",
				Name: "admins",
			},
		},
		RoleRef: corev2.RoleRef{
			Type: "Role",
			Name: "doesnotexist",
		},
		ObjectMeta: corev2.ObjectMeta{
			Name:      "myrolebinding",
			Namespace: "default",
		},
	}
}

func TestRoleNotFound_GH4268(t *testing.T) {
	st := new(mockstore.V2MockStore)
	faultyRoleBindings := []*corev2.RoleBinding{getFaultyRoleBinding()}
	rb := &corev2.RoleBinding{}
	rb.Namespace = "default"
	rbListReq := storev2.NewResourceRequestFromResource(rb)
	st.On("List", mock.Anything, rbListReq, mock.Anything).Return(mockstore.WrapList[*corev2.RoleBinding](faultyRoleBindings), nil)
	crbListReq := storev2.NewResourceRequestFromResource(new(corev2.ClusterRoleBinding))
	st.On("List", mock.Anything, crbListReq, mock.Anything).Return(mockstore.WrapList[*corev2.ClusterRoleBinding](nil), nil)
	st.On("Get", mock.Anything, mock.Anything).Return(nil, &store.ErrNotFound{})

	// testHandler is a catch-all handler that returns 200 OK
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	// Prepare our HTTP server
	w := httptest.NewRecorder()

	// Prepare the request
	r, err := http.NewRequest("GET", "/api/core/v2/namespaces/default/checks/foo", nil)
	if err != nil {
		t.Fatal("Couldn't create request: ", err)
	}

	// Inject the claims into the request context
	claims := corev2.Claims{
		StandardClaims: jwt.StandardClaims{Subject: "foo"},
		Groups:         []string{"admins"},
	}
	ctx := sensuJWT.SetClaimsIntoContext(r, &claims)

	// Prepare our middlewares
	namespaceMiddleware := middlewares.Namespace{}
	attributesMiddleware := middlewares.AuthorizationAttributes{}
	authorizationMiddleware := middlewares.Authorization{Authorizer: &rbac.Authorizer{Store: st}}

	// Prepare the router
	router := mux.NewRouter()
	router.PathPrefix("/api/{group}/{version}/{resource:users}/{id}/{subresource}").Handler(testHandler)
	router.PathPrefix("/api/{group}/{version}/namespaces/{namespace}/{resource}/{id}").Handler(testHandler)
	router.PathPrefix("/api/{group}/{version}/namespaces/{namespace}/{resource}").Handler(testHandler)
	router.PathPrefix("/api/{group}/{version}/{resource}/{id}").Handler(testHandler)
	router.PathPrefix("/api/{group}/{version}/{resource}").Handler(testHandler)
	router.PathPrefix("/api/{group}/{version}/{resource:namespaces}/{id}").Handler(testHandler)
	router.PathPrefix("/api/{group}/{version}/{resource:namespaces}").Handler(testHandler)
	router.PathPrefix("/api/{group}/{version}/{resource:users}/{id}").Handler(testHandler)
	router.PathPrefix("/").Handler(testHandler) // catch all for legacy routes
	router.Use(namespaceMiddleware.Then, attributesMiddleware.Then, authorizationMiddleware.Then)

	// Serve the request
	router.ServeHTTP(w, r.WithContext(ctx))
	if got, want := w.Code, http.StatusForbidden; got != want {
		t.Errorf("bad status: got %d, want %d", got, want)
	}
	if !strings.Contains(w.Body.String(), "role not found") {
		t.Error(w.Body.String())
	}
}
