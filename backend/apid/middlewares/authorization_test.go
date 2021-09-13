package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	sensuJWT "github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/seeds"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd/testutil"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func seedStore(t *testing.T, store store.Store) {
	t.Helper()

	if err := seeds.SeedInitialData(store); err != nil {
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
	if err := store.CreateClusterRoleBinding(context.Background(), localAdmins); err != nil {
		t.Fatalf("Could not add the admin ClusterRoleBinding: %s", err)
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
	if err := store.CreateRoleBinding(context.Background(), admins); err != nil {
		t.Fatalf("Could not add the admin RoleBinding: %s", err)
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
	if err := store.CreateRoleBinding(context.Background(), editors); err != nil {
		t.Fatalf("Could not add the edit RoleBinding: %s", err)
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
	if err := store.CreateRoleBinding(context.Background(), viewers); err != nil {
		t.Fatalf("Could not add the view RoleBinding: %s", err)
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
	if err := store.CreateRole(context.Background(), fooViewerRole); err != nil {
		t.Fatalf("Could not add the foo-viewer Role: %s", err)
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
	if err := store.CreateRoleBinding(context.Background(), fooViewerRoleBinding); err != nil {
		t.Fatalf("Could not add the foo-viewer RoleBinding: %s", err)
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
	if err := store.CreateRoleBinding(context.Background(), rwRoleBinding); err != nil {
		t.Fatalf("Could not add the rw RoleBinding: %s", err)
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
	if err := store.CreateRole(context.Background(), rwRole); err != nil {
		t.Fatalf("Could not add the rw Role: %s", err)
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
		description          string
		method               string
		url                  string
		group                string         // Group the user belongs to
		attributesMiddleware HTTPMiddleware // Legacy or Kubernetes-like routes
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
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "cluster-admins can create users",
			method:               "POST",
			url:                  "/api/core/v2/users",
			group:                "cluster-admins",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "cluster-admins can list ClusterRoles",
			method:               "GET",
			url:                  "/api/core/v2/clusterroles",
			group:                "cluster-admins",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "cluster-admins can create ClusterRoles",
			method:               "POST",
			url:                  "/api/core/v2/clusterroles",
			group:                "cluster-admins",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "cluster-admins can access checks in default namespace",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/default/checks/check-cpu",
			group:                "cluster-admins",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "cluster-admins can access checks of any namespace",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/acme/checks/check-cpu",
			group:                "cluster-admins",
			attributesMiddleware: AuthorizationAttributes{},
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
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "local-admins can't create namespaces",
			method:               "POST",
			url:                  "/api/core/v2/namespaces",
			group:                "local-admins",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "local-admins can list namespaces",
			method:               "GET",
			url:                  "/api/core/v2/namespaces",
			group:                "local-admins",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "local-admins can access resource of any namespaces",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/acme/checks/check-cpu",
			group:                "local-admins",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "local-admins can create RoleBindings",
			method:               "POST",
			url:                  "/api/core/v2/namespaces/acme/rolebindings",
			group:                "local-admins",
			attributesMiddleware: AuthorizationAttributes{},
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
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "admins can't create namespaces",
			method:               "POST",
			url:                  "/api/core/v2/namespaces",
			group:                "admins",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "admins can't access resource outside of their namespaces",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/acme/checks/check-cpu",
			group:                "admins",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "admins can create RoleBindings within their namespace",
			method:               "POST",
			url:                  "/api/core/v2/namespaces/default/rolebindings",
			group:                "admins",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "admins can access resource within their namespace",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/default/checks/check-cpu",
			group:                "admins",
			attributesMiddleware: AuthorizationAttributes{},
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
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "editors can't create namespaces",
			method:               "POST",
			url:                  "/api/core/v2/namespaces",
			group:                "editors",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "editors can't access resource outside of their namespaces",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/acme/checks/check-cpu",
			group:                "editors",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "editors can't create RoleBindings within their namespace",
			method:               "POST",
			url:                  "/api/core/v2/namespaces/default/rolebindings",
			group:                "editors",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "editors can access resource within their namespace",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/default/checks/check-cpu",
			group:                "editors",
			attributesMiddleware: AuthorizationAttributes{},
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
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "viewers can't create namespaces",
			method:               "POST",
			url:                  "/api/core/v2/namespaces",
			group:                "viewers",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "viewers can't access resource outside of their namespaces",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/acme/checks/check-cpu",
			group:                "viewers",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "viewers can't create RoleBindings within their namespace",
			method:               "POST",
			url:                  "/api/core/v2/namespaces/default/rolebindings",
			group:                "viewers",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "viewers can't create resources within their namespace",
			method:               "PUT",
			url:                  "/api/core/v2/namespaces/default/checks/check-cpu",
			group:                "viewers",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "viewers can access resource within their namespace",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/default/checks/check-cpu",
			group:                "viewers",
			attributesMiddleware: AuthorizationAttributes{},
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
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "system:agents can't create namespaces",
			method:               "POST",
			url:                  "/api/core/v2/namespaces",
			group:                "system:agents",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "system:agents can't access resource outside of their namespaces",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/acme/checks/check-cpu",
			group:                "system:agents",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "system:agents can't create RoleBindings within their namespace",
			method:               "POST",
			url:                  "/api/core/v2/namespaces/default/rolebindings",
			group:                "system:agents",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "system:agents can't create any resources within their namespace",
			method:               "PUT",
			url:                  "/api/core/v2/namespaces/default/checks/check-cpu",
			group:                "system:agents",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "system:agents can list namespaces",
			method:               "GET",
			url:                  "/api/core/v2/namespaces",
			group:                "system:agents",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "system:agents can create events",
			method:               "POST",
			url:                  "/api/core/v2/namespaces/default/events",
			group:                "system:agents",
			attributesMiddleware: AuthorizationAttributes{},
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
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "foo-viewers can't list checks",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/default/checks",
			group:                "foo-viewers",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "foo-viewers can't update the foo check",
			method:               "PUT",
			url:                  "/api/core/v2/namespaces/default/checks/foo",
			group:                "foo-viewers",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "foo-viewers can view the foo check",
			method:               "GET",
			url:                  "/api/core/v2/namespaces/default/checks/foo",
			group:                "foo-viewers",
			attributesMiddleware: AuthorizationAttributes{},
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
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "system:users can't modify another user password",
			method:               "PUT",
			url:                  "/api/core/v2/users/bar/password",
			group:                "system:users",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         403,
		},
		{
			description:          "system:users can view themselves",
			method:               "GET",
			url:                  "/api/core/v2/users/foo",
			group:                "system:users",
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         200,
		},
		{
			description:          "system:users can modify their own user password",
			method:               "PUT",
			url:                  "/api/core/v2/users/foo/password",
			group:                "system:users",
			attributesMiddleware: AuthorizationAttributes{},
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
			attributesMiddleware: AuthorizationAttributes{},
			expectedCode:         200,
		},
	}
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
			namespaceMiddleware := Namespace{}
			attributesMiddleware := tt.attributesMiddleware
			authorizationMiddleware := Authorization{Authorizer: &rbac.Authorizer{Store: store}}

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
	store := new(mockstore.MockStore)
	faultyRoleBindings := []*corev2.RoleBinding{getFaultyRoleBinding()}
	store.On("ListRoleBindings", mock.Anything, mock.Anything).Return(faultyRoleBindings, nil)
	store.On("ListClusterRoleBindings", mock.Anything, mock.Anything).Return([]*corev2.ClusterRoleBinding{}, nil)
	store.On("GetRole", mock.Anything, mock.Anything).Return((*corev2.Role)(nil), nil)

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
	namespaceMiddleware := Namespace{}
	attributesMiddleware := AuthorizationAttributes{}
	authorizationMiddleware := Authorization{Authorizer: &rbac.Authorizer{Store: store}}

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
