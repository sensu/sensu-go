package rbac

import (
	"context"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
)

func TestAuthorize(t *testing.T) {
	type storeFunc func(*mockstore.MockStore)
	var nilClusterRoleBindings []*types.ClusterRoleBinding
	var nilRoleBindings []*types.RoleBinding
	tests := []struct {
		name      string
		attrs     *authorization.Attributes
		storeFunc storeFunc
		want      bool
		wantErr   bool
	}{
		{
			name:  "no bindings",
			attrs: &authorization.Attributes{Namespace: "acme"},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("ListClusterRoleBindings", mock.AnythingOfType("*context.emptyCtx"), &store.SelectionPredicate{}).
					Return(nilClusterRoleBindings, nil)
				s.On("ListRoleBindings", mock.AnythingOfType("*context.emptyCtx"), &store.SelectionPredicate{}).
					Return(nilRoleBindings, nil)
			},
			want: false,
		},
		{
			name: "ClusterRoleBindings store err",
			storeFunc: func(s *mockstore.MockStore) {
				s.On("ListClusterRoleBindings", mock.AnythingOfType("*context.emptyCtx"), &store.SelectionPredicate{}).
					Return(nilClusterRoleBindings, errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "no matching ClusterRoleBinding",
			attrs: &authorization.Attributes{
				Namespace: "acme",
				User: types.User{
					Username: "foo",
				},
			},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("ListClusterRoleBindings", mock.AnythingOfType("*context.emptyCtx"), &store.SelectionPredicate{}).
					Return([]*types.ClusterRoleBinding{&types.ClusterRoleBinding{
						Subjects: []types.Subject{
							types.Subject{Type: types.UserType, Name: "bar"},
						},
					}}, nil)
				s.On("ListRoleBindings", mock.AnythingOfType("*context.emptyCtx"), &store.SelectionPredicate{}).
					Return(nilRoleBindings, nil)
			},
			want: false,
		},
		{
			name: "GetClusterRole store err",
			attrs: &authorization.Attributes{
				User: types.User{
					Username: "foo",
				},
			},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("ListClusterRoleBindings", mock.AnythingOfType("*context.emptyCtx"), &store.SelectionPredicate{}).
					Return([]*types.ClusterRoleBinding{&types.ClusterRoleBinding{
						RoleRef: types.RoleRef{
							Type: "ClusterRole",
							Name: "admin",
						},
						Subjects: []types.Subject{
							types.Subject{Type: types.UserType, Name: "foo"},
						},
					}}, nil)
				s.On("GetClusterRole", mock.AnythingOfType("*context.emptyCtx"), "admin", mock.Anything).
					Return(nil, errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "matching ClusterRoleBinding",
			attrs: &authorization.Attributes{
				Verb:         "create",
				Resource:     "checks",
				ResourceName: "check-cpu",
				User: types.User{
					Username: "foo",
				},
			},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("ListClusterRoleBindings", mock.AnythingOfType("*context.emptyCtx"), &store.SelectionPredicate{}).
					Return([]*types.ClusterRoleBinding{&types.ClusterRoleBinding{
						RoleRef: types.RoleRef{
							Type: "ClusterRole",
							Name: "admin",
						},
						Subjects: []types.Subject{
							types.Subject{Type: types.UserType, Name: "foo"},
						},
					}}, nil)
				s.On("GetClusterRole", mock.AnythingOfType("*context.emptyCtx"), "admin", mock.Anything).
					Return(&types.ClusterRole{Rules: []types.Rule{
						types.Rule{
							Verbs:         []string{"create"},
							Resources:     []string{"checks"},
							ResourceNames: []string{"check-cpu"},
						},
					}}, nil)
			},
			want: true,
		},
		{
			name:  "RoleBindings store err",
			attrs: &authorization.Attributes{Namespace: "acme"},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("ListClusterRoleBindings", mock.AnythingOfType("*context.emptyCtx"), &store.SelectionPredicate{}).
					Return(nilClusterRoleBindings, nil)
				s.On("ListRoleBindings", mock.AnythingOfType("*context.emptyCtx"), &store.SelectionPredicate{}).
					Return(nilRoleBindings, errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "no matching RoleBindings",
			attrs: &authorization.Attributes{
				Namespace: "acme",
				User: types.User{
					Username: "foo",
				},
			},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("ListClusterRoleBindings", mock.AnythingOfType("*context.emptyCtx"), &store.SelectionPredicate{}).
					Return(nilClusterRoleBindings, nil)
				s.On("ListRoleBindings", mock.AnythingOfType("*context.emptyCtx"), &store.SelectionPredicate{}).
					Return([]*types.RoleBinding{&types.RoleBinding{
						RoleRef: types.RoleRef{
							Type: "Role",
							Name: "admin",
						},
						Subjects: []types.Subject{
							types.Subject{Type: types.UserType, Name: "foo"},
						},
					}}, nil)
				s.On("GetRole", mock.AnythingOfType("*context.emptyCtx"), "admin", mock.Anything).
					Return(nil, nil)
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "GetRole store err",
			attrs: &authorization.Attributes{
				Namespace: "acme",
				User: types.User{
					Username: "foo",
				},
			},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("ListClusterRoleBindings", mock.AnythingOfType("*context.emptyCtx"), &store.SelectionPredicate{}).
					Return(nilClusterRoleBindings, nil)
				s.On("ListRoleBindings", mock.AnythingOfType("*context.emptyCtx"), &store.SelectionPredicate{}).
					Return([]*types.RoleBinding{&types.RoleBinding{
						RoleRef: types.RoleRef{
							Type: "Role",
							Name: "admin",
						},
						Subjects: []types.Subject{
							types.Subject{Type: types.UserType, Name: "foo"},
						},
					}}, nil)
				s.On("GetRole", mock.AnythingOfType("*context.emptyCtx"), "admin", mock.Anything).
					Return(nil, errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "matching RoleBinding",
			attrs: &authorization.Attributes{
				Namespace: "acme",
				User: types.User{
					Username: "foo",
				},
				Verb:         "create",
				Resource:     "checks",
				ResourceName: "check-cpu",
			},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("ListClusterRoleBindings", mock.AnythingOfType("*context.emptyCtx"), &store.SelectionPredicate{}).
					Return(nilClusterRoleBindings, nil)

				s.On("ListRoleBindings", mock.AnythingOfType("*context.emptyCtx"), &store.SelectionPredicate{}).
					Return([]*types.RoleBinding{&types.RoleBinding{
						RoleRef: types.RoleRef{
							Type: "Role",
							Name: "admin",
						},
						Subjects: []types.Subject{
							types.Subject{Type: types.UserType, Name: "foo"},
						},
					}}, nil)
				s.On("GetRole", mock.AnythingOfType("*context.emptyCtx"), "admin", mock.Anything).
					Return(&types.Role{Rules: []types.Rule{
						types.Rule{
							Verbs:         []string{"create"},
							Resources:     []string{"checks"},
							ResourceNames: []string{"check-cpu"},
						},
					}}, nil)
			},
			want: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := &mockstore.MockStore{}
			a := &Authorizer{
				Store: store,
			}
			tc.storeFunc(store)

			got, err := a.Authorize(context.Background(), tc.attrs)
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
		subjects []types.Subject
		want     bool
	}{
		{
			name: "not matching",
			user: types.User{Username: "foo"},
			subjects: []types.Subject{
				types.Subject{Type: types.UserType, Name: "bar"},
				types.Subject{Type: types.GroupType, Name: "foo"},
			},
			want: false,
		},
		{
			name: "matching via username",
			user: types.User{Username: "foo"},
			subjects: []types.Subject{
				types.Subject{Type: types.UserType, Name: "bar"},
				types.Subject{Type: types.UserType, Name: "foo"},
			},
			want: true,
		},
		{
			name: "matching via group",
			user: types.User{Username: "foo", Groups: []string{"acme"}},
			subjects: []types.Subject{
				types.Subject{Type: types.GroupType, Name: "acme"},
			},
			want: true,
		},
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
		name  string
		attrs *authorization.Attributes
		rule  types.Rule
		want  bool
	}{
		{
			name: "verb does not match",
			attrs: &authorization.Attributes{
				Verb: "create",
			},
			rule: types.Rule{
				Verbs: []string{"get"},
			},
			want: false,
		},
		{
			name: "resource does not match",
			attrs: &authorization.Attributes{
				Verb:     "create",
				Resource: "events",
			},
			rule: types.Rule{
				Verbs:     []string{"create"},
				Resources: []string{"checks", "handlers"},
			},
			want: false,
		},
		{
			name: "resource name does not match",
			attrs: &authorization.Attributes{
				Verb:         "create",
				Resource:     "checks",
				ResourceName: "check-cpu",
			},
			rule: types.Rule{
				Verbs:         []string{"create"},
				Resources:     []string{"checks"},
				ResourceNames: []string{"check-mem"},
			},
			want: false,
		},
		{
			name: "matches",
			attrs: &authorization.Attributes{
				Verb:         "create",
				Resource:     "checks",
				ResourceName: "check-cpu",
			},
			rule: types.Rule{
				Verbs:         []string{"create"},
				Resources:     []string{"checks"},
				ResourceNames: []string{"check-cpu"},
			},
			want: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got, _ := ruleAllows(tc.attrs, tc.rule); got != tc.want {
				t.Errorf("ruleAllows() = %v, want %v", got, tc.want)
			}
		})
	}
}
