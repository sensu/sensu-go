package routers

import (
	"context"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/sensu/sensu-go/testing/mockauthorizer"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

func TestNamespacesRouter(t *testing.T) {
	// Setup the router
	s := &mockstore.MockStore{}
	clusterRole := corev2.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
		Rules: []corev2.Rule{
			{
				Verbs:     []string{corev2.VerbAll},
				Resources: []string{corev2.ResourceAll},
			},
		},
	}
	clusterRoleBinding := corev2.ClusterRoleBinding{
		Subjects: []corev2.Subject{
			{
				Type: "Group",
				Name: "cluster-admins",
			},
		},
		RoleRef: corev2.RoleRef{
			Type: "ClusterRole",
			Name: "cluster-admin",
		},
		ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
	}

	s.On("ListClusterRoleBindings", mock.Anything, mock.Anything).Return([]*corev2.ClusterRoleBinding{&clusterRoleBinding}, nil)
	s.On("GetClusterRole", mock.Anything, mock.Anything).Return(&clusterRole, nil)
	s.On("ListRoleBindings", mock.Anything, mock.Anything).Return([]*corev2.RoleBinding{}, nil)

	// Mock role & role binding creation, which happens upon namespace creation
	s.On("CreateOrUpdateResource", mock.Anything, mock.AnythingOfType("*v2.Role")).Return(nil)
	s.On("CreateOrUpdateResource", mock.Anything, mock.AnythingOfType("*v2.RoleBinding")).Return(nil)

	// Mock role & role binding deletion, which happens upon namespace deletion
	s.On("DeleteResource", mock.Anything, "rbac/rolebindings", "system:pipeline").Return(nil)
	s.On("DeleteResource", mock.Anything, "rbac/roles", "system:pipeline").Return(nil)

	// Mock an authorizer since the api client performs authorization on its own
	authorizer := &mockauthorizer.Authorizer{}
	authorizer.On("Authorize", mock.Anything, mock.Anything).Return(true, nil)

	s2 := new(mockstore.V2MockStore)
	s2.On("List", mock.Anything, mock.Anything).Return(wrap.List{}, nil)

	router := NewNamespacesRouter(s, s, authorizer, s2)
	parentRouter := mux.NewRouter().PathPrefix(corev2.URLPrefix).Subrouter()
	parentRouter.Use(mockedClaims)
	router.Mount(parentRouter)

	empty := &corev2.Namespace{}
	fixture := corev2.FixtureNamespace("foo")

	tests := []routerTestCase{}
	tests = append(tests, getTestCases(fixture)...)
	tests = append(tests, createTestCases(empty)...)
	tests = append(tests, updateTestCases(fixture)...)
	tests = append(tests, deleteTestCases(fixture)...)
	for _, tt := range tests {
		run(t, tt, parentRouter, s)
	}
}

func TestNamespaceRouterList(t *testing.T) {
	namespaces := []*corev2.Namespace{
		corev2.FixtureNamespace("default"),
	}
	clusterRole := corev2.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
		Rules: []corev2.Rule{
			{
				Verbs:     []string{corev2.VerbAll},
				Resources: []string{corev2.ResourceAll},
			},
		},
	}
	clusterRoleBinding := corev2.ClusterRoleBinding{
		Subjects: []corev2.Subject{
			{
				Type: "Group",
				Name: "cluster-admins",
			},
		},
		RoleRef: corev2.RoleRef{
			Type: "ClusterRole",
			Name: "cluster-admin",
		},
		ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
	}

	s := new(mockstore.MockStore)
	s.On("ListClusterRoleBindings", mock.Anything, mock.Anything).Return([]*corev2.ClusterRoleBinding{&clusterRoleBinding}, nil)
	s.On("GetClusterRole", mock.Anything, mock.Anything).Return(&clusterRole, nil)
	s.On("ListRoleBindings", mock.Anything, mock.Anything).Return([]*corev2.RoleBinding{}, nil)
	s.On("ListResources", mock.Anything, corev2.NamespacesResource, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		resources := args[2].(*[]*corev2.Namespace)
		*resources = append(*resources, namespaces...)

		pred := args[3].(*store.SelectionPredicate)
		*pred = store.SelectionPredicate{Continue: "sensu-continue"}
	}).Return(nil)

	ctx := context.Background()
	ctx = context.WithValue(ctx, corev2.ClaimsKey, corev2.FixtureClaims("foo", []string{"cluster-admins"}))

	auth := &rbac.Authorizer{Store: s}
	s2 := new(mockstore.V2MockStore)
	s2.On("List", mock.Anything, mock.Anything).Return(wrap.List{}, nil)
	router := NewNamespacesRouter(s, s, auth, s2)
	pred := &store.SelectionPredicate{Limit: 1}
	got, err := router.list(ctx, pred)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("expected namespaces to be returned")
	}
	if pred.Continue != "sensu-continue" {
		t.Fatalf("expected a continue token, got %q", pred.Continue)
	}
}

func mockedClaims(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), corev2.ClaimsKey, corev2.FixtureClaims("foo", []string{"cluster-admins"}))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
