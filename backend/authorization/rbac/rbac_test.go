package rbac

import (
	"errors"
	"testing"

	"github.com/sensu/sensu-go/internal/apis/rbac"
	"github.com/sensu/sensu-go/testing/mockstorage"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
)

func TestAuthorize(t *testing.T) {
	type storeFunc func(*mockstorage.MockStorage)
	tests := []struct {
		name      string
		reqInfo   *types.RequestInfo
		storeFunc storeFunc
		want      bool
		wantErr   bool
	}{
		{
			name:    "no bindings",
			reqInfo: &types.RequestInfo{Namespace: "acme"},
			storeFunc: func(store *mockstorage.MockStorage) {
				store.On("List", mock.AnythingOfType("*context.emptyCtx"), "clusterrolebindings", mock.Anything).Return(nil)
				store.On("List", mock.AnythingOfType("*context.emptyCtx"), "rolebindings/acme", mock.Anything).Return(nil)
			},
			want: false,
		},
		{
			name: "clusterrolebindings store err",
			storeFunc: func(store *mockstorage.MockStorage) {
				store.On("List", mock.AnythingOfType("*context.emptyCtx"), "clusterrolebindings", mock.Anything).
					Return(errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "no matching clusterRoleBindings",
			reqInfo: &types.RequestInfo{
				Namespace: "acme",
				User: types.User{
					Username: "foo",
				},
			},
			storeFunc: func(store *mockstorage.MockStorage) {
				store.On("List", mock.AnythingOfType("*context.emptyCtx"), "clusterrolebindings", mock.Anything).Return(nil).
					Run(func(args mock.Arguments) {
						clusterRoleBindings := args.Get(2).(*[]rbac.ClusterRoleBinding)
						*clusterRoleBindings = append(*clusterRoleBindings, rbac.ClusterRoleBinding{
							Subjects: []rbac.Subject{
								rbac.Subject{Kind: rbac.UserKind, Name: "bar"},
							},
						})
					})
				store.On("List", mock.AnythingOfType("*context.emptyCtx"), "rolebindings/acme", mock.Anything).Return(nil)
			},
			want: false,
		},
		{
			name: "clusterroles/admin store err",
			reqInfo: &types.RequestInfo{
				User: types.User{
					Username: "foo",
				},
			},
			storeFunc: func(store *mockstorage.MockStorage) {
				store.On("List", mock.AnythingOfType("*context.emptyCtx"), "clusterrolebindings", mock.Anything).Return(nil).
					Run(func(args mock.Arguments) {
						clusterRoleBindings := args.Get(2).(*[]rbac.ClusterRoleBinding)
						*clusterRoleBindings = append(*clusterRoleBindings, rbac.ClusterRoleBinding{
							RoleRef: rbac.RoleRef{
								Name: "admin",
							},
							Subjects: []rbac.Subject{
								rbac.Subject{Kind: rbac.UserKind, Name: "foo"},
							},
						})
					})
				store.On("Get", mock.AnythingOfType("*context.emptyCtx"), "clusterroles/admin", mock.Anything).Return(errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "matching clusterRoleBinding",
			reqInfo: &types.RequestInfo{
				Verb:         "create",
				APIGroup:     "checks.sensu.io",
				Resource:     "checks",
				ResourceName: "check-cpu",
				User: types.User{
					Username: "foo",
				},
			},
			storeFunc: func(store *mockstorage.MockStorage) {
				store.On("List", mock.AnythingOfType("*context.emptyCtx"), "clusterrolebindings", mock.Anything).Return(nil).
					Run(func(args mock.Arguments) {
						clusterRoleBindings := args.Get(2).(*[]rbac.ClusterRoleBinding)
						*clusterRoleBindings = append(*clusterRoleBindings, rbac.ClusterRoleBinding{
							RoleRef: rbac.RoleRef{
								Name: "admin",
							},
							Subjects: []rbac.Subject{
								rbac.Subject{Kind: rbac.UserKind, Name: "foo"},
							},
						})
					})
				store.On("Get", mock.AnythingOfType("*context.emptyCtx"), "clusterroles/admin", mock.Anything).Return(nil).
					Run(func(args mock.Arguments) {
						clusterRole := args.Get(2).(*rbac.ClusterRole)
						*clusterRole = rbac.ClusterRole{Rules: []rbac.Rule{
							rbac.Rule{
								Verbs:         []string{"create"},
								APIGroups:     []string{"checks.sensu.io"},
								Resources:     []string{"checks"},
								ResourceNames: []string{"check-cpu"},
							},
						}}
					})
			},
			want: true,
		},
		{
			name:    "rolebindings store err",
			reqInfo: &types.RequestInfo{Namespace: "acme"},
			storeFunc: func(store *mockstorage.MockStorage) {
				store.On("List", mock.AnythingOfType("*context.emptyCtx"), "clusterrolebindings", mock.Anything).Return(nil)
				store.On("List", mock.AnythingOfType("*context.emptyCtx"), "rolebindings/acme", mock.Anything).Return(errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "no matching roleBindings",
			reqInfo: &types.RequestInfo{
				Namespace: "acme",
				User: types.User{
					Username: "foo",
				},
			},
			storeFunc: func(store *mockstorage.MockStorage) {
				store.On("List", mock.AnythingOfType("*context.emptyCtx"), "clusterrolebindings", mock.Anything).Return(nil)
				store.On("List", mock.AnythingOfType("*context.emptyCtx"), "rolebindings/acme", mock.Anything).Return(nil).
					Run(func(args mock.Arguments) {
						roleBindings := args.Get(2).(*[]rbac.RoleBinding)
						*roleBindings = append(*roleBindings, rbac.RoleBinding{
							Subjects: []rbac.Subject{
								rbac.Subject{Kind: rbac.UserKind, Name: "bar"},
							},
						})
					})
			},
			want: false,
		},
		{
			name: "roles/admin store err",
			reqInfo: &types.RequestInfo{
				Namespace: "acme",
				User: types.User{
					Username: "foo",
				},
			},
			storeFunc: func(store *mockstorage.MockStorage) {
				store.On("List", mock.AnythingOfType("*context.emptyCtx"), "clusterrolebindings", mock.Anything).Return(nil)
				store.On("List", mock.AnythingOfType("*context.emptyCtx"), "rolebindings/acme", mock.Anything).Return(nil).
					Run(func(args mock.Arguments) {
						roleBindings := args.Get(2).(*[]rbac.RoleBinding)
						*roleBindings = append(*roleBindings, rbac.RoleBinding{
							RoleRef: rbac.RoleRef{Name: "admin"},
							Subjects: []rbac.Subject{
								rbac.Subject{Kind: rbac.UserKind, Name: "foo"},
							},
						})
					})
				store.On("Get", mock.AnythingOfType("*context.emptyCtx"), "roles/acme/admin", mock.Anything).Return(errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "matching roleBinding",
			reqInfo: &types.RequestInfo{
				Namespace: "acme",
				User: types.User{
					Username: "foo",
				},
				Verb:         "create",
				APIGroup:     "checks.sensu.io",
				Resource:     "checks",
				ResourceName: "check-cpu",
			},
			storeFunc: func(store *mockstorage.MockStorage) {
				store.On("List", mock.AnythingOfType("*context.emptyCtx"), "clusterrolebindings", mock.Anything).Return(nil)
				store.On("List", mock.AnythingOfType("*context.emptyCtx"), "rolebindings/acme", mock.Anything).Return(nil).
					Run(func(args mock.Arguments) {
						roleBindings := args.Get(2).(*[]rbac.RoleBinding)
						*roleBindings = append(*roleBindings, rbac.RoleBinding{
							RoleRef: rbac.RoleRef{Name: "admin"},
							Subjects: []rbac.Subject{
								rbac.Subject{Kind: rbac.UserKind, Name: "foo"},
							},
						})
					})
				store.On("Get", mock.AnythingOfType("*context.emptyCtx"), "roles/acme/admin", mock.Anything).Return(nil).
					Run(func(args mock.Arguments) {
						role := args.Get(2).(*rbac.Role)
						*role = rbac.Role{Rules: []rbac.Rule{
							rbac.Rule{
								Verbs:         []string{"create"},
								APIGroups:     []string{"checks.sensu.io"},
								Resources:     []string{"checks"},
								ResourceNames: []string{"check-cpu"},
							},
						}}
					})
			},
			want: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := &mockstorage.MockStorage{}
			a := &Authorizer{
				Store: store,
			}
			tc.storeFunc(store)

			got, err := a.Authorize(tc.reqInfo)
			if (err != nil) != tc.wantErr {
				t.Errorf("Authorizer.Authorize() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if got != tc.want {
				t.Errorf("Authorizer.Authorize() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMatchesUser(t *testing.T) {
	tests := []struct {
		name     string
		user     types.User
		subjects []rbac.Subject
		want     bool
	}{
		{
			name: "not matching",
			user: types.User{Username: "foo"},
			subjects: []rbac.Subject{
				rbac.Subject{Kind: rbac.UserKind, Name: "bar"},
				rbac.Subject{Kind: rbac.GroupKind, Name: "foo"},
			},
			want: false,
		},
		{
			name: "matching via username",
			user: types.User{Username: "foo"},
			subjects: []rbac.Subject{
				rbac.Subject{Kind: rbac.UserKind, Name: "bar"},
				rbac.Subject{Kind: rbac.UserKind, Name: "foo"},
			},
			want: true,
		},
		// {
		// 	name: "matching via group",
		// 	user: types.User{Username: "foo", Groups: []string{"acme"}},
		// 	subjects: []rbac.Subject{
		// 		rbac.Subject{Kind: rbac.GroupKind, Name: "acme"},
		// 	},
		// 	want: true,
		// },
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := matchesUser(tc.user, tc.subjects); got != tc.want {
				t.Errorf("matchesUser() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRuleAllows(t *testing.T) {
	tests := []struct {
		name    string
		reqInfo *types.RequestInfo
		rule    rbac.Rule
		want    bool
	}{
		{
			name: "verb does not match",
			reqInfo: &types.RequestInfo{
				Verb: "create",
			},
			rule: rbac.Rule{
				Verbs: []string{"get"},
			},
			want: false,
		},
		{
			name: "api group does not match",
			reqInfo: &types.RequestInfo{
				Verb:     "create",
				APIGroup: "rbac.authorization.sensu.io",
			},
			rule: rbac.Rule{
				Verbs:     []string{"create"},
				APIGroups: []string{"core.sensu.io"},
			},
			want: false,
		},
		{
			name: "resource does not match",
			reqInfo: &types.RequestInfo{
				Verb:     "create",
				APIGroup: "core.sensu.io",
				Resource: "events",
			},
			rule: rbac.Rule{
				Verbs:     []string{"create"},
				APIGroups: []string{"core.sensu.io"},
				Resources: []string{"checks", "handlers"},
			},
			want: false,
		},
		{
			name: "resource name does not match",
			reqInfo: &types.RequestInfo{
				Verb:         "create",
				APIGroup:     "core.sensu.io",
				Resource:     "checks",
				ResourceName: "check-cpu",
			},
			rule: rbac.Rule{
				Verbs:         []string{"create"},
				APIGroups:     []string{"core.sensu.io"},
				Resources:     []string{"checks"},
				ResourceNames: []string{"check-mem"},
			},
			want: false,
		},
		{
			name: "matches",
			reqInfo: &types.RequestInfo{
				Verb:         "create",
				APIGroup:     "checks.sensu.io",
				Resource:     "checks",
				ResourceName: "check-cpu",
			},
			rule: rbac.Rule{
				Verbs:         []string{"create"},
				APIGroups:     []string{"checks.sensu.io"},
				Resources:     []string{"checks"},
				ResourceNames: []string{"check-cpu"},
			},
			want: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ruleAllows(tc.reqInfo, tc.rule); got != tc.want {
				t.Errorf("ruleAllows() = %v, want %v", got, tc.want)
			}
		})
	}
}
