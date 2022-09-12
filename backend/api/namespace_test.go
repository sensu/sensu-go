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
			var (
				ctx                            = store.NamespaceContext(context.Background(), tt.namespace)
				listClusterRolesRequest        = storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v2", Type: "ClusterRole"}, "", "", new(corev2.ClusterRole).StoreName())
				listClusterRoleBindingsRequest = storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v2", Type: "ClusterRoleBinding"}, "", "", new(corev2.ClusterRoleBinding).StoreName())
				listRolesRequest               = storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v2", Type: "Role"}, tt.namespace, "", new(corev2.Role).StoreName())
				listRoleBindingsRequest        = storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v2", Type: "RoleBinding"}, tt.namespace, "", new(corev2.RoleBinding).StoreName())
				getNamespaceRequest            = storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v3", Type: "Namespace"}, "", tt.namespace, new(corev3.Namespace).StoreName())
				listResourceTemplatesRequest   = storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v3", Type: "ResourceTemplate"}, "", "", new(corev3.ResourceTemplate).StoreName())
			)

			clusterRolesList := mockstore.WrapList[*corev2.ClusterRole](tt.clusterRoles)
			clusterRoleBindingsList := mockstore.WrapList[*corev2.ClusterRoleBinding](tt.clusterRoleBindings)
			rolesList := mockstore.WrapList[*corev2.Role](tt.roles)
			roleBindingsList := mockstore.WrapList[*corev2.RoleBinding](tt.roleBindings)
			devNamespace := mockstore.Wrapper[*corev3.Namespace]{Value: corev3.FixtureNamespace(tt.namespace)}

			entityCfg := corev3.FixtureEntityConfig("foobar")
			// set templated namespace
			entityCfg.Metadata.Namespace = "{{ .Namespace }}"
			tmplEntityConfig, _ := json.Marshal(entityCfg)

			resourceTemplateList := mockstore.WrapList[*corev3.ResourceTemplate]{
				&corev3.ResourceTemplate{
					Metadata: &corev2.ObjectMeta{
						Namespace:   "default",
						Name:        "tmpl-entity-config",
						Labels:      make(map[string]string),
						Annotations: make(map[string]string),
					},
					APIVersion: "core/v3",
					Type:       "EntityConfig",
					Template:   string(tmplEntityConfig),
				},
			}

			store := new(mockstore.V2MockStore)
			store.On("List", mock.Anything, listClusterRolesRequest, mock.Anything).Return(clusterRolesList, nil)
			store.On("List", mock.Anything, listClusterRoleBindingsRequest, mock.Anything).Return(clusterRoleBindingsList, nil)
			store.On("List", mock.Anything, listRolesRequest, mock.Anything).Return(rolesList, nil)
			store.On("List", mock.Anything, listRoleBindingsRequest, mock.Anything).Return(roleBindingsList, nil)
			store.On("Get", mock.Anything, getNamespaceRequest).Return(devNamespace, nil)
			store.On("List", mock.Anything, listResourceTemplatesRequest, mock.Anything).Return(resourceTemplateList, nil)

			setupGetClusterRoleAndGetRole(ctx, store, tt.clusterRoles, tt.roles)

			ctx = contextWithUser(ctx, tt.attrs.User.Username, tt.attrs.User.Groups)

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
	namespaces := []*corev3.Namespace{
		corev3.FixtureNamespace("a"),
		corev3.FixtureNamespace("b"),
		corev3.FixtureNamespace("c"),
		corev3.FixtureNamespace("d"),
		corev3.FixtureNamespace("e"),
		corev3.FixtureNamespace("f"),
	}
	tests := []struct {
		Name                string
		Attrs               *authorization.Attributes
		ClusterRoles        []*corev2.ClusterRole
		ClusterRoleBindings []*corev2.ClusterRoleBinding
		Roles               []*corev2.Role
		RoleBindings        []*corev2.RoleBinding
		AllNamespaces       []*corev3.Namespace
		ExpNamespaces       []*corev3.Namespace
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
			ExpNamespaces: []*corev3.Namespace{
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
			ExpNamespaces: []*corev3.Namespace{
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
			ExpNamespaces: []*corev3.Namespace{
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
			ExpNamespaces: []*corev3.Namespace{
				namespaces[0],
			},
			WantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var (
				ctx                            = contextWithUser(defaultContext(), test.Attrs.User.Username, test.Attrs.User.Groups)
				listClusterRolesRequest        = storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v2", Type: "ClusterRole"}, "", "", new(corev2.ClusterRole).StoreName())
				listClusterRoleBindingsRequest = storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v2", Type: "ClusterRoleBinding"}, "", "", new(corev2.ClusterRoleBinding).StoreName())
				listRolesRequest               = storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v2", Type: "Role"}, "default", "", new(corev2.Role).StoreName())
				listRoleBindingsRequest        = storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v2", Type: "RoleBinding"}, "default", "", new(corev2.RoleBinding).StoreName())
				listNamespacesRequest          = storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v3", Type: "Namespace"}, "", "", new(corev3.Namespace).StoreName())
				listResourceTemplatesRequest   = storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v3", Type: "ResourceTemplate"}, "default", "", new(corev3.ResourceTemplate).StoreName())
			)

			clusterRolesList := mockstore.WrapList[*corev2.ClusterRole](test.ClusterRoles)
			clusterRoleBindingsList := mockstore.WrapList[*corev2.ClusterRoleBinding](test.ClusterRoleBindings)
			rolesList := mockstore.WrapList[*corev2.Role](test.Roles)
			roleBindingsList := mockstore.WrapList[*corev2.RoleBinding](test.RoleBindings)
			namespaceList := mockstore.WrapList[*corev3.Namespace](test.AllNamespaces)

			entityCfg := corev3.FixtureEntityConfig("foobar")
			entityCfg.Metadata.Namespace = "{{ .Namespace }}" // templated namespace
			tmplEntityConfig, _ := json.Marshal(entityCfg)
			resourceTemplateList := mockstore.WrapList[*corev3.ResourceTemplate]{
				&corev3.ResourceTemplate{
					Metadata: &corev2.ObjectMeta{
						Namespace:   "default",
						Name:        "tmpl-entity-config",
						Labels:      make(map[string]string),
						Annotations: make(map[string]string),
					},
					APIVersion: "core/v3",
					Type:       "EntityConfig",
					Template:   string(tmplEntityConfig),
				},
			}

			s := new(mockstore.V2MockStore)
			s.On("List", mock.Anything, listRoleBindingsRequest, mock.Anything).Return(roleBindingsList, nil)
			s.On("List", mock.Anything, listClusterRoleBindingsRequest, mock.Anything).Return(clusterRoleBindingsList, nil)
			s.On("List", mock.Anything, listClusterRolesRequest, mock.Anything).Return(clusterRolesList, nil)
			s.On("List", mock.Anything, listRolesRequest, mock.Anything).Return(rolesList, nil)
			s.On("List", mock.Anything, listNamespacesRequest, mock.Anything).Return(namespaceList, nil)
			s.On("List", mock.Anything, listResourceTemplatesRequest, mock.Anything).Return(resourceTemplateList, nil)
			setupGetClusterRoleAndGetRole(ctx, s, test.ClusterRoles, test.Roles)

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

func setupGetClusterRoleAndGetRole(ctx context.Context, store *mockstore.V2MockStore, clusterRoles []*corev2.ClusterRole, roles []*corev2.Role) {
	for _, role := range clusterRoles {
		req := storev2.NewResourceRequestFromResource(role)
		store.On("Get", mock.Anything, req).Return(mockstore.Wrapper[*corev2.ClusterRole]{Value: role}, nil)
	}

	for _, role := range roles {
		req := storev2.NewResourceRequestFromResource(role)
		store.On("Get", mock.Anything, req).Return(mockstore.Wrapper[*corev2.Role]{Value: role}, nil)
	}
}

func sortFunc(namespaces []*corev3.Namespace) func(i, j int) bool {
	return func(i, j int) bool {
		return namespaces[i].Metadata.Name < namespaces[j].Metadata.Name
	}
}

func TestNamespaceCreateSideEffects(t *testing.T) {
	clusterRoles := []*corev2.ClusterRole{
		{
			ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
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
			ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
		},
	}

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
	ctx := defaultContext()
	listClusterRolesRequest := storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v2", Type: "ClusterRole"}, "", "", new(corev2.ClusterRole).StoreName())
	listClusterRoleBindingsRequest := storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v2", Type: "ClusterRoleBinding"}, "", "", new(corev2.ClusterRoleBinding).StoreName())
	listRolesRequest := storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v2", Type: "Role"}, "default", "", new(corev2.Role).StoreName())
	listRoleBindingsRequest := storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v2", Type: "RoleBinding"}, "default", "", new(corev2.RoleBinding).StoreName())
	listResourceTemplatesRequest := storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v3", Type: "ResourceTemplate"}, "", "", new(corev3.ResourceTemplate).StoreName())
	s := new(mockstore.V2MockStore)
	s.On("List", mock.Anything, listClusterRolesRequest, mock.Anything).Return(mockstore.WrapList[*corev2.ClusterRole](clusterRoles), nil)
	s.On("List", mock.Anything, listClusterRoleBindingsRequest, mock.Anything).Return(mockstore.WrapList[*corev2.ClusterRoleBinding](clusterRoleBindings), nil)
	s.On("List", mock.Anything, listRolesRequest, mock.Anything).Return(mockstore.WrapList[*corev2.Role]{}, nil)
	s.On("List", mock.Anything, listRoleBindingsRequest, mock.Anything).Return(mockstore.WrapList[*corev2.RoleBinding]{}, nil)
	s.On("List", mock.Anything, listResourceTemplatesRequest, mock.Anything).Return(mockstore.WrapList[*corev3.ResourceTemplate]{resourceTemplate}, nil)
	s.On("CreateIfNotExists", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.On("CreateOrUpdate", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	setupGetClusterRoleAndGetRole(ctx, s, clusterRoles, nil)

	ctx = contextWithUser(ctx, "cluster-admin", []string{"cluster-admins"})

	auth := &rbac.Authorizer{Store: s}
	client := NewNamespaceClient(s, auth)

	namespace := corev3.FixtureNamespace("test_namespace")
	if err := client.CreateNamespace(ctx, namespace); err != nil {
		t.Fatal(err)
	}

	s.AssertNumberOfCalls(t, "CreateIfNotExists", 1)
	s.AssertNumberOfCalls(t, "CreateOrUpdate", 3)

	s.On("Delete", mock.Anything, mock.Anything).Return(nil)

	if err := client.DeleteNamespace(ctx, namespace.Metadata.Name); err != nil {
		t.Fatal(err)
	}

	s.AssertNumberOfCalls(t, "Delete", 3)
}

func TestNamespaceUpdateSideEffects(t *testing.T) {
	clusterRoles := []*corev2.ClusterRole{
		{
			ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
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
			ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
		},
	}
	entityCfg := corev3.FixtureEntityConfig("bar")
	entityCfg.Metadata.Namespace = "{{ .Namespace }}"
	tmplEntityConfig, _ := json.Marshal(entityCfg)

	resourceTemplate := &corev3.ResourceTemplate{
		Metadata: &corev2.ObjectMeta{
			Name:        "tmpl-entity-config",
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
		APIVersion: "core/v3",
		Type:       "EntityConfig",
		Template:   string(tmplEntityConfig),
	}
	ctx := defaultContext()
	listClusterRolesRequest := storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v2", Type: "ClusterRole"}, "", "", new(corev2.ClusterRole).StoreName())
	listClusterRoleBindingsRequest := storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v2", Type: "ClusterRoleBinding"}, "", "", new(corev2.ClusterRoleBinding).StoreName())
	listRolesRequest := storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v2", Type: "Role"}, "default", "", new(corev2.Role).StoreName())
	listRoleBindingsRequest := storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v2", Type: "RoleBinding"}, "default", "", new(corev2.RoleBinding).StoreName())
	listResourceTemplatesRequest := storev2.NewResourceRequest(corev2.TypeMeta{APIVersion: "core/v3", Type: "ResourceTemplate"}, "", "", new(corev3.ResourceTemplate).StoreName())
	s := new(mockstore.V2MockStore)
	s.On("List", mock.Anything, listClusterRolesRequest, mock.Anything).Return(mockstore.WrapList[*corev2.ClusterRole](clusterRoles), nil)
	s.On("List", mock.Anything, listClusterRoleBindingsRequest, mock.Anything).Return(mockstore.WrapList[*corev2.ClusterRoleBinding](clusterRoleBindings), nil)
	s.On("List", mock.Anything, listRolesRequest, mock.Anything).Return(mockstore.WrapList[*corev2.Role]{}, nil)
	s.On("List", mock.Anything, listRoleBindingsRequest, mock.Anything).Return(mockstore.WrapList[*corev2.RoleBinding]{}, nil)
	s.On("List", mock.Anything, listResourceTemplatesRequest, mock.Anything).Return(mockstore.WrapList[*corev3.ResourceTemplate]{resourceTemplate}, nil)
	s.On("CreateIfNotExists", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.On("CreateOrUpdate", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	setupGetClusterRoleAndGetRole(ctx, s, clusterRoles, nil)

	ctx = contextWithUser(ctx, "cluster-admin", []string{"cluster-admins"})

	auth := &rbac.Authorizer{Store: s}
	client := NewNamespaceClient(s, auth)

	namespace := corev3.FixtureNamespace("test_namespace")
	if err := client.UpdateNamespace(ctx, namespace); err != nil {
		t.Fatal(err)
	}

	s.AssertNumberOfCalls(t, "CreateOrUpdate", 4)

	s.On("Delete", mock.Anything, mock.Anything).Return(nil)

	if err := client.DeleteNamespace(ctx, namespace.Metadata.Name); err != nil {
		t.Fatal(err)
	}

	s.AssertNumberOfCalls(t, "Delete", 3)
}
