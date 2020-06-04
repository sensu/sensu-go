package routers

import (
	"context"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

func TestNamespacesRouter(t *testing.T) {
	// Setup the router
	s := &mockstore.MockStore{}
	router := NewNamespacesRouter(s, nil)
	parentRouter := mux.NewRouter().PathPrefix(corev2.URLPrefix).Subrouter()
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

	store := new(mockstore.MockStore)
	store.On("ListClusterRoleBindings", mock.Anything, mock.Anything).Return([]*corev2.ClusterRoleBinding{&clusterRoleBinding}, nil)
	store.On("GetClusterRole", mock.Anything, mock.Anything).Return(&clusterRole, nil)
	store.On("ListRoleBindings", mock.Anything, mock.Anything).Return([]*corev2.RoleBinding{}, nil)
	store.On("ListResources", mock.Anything, corev2.NamespacesResource, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		resources := args[2].(*[]*corev2.Namespace)
		*resources = append(*resources, namespaces...)
	}).Return(nil)

	ctx := context.Background()
	ctx = context.WithValue(ctx, corev2.ClaimsKey, corev2.FixtureClaims("foo", []string{"cluster-admins"}))

	auth := &rbac.Authorizer{Store: store}
	router := NewNamespacesRouter(store, auth)
	got, err := router.list(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("expected namespaces to be returned")
	}
}
