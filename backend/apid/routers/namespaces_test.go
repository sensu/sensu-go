package routers

import (
	"context"
	"net/http"
	"reflect"
	"sort"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
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
	tests = append(tests, listTestCases(empty)...)
	tests = append(tests, createTestCases(empty)...)
	tests = append(tests, updateTestCases(fixture)...)
	tests = append(tests, deleteTestCases(fixture)...)
	for _, tt := range tests {
		run(t, tt, parentRouter, s)
	}
}

func TestNamespaceRouterGet(t *testing.T) {
	namespaces := []*corev2.Namespace{
		corev2.FixtureNamespace("a"),
		corev2.FixtureNamespace("b"),
		corev2.FixtureNamespace("c"),
		corev2.FixtureNamespace("d"),
		corev2.FixtureNamespace("e"),
		corev2.FixtureNamespace("f"),
	}
	tests := []struct {
		Name                string
		Attrs               *authorization.Attributes
		ClusterRoles        []*corev2.ClusterRole
		ClusterRoleBindings []*corev2.ClusterRoleBinding
		Roles               []*corev2.Role
		RoleBindings        []*corev2.RoleBinding
		AllNamespaces       []*corev2.Namespace
		ExpNamespaces       []*corev2.Namespace
		ExpError            bool
	}{
		{
			Name: "all access",
			Attrs: &authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "",
				Resource:     corev2.NamespacesResource,
				ResourceName: "",
				User: corev2.User{
					Username: "admin",
					Groups:   []string{"cluster-admins"},
				},
			},
			ClusterRoles: []*corev2.ClusterRole{
				{
					ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
					Rules: []corev2.Rule{
						{
							Verbs:     []string{corev2.VerbAll},
							Resources: []string{corev2.ResourceAll},
						},
					},
				},
			},
			ClusterRoleBindings: []*corev2.ClusterRoleBinding{
				{
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
				},
			},
			RoleBindings:  []*corev2.RoleBinding{},
			AllNamespaces: namespaces,
			ExpNamespaces: namespaces,
		},
		{
			Name: "no access",
			Attrs: &authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     corev2.NamespacesResource,
				ResourceName: "",
				User: corev2.User{
					Username: "regular-user",
					Groups:   []string{"plebs"},
				},
			},
			ClusterRoles: []*corev2.ClusterRole{
				{
					ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
					Rules: []corev2.Rule{
						{
							Verbs:     []string{corev2.VerbAll},
							Resources: []string{corev2.ResourceAll},
						},
					},
				},
			},
			ClusterRoleBindings: []*corev2.ClusterRoleBinding{
				{
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
				},
			},
			RoleBindings:  []*corev2.RoleBinding{},
			AllNamespaces: namespaces,
			ExpNamespaces: []*corev2.Namespace{},
		},
		{
			Name: "partial access",
			Attrs: &authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     corev2.NamespacesResource,
				ResourceName: "",
				User: corev2.User{
					Username: "regular-user",
					Groups:   []string{"plebs"},
				},
			},
			ClusterRoles: []*corev2.ClusterRole{
				{
					ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
					Rules: []corev2.Rule{
						{
							Verbs:     []string{corev2.VerbAll},
							Resources: []string{corev2.ResourceAll},
						},
					},
				},
			},
			ClusterRoleBindings: []*corev2.ClusterRoleBinding{
				{
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
				},
			},
			Roles: []*corev2.Role{
				{
					ObjectMeta: corev2.NewObjectMeta("pleb", "default"),
					Rules: []corev2.Rule{
						{
							Verbs:         []string{"get"},
							Resources:     []string{corev2.NamespacesResource},
							ResourceNames: []string{"a", "c", "e"},
						},
					},
				},
			},
			RoleBindings: []*corev2.RoleBinding{
				{
					Subjects: []corev2.Subject{
						{
							Type: "Group",
							Name: "plebs",
						},
					},
					RoleRef: corev2.RoleRef{
						Type: "Role",
						Name: "pleb",
					},
					ObjectMeta: corev2.NewObjectMeta("pleb", "default"),
				},
			},
			AllNamespaces: namespaces,
			ExpNamespaces: []*corev2.Namespace{
				namespaces[0],
				namespaces[2],
				namespaces[4],
			},
		},
		{
			Name: "implicit access via resources in namespace",
			Attrs: &authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "a",
				Resource:     corev2.ChecksResource,
				ResourceName: "",
				User: corev2.User{
					Username: "regular-user",
					Groups:   []string{"plebs"},
				},
			},
			ClusterRoles: []*corev2.ClusterRole{
				{
					ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
					Rules: []corev2.Rule{
						{
							Verbs:     []string{corev2.VerbAll},
							Resources: []string{corev2.ResourceAll},
						},
					},
				},
			},
			ClusterRoleBindings: []*corev2.ClusterRoleBinding{
				{
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
				},
			},
			Roles: []*corev2.Role{
				{
					ObjectMeta: corev2.NewObjectMeta("pleb", "a"),
					Rules: []corev2.Rule{
						{
							Verbs:     []string{"get"},
							Resources: []string{corev2.ChecksResource},
						},
					},
				},
			},
			RoleBindings: []*corev2.RoleBinding{
				{
					Subjects: []corev2.Subject{
						{
							Type: "Group",
							Name: "plebs",
						},
					},
					RoleRef: corev2.RoleRef{
						Type: "Role",
						Name: "pleb",
					},
					ObjectMeta: corev2.NewObjectMeta("pleb", "a"),
				},
			},
			AllNamespaces: namespaces,
			ExpNamespaces: []*corev2.Namespace{
				namespaces[0],
			},
		},
	}

	for _, test := range tests {
		store := new(mockstore.MockStore)
		store.On("ListClusterRoles", mock.Anything, mock.Anything).Return(test.ClusterRoles, nil)
		store.On("ListClusterRoleBindings", mock.Anything, mock.Anything).Return(test.ClusterRoleBindings, nil)
		store.On("ListRoles", mock.Anything, mock.Anything).Return(test.Roles, nil)
		store.On("ListRoleBindings", mock.Anything, mock.Anything).Return(test.RoleBindings, nil)
		store.On("ListResources", mock.Anything, corev2.NamespacesResource, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			resources := args[2].(*[]*corev2.Namespace)
			*resources = append(*resources, test.AllNamespaces...)
		}).Return(nil)
		setupGetClusterRoleAndGetRole(store, test.ClusterRoles, test.Roles)

		ctx := authorization.SetAttributes(context.Background(), test.Attrs)

		// The path doesn't matter as we really only are extracting the context from
		// the request here.
		req, err := http.NewRequest("GET", "/asdf", nil)
		if err != nil {
			t.Fatal(err)
		}
		req = req.WithContext(ctx)

		auth := &rbac.Authorizer{Store: store}
		router := NewNamespacesRouter(store, auth)

		got, err := router.get(req)
		if err != nil {
			t.Fatal(err)
		}

		sort.Slice(got, sortFunc(got.([]*corev2.Namespace)))

		if got, want := got, test.ExpNamespaces; !reflect.DeepEqual(got, want) {
			t.Fatalf("bad namespaces: got %+v, want %+v", got, want)
		}
	}
}

func sortFunc(namespaces []*corev2.Namespace) func(i, j int) bool {
	return func(i, j int) bool {
		return namespaces[i].Name < namespaces[j].Name
	}
}

func setupGetClusterRoleAndGetRole(store *mockstore.MockStore, clusterRoles []*corev2.ClusterRole, roles []*corev2.Role) {
	for _, role := range clusterRoles {
		store.On("GetClusterRole", mock.Anything, role.Name).Return(role, nil)
	}

	for _, role := range roles {
		store.On("GetRole", mock.Anything, role.Name).Return(role, nil)
	}
}
