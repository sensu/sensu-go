package api

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"
	"testing"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
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
							Type: corev2.GroupType,
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
							Type: corev2.GroupType,
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
							Type: corev2.GroupType,
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
							Type: corev2.GroupType,
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
							Type: corev2.GroupType,
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
							Type: corev2.GroupType,
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
							Type: corev2.GroupType,
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
							Type: corev2.GroupType,
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

			entityCfg := corev3.FixtureEntityConfig("foobar")
			// set templated namespace
			entityCfg.Metadata.Namespace = "{{ .Namespace }}"
			tmplEntityConfig, _ := json.Marshal(entityCfg)

			resourceTemplate := &corev3.ResourceTemplate{
				Metadata: &corev2.ObjectMeta{
					Namespace:   "default",
					Name:        "tmpl-entity-config",
					Labels:      make(map[string]string),
					Annotations: make(map[string]string),
				},
				APIVersion: "core/v3",
				Type:       "EntityConfig",
				Template:   string(tmplEntityConfig),
			}
			wrappedResourceTemplate, err := storev2.WrapResource(resourceTemplate)
			if err != nil {
				t.Fatal(err)
			}
			wrapList := wrap.List{wrappedResourceTemplate.(*wrap.Wrapper)}
			s2 := new(mockstore.V2MockStore)
			s2.On("CreateOrUpdate", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			s2.On("List", mock.Anything, mock.Anything, mock.Anything).Return(wrapList, nil)

			auth := &rbac.Authorizer{Store: store}
			client := NewNamespaceClient(store, store, auth, s2)

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
							Type: corev2.GroupType,
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
							Type: corev2.GroupType,
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
							Type: corev2.GroupType,
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
							Type: corev2.GroupType,
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
							Type: corev2.GroupType,
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
							Type: corev2.GroupType,
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
							Type: corev2.GroupType,
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

			entityCfg := corev3.FixtureEntityConfig("foobar")
			entityCfg.Metadata.Namespace = "{{ .Namespace }}" // templated namespace
			tmplEntityConfig, _ := json.Marshal(entityCfg)

			resourceTemplate := &corev3.ResourceTemplate{
				Metadata: &corev2.ObjectMeta{
					Namespace:   "default",
					Name:        "tmpl-entity-config",
					Labels:      make(map[string]string),
					Annotations: make(map[string]string),
				},
				APIVersion: "core/v3",
				Type:       "EntityConfig",
				Template:   string(tmplEntityConfig),
			}
			wrappedResourceTemplate, err := storev2.WrapResource(resourceTemplate)
			if err != nil {
				t.Fatal(err)
			}
			wrapList := wrap.List{wrappedResourceTemplate.(*wrap.Wrapper)}
			s2 := new(mockstore.V2MockStore)
			s2.On("CreateOrUpdate", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			s2.On("List", mock.Anything, mock.Anything, mock.Anything).Return(wrapList, nil)

			ctx := contextWithUser(defaultContext(), test.Attrs.User.Username, test.Attrs.User.Groups)

			auth := &rbac.Authorizer{Store: s}
			client := NewNamespaceClient(s, s, auth, s2)

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

func TestNamespaceCreateSideEffects(t *testing.T) {
	clusterRoles := []*corev2.ClusterRole{
		{
			ObjectMeta: corev2.NewObjectMeta("cluster-admin", "cluster-admin"),
			Rules: []corev2.Rule{
				{
					Verbs:     []string{corev2.VerbAll},
					Resources: []string{corev2.ResourceAll},
				},
			},
		},
	}
	clusterRoleBindings := []*corev2.ClusterRoleBinding{
		{
			Subjects: []corev2.Subject{
				{
					Type: corev2.GroupType,
					Name: "cluster-admins",
				},
			},
			RoleRef: corev2.RoleRef{
				Type: "ClusterRole",
				Name: "cluster-admin",
			},
			ObjectMeta: corev2.NewObjectMeta("cluster-admin", "cluster-admin"),
		},
	}
	s := new(mockstore.MockStore)
	s.On("ListClusterRoles", mock.Anything, mock.Anything).Return(clusterRoles, nil)
	s.On("ListClusterRoleBindings", mock.Anything, mock.Anything).Return(clusterRoleBindings, nil)
	s.On("ListRoles", mock.Anything, mock.Anything).Return(([]*corev2.Role)(nil), nil)
	s.On("ListRoleBindings", mock.Anything, mock.Anything).Return(([]*corev2.RoleBinding)(nil), nil)
	s.On("CreateResource", mock.Anything, mock.Anything).Return(nil)
	s.On("CreateOrUpdateResource", mock.Anything, mock.Anything).Return(nil)
	setupGetClusterRoleAndGetRole(s, clusterRoles, nil)

	entityCfg := corev3.FixtureEntityConfig("bar")
	entityCfg.Metadata.Namespace = "{{ .Namespace }}"
	tmplEntityConfig, _ := json.Marshal(entityCfg)

	resourceTemplate := &corev3.ResourceTemplate{
		Metadata: &corev2.ObjectMeta{
			Namespace:   "default",
			Name:        "tmpl-entity-config",
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
		APIVersion: "core/v3",
		Type:       "EntityConfig",
		Template:   string(tmplEntityConfig),
	}
	wrappedResourceTemplate, err := storev2.WrapResource(resourceTemplate)
	if err != nil {
		t.Fatal(err)
	}
	wrapList := wrap.List{wrappedResourceTemplate.(*wrap.Wrapper)}
	s2 := new(mockstore.V2MockStore)
	s2.On("CreateOrUpdate", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s2.On("List", mock.Anything, mock.Anything, mock.Anything).Return(wrapList, nil)

	ctx := contextWithUser(context.Background(), "cluster-admin", []string{"cluster-admins"})

	auth := &rbac.Authorizer{Store: s}
	client := NewNamespaceClient(s, s, auth, s2)

	namespace := &corev2.Namespace{Name: "test_namespace"}
	if err := client.CreateNamespace(ctx, namespace); err != nil {
		t.Fatal(err)
	}

	expRole := &corev2.Role{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "test_namespace",
			Name:      pipelineRoleName,
			CreatedBy: "cluster-admin",
		},
		Rules: []corev2.Rule{
			{
				Verbs:     []string{"get", "list"},
				Resources: []string{new(corev2.Event).RBACName()},
			},
		},
	}
	expBinding := &corev2.RoleBinding{
		Subjects: []corev2.Subject{
			{
				Type: corev2.GroupType,
				Name: pipelineRoleName,
			},
		},
		RoleRef: corev2.RoleRef{
			Name: pipelineRoleName,
			Type: "Role",
		},
		ObjectMeta: corev2.ObjectMeta{
			Name:      pipelineRoleName,
			Namespace: "test_namespace",
			CreatedBy: "cluster-admin",
		},
	}
	s.AssertNumberOfCalls(t, "CreateResource", 1)
	s.AssertNumberOfCalls(t, "CreateOrUpdateResource", 2)
	s.AssertCalled(t, "CreateOrUpdateResource", mock.Anything, expRole)
	s.AssertCalled(t, "CreateOrUpdateResource", mock.Anything, expBinding)

	s.On("DeleteResource", mock.Anything, expBinding.StorePrefix(), pipelineRoleName).Run(func(args mock.Arguments) {
		ctx := args[0].(context.Context)
		if ns := corev2.ContextNamespace(ctx); ns != namespace.Name {
			t.Fatalf("expected namespace %q, got %q", namespace.Name, ns)
		}
	}).Return(nil)
	s.On("DeleteResource", mock.Anything, expRole.StorePrefix(), pipelineRoleName).Run(func(args mock.Arguments) {
		ctx := args[0].(context.Context)
		if ns := corev2.ContextNamespace(ctx); ns != namespace.Name {
			t.Fatalf("expected namespace %q, got %q", namespace.Name, ns)
		}
	}).Return(nil)
	s.On("DeleteNamespace", mock.Anything, namespace.Name).Return(nil)

	if err := client.DeleteNamespace(ctx, namespace.Name); err != nil {
		t.Fatal(err)
	}

	s.AssertNumberOfCalls(t, "DeleteResource", 2)
	s.AssertCalled(t, "DeleteResource", mock.Anything, expRole.StorePrefix(), pipelineRoleName)
	s.AssertCalled(t, "DeleteResource", mock.Anything, expBinding.StorePrefix(), pipelineRoleName)

	s.AssertNumberOfCalls(t, "DeleteNamespace", 1)
	s.AssertCalled(t, "DeleteNamespace", mock.Anything, namespace.Name)

	s2.AssertCalled(t, "List", mock.Anything, mock.Anything, mock.Anything)
	s2.AssertCalled(t, "CreateOrUpdate", mock.Anything, mock.Anything, mock.Anything)
}

func TestNamespaceUpdateSideEffects(t *testing.T) {
	clusterRoles := []*corev2.ClusterRole{
		{
			ObjectMeta: corev2.NewObjectMeta("cluster-admin", "cluster-admin"),
			Rules: []corev2.Rule{
				{
					Verbs:     []string{corev2.VerbAll},
					Resources: []string{corev2.ResourceAll},
				},
			},
		},
	}
	clusterRoleBindings := []*corev2.ClusterRoleBinding{
		{
			Subjects: []corev2.Subject{
				{
					Type: corev2.GroupType,
					Name: "cluster-admins",
				},
			},
			RoleRef: corev2.RoleRef{
				Type: "ClusterRole",
				Name: "cluster-admin",
			},
			ObjectMeta: corev2.NewObjectMeta("cluster-admin", "cluster-admin"),
		},
	}
	s := new(mockstore.MockStore)
	s.On("ListClusterRoles", mock.Anything, mock.Anything).Return(clusterRoles, nil)
	s.On("ListClusterRoleBindings", mock.Anything, mock.Anything).Return(clusterRoleBindings, nil)
	s.On("ListRoles", mock.Anything, mock.Anything).Return(([]*corev2.Role)(nil), nil)
	s.On("ListRoleBindings", mock.Anything, mock.Anything).Return(([]*corev2.RoleBinding)(nil), nil)
	s.On("CreateResource", mock.Anything, mock.Anything).Return(nil)
	s.On("CreateOrUpdateResource", mock.Anything, mock.Anything).Return(nil)
	setupGetClusterRoleAndGetRole(s, clusterRoles, nil)

	entityCfg := corev3.FixtureEntityConfig("bar")
	entityCfg.Metadata.Namespace = "{{ .Namespace }}"
	tmplEntityConfig, _ := json.Marshal(entityCfg)

	resourceTemplate := &corev3.ResourceTemplate{
		Metadata: &corev2.ObjectMeta{
			Namespace:   "default",
			Name:        "tmpl-entity-config",
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
		APIVersion: "core/v3",
		Type:       "EntityConfig",
		Template:   string(tmplEntityConfig),
	}
	wrappedResourceTemplate, err := storev2.WrapResource(resourceTemplate)
	if err != nil {
		t.Fatal(err)
	}
	wrapList := wrap.List{wrappedResourceTemplate.(*wrap.Wrapper)}
	s2 := new(mockstore.V2MockStore)
	s2.On("CreateOrUpdate", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s2.On("List", mock.Anything, mock.Anything, mock.Anything).Return(wrapList, nil)

	ctx := contextWithUser(context.Background(), "cluster-admin", []string{"cluster-admins"})

	auth := &rbac.Authorizer{Store: s}
	client := NewNamespaceClient(s, s, auth, s2)

	namespace := &corev2.Namespace{Name: "test_namespace"}
	if err := client.UpdateNamespace(ctx, namespace); err != nil {
		t.Fatal(err)
	}

	expRole := &corev2.Role{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "test_namespace",
			Name:      pipelineRoleName,
			CreatedBy: "cluster-admin",
		},
		Rules: []corev2.Rule{
			{
				Verbs:     []string{"get", "list"},
				Resources: []string{new(corev2.Event).RBACName()},
			},
		},
	}
	expBinding := &corev2.RoleBinding{
		Subjects: []corev2.Subject{
			{
				Type: corev2.GroupType,
				Name: pipelineRoleName,
			},
		},
		RoleRef: corev2.RoleRef{
			Name: pipelineRoleName,
			Type: "Role",
		},
		ObjectMeta: corev2.ObjectMeta{
			Name:      pipelineRoleName,
			Namespace: "test_namespace",
			CreatedBy: "cluster-admin",
		},
	}
	s.AssertNumberOfCalls(t, "CreateOrUpdateResource", 3)
	s.AssertCalled(t, "CreateOrUpdateResource", mock.Anything, expRole)
	s.AssertCalled(t, "CreateOrUpdateResource", mock.Anything, expBinding)

	s.On("DeleteResource", mock.Anything, expBinding.StorePrefix(), pipelineRoleName).Run(func(args mock.Arguments) {
		ctx := args[0].(context.Context)
		if ns := corev2.ContextNamespace(ctx); ns != namespace.Name {
			t.Fatalf("expected namespace %q, got %q", namespace.Name, ns)
		}
	}).Return(nil)
	s.On("DeleteResource", mock.Anything, expRole.StorePrefix(), pipelineRoleName).Run(func(args mock.Arguments) {
		ctx := args[0].(context.Context)
		if ns := corev2.ContextNamespace(ctx); ns != namespace.Name {
			t.Fatalf("expected namespace %q, got %q", namespace.Name, ns)
		}
	}).Return(nil)
	s.On("DeleteNamespace", mock.Anything, namespace.Name).Return(nil)

	if err := client.DeleteNamespace(ctx, namespace.Name); err != nil {
		t.Fatal(err)
	}

	s.AssertNumberOfCalls(t, "DeleteResource", 2)
	s.AssertCalled(t, "DeleteResource", mock.Anything, expRole.StorePrefix(), pipelineRoleName)
	s.AssertCalled(t, "DeleteResource", mock.Anything, expBinding.StorePrefix(), pipelineRoleName)

	s.AssertNumberOfCalls(t, "DeleteNamespace", 1)
	s.AssertCalled(t, "DeleteNamespace", mock.Anything, namespace.Name)

	s2.AssertCalled(t, "List", mock.Anything, mock.Anything, mock.Anything)
	s2.AssertCalled(t, "CreateOrUpdate", mock.Anything, mock.Anything, mock.Anything)
}
