package api

import (
	"reflect"
	"sort"
	"testing"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestFetchNamespace(t *testing.T) {
	tests := []struct {
		name                string
		namespace           string
		attrs               *authorization.Attributes
		clusterRoles        []*corev2.ClusterRole
		clusterRoleBindings []*corev2.ClusterRoleBinding
		roles               []*corev2.Role
		roleBindings        []*corev2.RoleBinding
		wantNamespace       bool
		wantErr             bool
	}{
		{
			name:      "no access",
			namespace: "dev",
			attrs: &authorization.Attributes{
				User: corev2.User{
					Username: "foo",
				},
			},
			clusterRoles: []*corev2.ClusterRole{
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
			clusterRoleBindings: []*corev2.ClusterRoleBinding{
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
			wantNamespace: false,
			wantErr:       true,
		},
		{
			name:      "explicit access through ClusterRole & ClusterRoleBinding",
			namespace: "dev",
			attrs: &authorization.Attributes{
				User: corev2.User{
					Username: "foo",
					Groups:   []string{"cluster-admins"},
				},
			},
			clusterRoles: []*corev2.ClusterRole{
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
			clusterRoleBindings: []*corev2.ClusterRoleBinding{
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
			wantNamespace: true,
			wantErr:       false,
		},
		{
			name:      "explicit access to a single namespace through ClusterRole & ClusterRoleBinding",
			namespace: "dev",
			attrs: &authorization.Attributes{
				User: corev2.User{
					Username: "foo",
					Groups:   []string{"cluster-admins"},
				},
			},
			clusterRoles: []*corev2.ClusterRole{
				{
					ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
					Rules: []corev2.Rule{
						{
							Verbs:         []string{corev2.VerbAll},
							Resources:     []string{corev2.NamespacesResource},
							ResourceNames: []string{"dev"},
						},
					},
				},
			},
			clusterRoleBindings: []*corev2.ClusterRoleBinding{
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
			wantNamespace: true,
			wantErr:       false,
		},
		{
			name:      "explicit access to a single namespace through ClusterRole & ClusterRoleBinding should match the namespace",
			namespace: "dev",
			attrs: &authorization.Attributes{
				User: corev2.User{
					Username: "foo",
					Groups:   []string{"cluster-admins"},
				},
			},
			clusterRoles: []*corev2.ClusterRole{
				{
					ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
					Rules: []corev2.Rule{
						{
							Verbs:         []string{corev2.VerbAll},
							Resources:     []string{corev2.NamespacesResource},
							ResourceNames: []string{"default"},
						},
					},
				},
			},
			clusterRoleBindings: []*corev2.ClusterRoleBinding{
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
			wantNamespace: false,
			wantErr:       true,
		},
		{
			name:      "implicit access through ClusterRole & ClusterRoleBinding",
			namespace: "dev",
			attrs: &authorization.Attributes{
				User: corev2.User{
					Username: "foo",
					Groups:   []string{"cluster-admins"},
				},
			},
			clusterRoles: []*corev2.ClusterRole{
				{
					ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
					Rules: []corev2.Rule{
						{
							Verbs:     []string{corev2.VerbAll},
							Resources: []string{corev2.ChecksResource},
						},
					},
				},
			},
			clusterRoleBindings: []*corev2.ClusterRoleBinding{
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
			wantNamespace: true,
			wantErr:       false,
		},
		{
			name:      "implicit access through ClusterRole & ClusterRoleBinding should only work for namespaced resources",
			namespace: "dev",
			attrs: &authorization.Attributes{
				User: corev2.User{
					Username: "foo",
					Groups:   []string{"cluster-admins"},
				},
			},
			clusterRoles: []*corev2.ClusterRole{
				{
					ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
					Rules: []corev2.Rule{
						{
							Verbs:     []string{corev2.VerbAll},
							Resources: []string{corev2.UsersResource},
						},
					},
				},
			},
			clusterRoleBindings: []*corev2.ClusterRoleBinding{
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
			wantNamespace: false,
			wantErr:       true,
		},
		{
			name:      "implicit access through Role & RoleBinding",
			namespace: "dev",
			attrs: &authorization.Attributes{
				User: corev2.User{
					Username: "foo",
					Groups:   []string{"check-reader"},
				},
			},
			roles: []*corev2.Role{
				{
					ObjectMeta: corev2.NewObjectMeta("check-reader", "dev"),
					Rules: []corev2.Rule{
						{
							Verbs:     []string{"get"},
							Resources: []string{corev2.ChecksResource},
						},
					},
				},
			},
			roleBindings: []*corev2.RoleBinding{
				{
					Subjects: []corev2.Subject{
						{
							Type: "User",
							Name: "foo",
						},
					},
					RoleRef: corev2.RoleRef{
						Type: "Role",
						Name: "check-reader",
					},
					ObjectMeta: corev2.NewObjectMeta("check-reader", "dev"),
				},
			},
			wantNamespace: true,
			wantErr:       false,
		},
		{
			name:      "implicit access through ClusterRole & RoleBinding",
			namespace: "dev",
			attrs: &authorization.Attributes{
				User: corev2.User{
					Username: "foo",
					Groups:   []string{"ops"},
				},
			},
			clusterRoles: []*corev2.ClusterRole{
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
			roleBindings: []*corev2.RoleBinding{
				{
					Subjects: []corev2.Subject{
						{
							Type: "Group",
							Name: "ops",
						},
					},
					RoleRef: corev2.RoleRef{
						Type: "ClusterRole",
						Name: "cluster-admin",
					},
					ObjectMeta: corev2.NewObjectMeta("ops", "dev"),
				},
			},
			wantNamespace: true,
			wantErr:       false,
		},
		{
			name:      "explicit access to all namespaces can only be granted through ClusterRoleBindings",
			namespace: "default",
			attrs: &authorization.Attributes{
				User: corev2.User{
					Username: "foo",
					Groups:   []string{"local-admins"},
				},
			},
			roles: []*corev2.Role{
				{
					ObjectMeta: corev2.NewObjectMeta("local-admin", "dev"),
					Rules: []corev2.Rule{
						{
							Verbs:     []string{corev2.VerbAll},
							Resources: []string{corev2.ResourceAll},
						},
					},
				},
			},
			roleBindings: []*corev2.RoleBinding{
				{
					Subjects: []corev2.Subject{
						{
							Type: "Group",
							Name: "local-admins",
						},
					},
					RoleRef: corev2.RoleRef{
						Type: "Role",
						Name: "local-admin",
					},
					ObjectMeta: corev2.NewObjectMeta("local-admin", "dev"),
				},
			},
			wantNamespace: false,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := new(mockstore.MockStore)
			store.On("ListClusterRoles", mock.Anything, mock.Anything).Return(tt.clusterRoles, nil)
			store.On("ListClusterRoleBindings", mock.Anything, mock.Anything).Return(tt.clusterRoleBindings, nil)
			store.On("ListRoles", mock.Anything, mock.Anything).Return(tt.roles, nil)
			store.On("ListRoleBindings", mock.Anything, mock.Anything).Return(tt.roleBindings, nil)
			store.On("GetResource", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Run(func(args mock.Arguments) {
				resource := args[2].(*corev2.Namespace)
				*resource = *corev2.FixtureNamespace("dev")
			}).Return(nil)
			setupGetClusterRoleAndGetRole(store, tt.clusterRoles, tt.roles)

			ctx := contextWithUser(defaultContext(), tt.attrs.User.Username, tt.attrs.User.Groups)

			auth := &rbac.Authorizer{Store: store}
			client := NewNamespaceClient(store, auth)

			got, err := client.FetchNamespace(ctx, tt.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("NamespaceClient.FetchNamespace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantNamespace && (got == nil) {
				t.Fatalf("permission error: want access to namespace? %+v, got %+v", tt.wantNamespace, got)
			}
		})
	}
}

func TestNamespaceList(t *testing.T) {
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
		WantErr             bool
	}{
		{
			Name: "explicit access to all namespaces through ClusterRole & ClusterRoleBinding",
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
			ExpNamespaces: nil,
			WantErr:       true,
		},
		{
			Name: "explicit partial access through ClusterRole & ClusterRoleBinding",
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
					ObjectMeta: corev2.NewObjectMeta("pleb", ""),
					Rules: []corev2.Rule{
						{
							Verbs:         []string{"get"},
							Resources:     []string{corev2.NamespacesResource},
							ResourceNames: []string{"a", "c", "e"},
						},
					},
				},
			},
			ClusterRoleBindings: []*corev2.ClusterRoleBinding{
				{
					Subjects: []corev2.Subject{
						{
							Type: "Group",
							Name: "plebs",
						},
					},
					RoleRef: corev2.RoleRef{
						Type: "ClusterRole",
						Name: "pleb",
					},
					ObjectMeta: corev2.NewObjectMeta("pleb", ""),
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
			Name: "implicit access through ClusterRole & ClusterRoleBinding",
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
					ObjectMeta: corev2.NewObjectMeta("pleb", ""),
					Rules: []corev2.Rule{
						{
							Verbs:     []string{"get", "list"},
							Resources: []string{corev2.ChecksResource},
						},
					},
				},
			},
			ClusterRoleBindings: []*corev2.ClusterRoleBinding{
				{
					Subjects: []corev2.Subject{
						{
							Type: "Group",
							Name: "plebs",
						},
					},
					RoleRef: corev2.RoleRef{
						Type: "ClusterRole",
						Name: "pleb",
					},
					ObjectMeta: corev2.NewObjectMeta("pleb", ""),
				},
			},
			AllNamespaces: namespaces,
			ExpNamespaces: namespaces,
		},
		{
			Name: "implicit access through Role & RoleBinding",
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
			Roles: []*corev2.Role{
				{
					ObjectMeta: corev2.NewObjectMeta("pleb", "a"),
					Rules: []corev2.Rule{
						{
							Verbs:     []string{"delete"},
							Resources: []string{corev2.ChecksResource},
						},
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
		{
			Name: "implicit access through ClusterRole & RoleBinding",
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
					ObjectMeta: corev2.NewObjectMeta("pleb", ""),
					Rules: []corev2.Rule{
						{
							Verbs:     []string{"get", "list"},
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
						Type: "ClusterRole",
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
		{
			Name: "explicit access to all namespaces can only be granted through ClusterRoleBindings",
			Attrs: &authorization.Attributes{
				APIGroup:   "core",
				APIVersion: "v2",
				Resource:   corev2.NamespacesResource,
				User: corev2.User{
					Username: "operator",
					Groups:   []string{"local-admins"},
				},
			},
			Roles: []*corev2.Role{
				{
					ObjectMeta: corev2.NewObjectMeta("local-admin", "a"),
					Rules: []corev2.Rule{
						{
							Verbs:     []string{corev2.VerbAll},
							Resources: []string{corev2.ResourceAll},
						},
					},
				},
			},
			RoleBindings: []*corev2.RoleBinding{
				{
					Subjects: []corev2.Subject{
						{
							Type: "Group",
							Name: "local-admins",
						},
					},
					RoleRef: corev2.RoleRef{
						Type: "Role",
						Name: "local-admin",
					},
					ObjectMeta: corev2.NewObjectMeta("local-admin", "a"),
				},
			},
			AllNamespaces: namespaces,
			ExpNamespaces: []*corev2.Namespace{
				namespaces[0],
			},
			WantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			s := new(mockstore.MockStore)
			s.On("ListClusterRoles", mock.Anything, mock.Anything).Return(test.ClusterRoles, nil)
			s.On("ListClusterRoleBindings", mock.Anything, mock.Anything).Return(test.ClusterRoleBindings, nil)
			s.On("ListRoles", mock.Anything, mock.Anything).Return(test.Roles, nil)
			s.On("ListRoleBindings", mock.Anything, mock.Anything).Return(test.RoleBindings, nil)
			s.On("ListResources", mock.Anything, corev2.NamespacesResource, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				resources := args[2].(*[]*corev2.Namespace)
				*resources = append(*resources, test.AllNamespaces...)
			}).Return(nil)
			setupGetClusterRoleAndGetRole(s, test.ClusterRoles, test.Roles)

			ctx := contextWithUser(defaultContext(), test.Attrs.User.Username, test.Attrs.User.Groups)

			auth := &rbac.Authorizer{Store: s}
			client := NewNamespaceClient(s, auth)

			got, err := client.ListNamespaces(ctx, &store.SelectionPredicate{})
			if (err != nil) != test.WantErr {
				t.Errorf("NamespaceClient.ListNamespaces() error = %v, wantErr %v", err, test.WantErr)
				return
			}

			sort.Slice(got, sortFunc(got))

			if got, want := got, test.ExpNamespaces; !reflect.DeepEqual(got, want) {
				t.Fatalf("bad namespaces: got %+v, want %+v", got, want)
			}
		})
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

func sortFunc(namespaces []*corev2.Namespace) func(i, j int) bool {
	return func(i, j int) bool {
		return namespaces[i].Name < namespaces[j].Name
	}
}
