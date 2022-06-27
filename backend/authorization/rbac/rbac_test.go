package rbac

import (
	"context"
	"errors"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

type wrapList[T corev3.Resource] []T

func (w wrapList[T]) Unwrap() ([]corev3.Resource, error) {
	result := make([]corev3.Resource, 0)
	for _, resource := range w {
		result = append(result, resource)
	}
	return result, nil
}

func (w wrapList[T]) UnwrapInto(target interface{}) error {
	list, ok := target.(*[]T)
	if !ok {
		return errors.New("bad target")
	}
	*list = w
	return nil
}

func (w wrapList[T]) Len() int {
	return len(w)
}

type wrapper[T corev3.Resource] struct {
	Value T
}

func (w wrapper[T]) Unwrap() (corev3.Resource, error) {
	return w.Value, nil
}

func (w wrapper[T]) UnwrapInto(target interface{}) error {
	val, ok := target.(T)
	if !ok {
		panic("bad target")
	}
	reflect.ValueOf(val).Elem().Set(reflect.ValueOf(w.Value).Elem())
	return nil
}

func TestAuthorize(t *testing.T) {
	type storeFunc func(*mockstore.V2MockStore)
	ctx := store.NamespaceContext(context.Background(), "acme")
	clusterRoleBindingListRequest := storev2.NewResourceRequest(ctx, "acme", "", new(corev2.ClusterRoleBinding).StoreName())
	roleBindingListRequest := storev2.NewResourceRequest(ctx, "acme", "", new(corev2.RoleBinding).StoreName())
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
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("List", clusterRoleBindingListRequest, mock.Anything).
					Return(wrapList[*corev2.ClusterRoleBinding](nil), nil)
				s.On("List", roleBindingListRequest, mock.Anything).
					Return(wrapList[*corev2.RoleBinding](nil), nil)
			},
			want: false,
		},
		{
			name: "ClusterRoleBindings store err",
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("List", mock.Anything, mock.Anything).
					Return(wrapList[*corev2.ClusterRoleBinding](nil), errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "no matching ClusterRoleBinding",
			attrs: &authorization.Attributes{
				Namespace: "acme",
				User: corev2.User{
					Username: "foo",
				},
			},
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("List", clusterRoleBindingListRequest, mock.Anything).
					Return(wrapList[*corev2.ClusterRoleBinding]{{
						Subjects: []corev2.Subject{
							{Type: corev2.UserType, Name: "bar"},
						},
					}}, nil)
				s.On("List", roleBindingListRequest, mock.Anything).
					Return(wrapList[*corev2.RoleBinding](nil), nil)
			},
			want: false,
		},
		{
			name: "GetClusterRole store err",
			attrs: &authorization.Attributes{
				Namespace: "acme",
				User: corev2.User{
					Username: "foo",
				},
			},
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("List", mock.Anything, mock.Anything).
					Return(wrapList[*corev2.ClusterRoleBinding]{{
						RoleRef: corev2.RoleRef{
							Type: "ClusterRole",
							Name: "admin",
						},
						Subjects: []corev2.Subject{
							{Type: corev2.UserType, Name: "foo"},
						},
					}}, (error)(nil))
				s.On("Get", mock.Anything).
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
				User: corev2.User{
					Username: "foo",
				},
			},
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("List", clusterRoleBindingListRequest, mock.Anything).
					Return(wrapList[*corev2.ClusterRoleBinding]{{
						RoleRef: corev2.RoleRef{
							Type: "ClusterRole",
							Name: "admin",
						},
						Subjects: []corev2.Subject{
							{Type: corev2.UserType, Name: "foo"},
						},
					}}, nil)
				s.On("Get", mock.Anything).
					Return(wrapper[*corev2.ClusterRole](wrapper[*corev2.ClusterRole]{Value: &corev2.ClusterRole{Rules: []corev2.Rule{
						{
							Verbs:         []string{"create"},
							Resources:     []string{"checks"},
							ResourceNames: []string{"check-cpu"},
						},
					}}}), nil)
			},
			want: true,
		},
		{
			name:  "RoleBindings store err",
			attrs: &authorization.Attributes{Namespace: "acme"},
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("List", clusterRoleBindingListRequest, mock.Anything).
					Return(wrapList[*corev2.ClusterRoleBinding](nil), nil)
				s.On("List", roleBindingListRequest, mock.Anything).
					Return(wrapList[*corev2.RoleBinding](nil), errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "no matching RoleBindings",
			attrs: &authorization.Attributes{
				Namespace: "acme",
				User: corev2.User{
					Username: "foo",
				},
			},
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("List", clusterRoleBindingListRequest, mock.Anything).
					Return(wrapList[*corev2.ClusterRoleBinding](nil), nil)
				s.On("List", roleBindingListRequest, mock.Anything).
					Return(wrapList[*corev2.RoleBinding]{{
						RoleRef: corev2.RoleRef{
							Type: "Role",
							Name: "admin",
						},
						Subjects: []corev2.Subject{
							{Type: corev2.UserType, Name: "foo"},
						},
					}}, nil)
				s.On("Get", mock.Anything).
					Return(nil, &store.ErrNotFound{})
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "GetRole store err",
			attrs: &authorization.Attributes{
				Namespace: "acme",
				User: corev2.User{
					Username: "foo",
				},
			},
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("List", clusterRoleBindingListRequest, mock.Anything).
					Return(wrapList[*corev2.ClusterRoleBinding](nil), nil)
				s.On("List", roleBindingListRequest, mock.Anything).
					Return(wrapList[*corev2.RoleBinding]{{
						RoleRef: corev2.RoleRef{
							Type: "Role",
							Name: "admin",
						},
						Subjects: []corev2.Subject{
							{Type: corev2.UserType, Name: "foo"},
						},
					}}, nil)
				s.On("Get", mock.Anything).
					Return(nil, &store.ErrNotFound{})
			},
			wantErr: true,
		},
		{
			name: "matching RoleBinding",
			attrs: &authorization.Attributes{
				Namespace: "acme",
				User: corev2.User{
					Username: "foo",
				},
				Verb:         "create",
				Resource:     "checks",
				ResourceName: "check-cpu",
			},
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("List", clusterRoleBindingListRequest, mock.Anything).
					Return(wrapList[*corev2.ClusterRoleBinding](nil), nil)

				s.On("List", roleBindingListRequest, mock.Anything).
					Return(wrapList[*corev2.RoleBinding]{{
						RoleRef: corev2.RoleRef{
							Type: "Role",
							Name: "admin",
						},
						Subjects: []corev2.Subject{
							{Type: corev2.UserType, Name: "foo"},
						},
					}}, nil)
				s.On("Get", mock.Anything).
					Return(wrapper[*corev2.Role]{Value: &corev2.Role{Rules: []corev2.Rule{
						{
							Verbs:         []string{"create"},
							Resources:     []string{"checks"},
							ResourceNames: []string{"check-cpu"},
						},
					}}}, nil)
			},
			want: true,
		},
		{
			name: "role bindings do not match cluster width resource request",
			attrs: &authorization.Attributes{
				User: corev2.User{
					Username: "foo",
				},
				Verb:     "list",
				Resource: "users",
			},
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("List", clusterRoleBindingListRequest, mock.Anything).
					Return(wrapList[*corev2.ClusterRoleBinding](nil), nil)

				s.On("List", roleBindingListRequest, mock.Anything).
					Return(wrapList[*corev2.RoleBinding]{{
						RoleRef: corev2.RoleRef{
							Type: "ClusterRole",
							Name: "cluster-admin",
						},
						Subjects: []corev2.Subject{
							{Type: corev2.UserType, Name: "foo"},
						},
					}}, nil)
				s.On("Get", mock.Anything).
					Return(wrapper[*corev2.ClusterRole]{Value: &corev2.ClusterRole{Rules: []corev2.Rule{
						{
							Verbs:     []string{"*"},
							Resources: []string{"*"},
						},
					}}}, nil)
			},
			want: false,
		},
		{
			name: "role specified by role binding not found",
			attrs: &authorization.Attributes{
				Namespace: "acme",
				User: corev2.User{
					Username: "foo",
				},
				Verb:         "create",
				Resource:     "checks",
				ResourceName: "check-cpu",
			},
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("List", clusterRoleBindingListRequest, mock.Anything).
					Return(wrapList[*corev2.ClusterRoleBinding](nil), nil)

				s.On("List", roleBindingListRequest, mock.Anything).
					Return(wrapList[*corev2.RoleBinding]{{
						RoleRef: corev2.RoleRef{
							Type: "Role",
							Name: "admin",
						},
						Subjects: []corev2.Subject{
							{Type: corev2.UserType, Name: "foo"},
						},
					}}, nil)
				s.On("Get", mock.Anything).
					Return(wrapper[*corev2.Role]{Value: nil}, &store.ErrNotFound{})
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := &mockstore.V2MockStore{}
			tc.storeFunc(store)
			a := &Authorizer{
				Store: store,
			}

			got, err := a.Authorize(ctx, tc.attrs)
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
		user     corev2.User
		subjects []corev2.Subject
		want     bool
	}{
		{
			name: "not matching",
			user: corev2.User{Username: "foo"},
			subjects: []corev2.Subject{
				{Type: corev2.UserType, Name: "bar"},
				{Type: corev2.GroupType, Name: "foo"},
			},
			want: false,
		},
		{
			name: "matching via username",
			user: corev2.User{Username: "foo"},
			subjects: []corev2.Subject{
				{Type: corev2.UserType, Name: "bar"},
				{Type: corev2.UserType, Name: "foo"},
			},
			want: true,
		},
		{
			name: "matching via group",
			user: corev2.User{Username: "foo", Groups: []string{"acme"}},
			subjects: []corev2.Subject{
				{Type: corev2.GroupType, Name: "acme"},
			},
			want: true,
		},
		{
			name: "matching is sensitive to all characters",
			user: corev2.User{Username: "foo", Groups: []string{"foo bar"}},
			subjects: []corev2.Subject{
				{Type: corev2.GroupType, Name: "foo bar"},
			},
			want: true,
		},
		{
			name: "matching is sensitive to all characters bis",
			user: corev2.User{Username: "foo", Groups: []string{"foo bar"}},
			subjects: []corev2.Subject{
				{Type: corev2.GroupType, Name: "foobar"},
			},
			want: false,
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
		rule  corev2.Rule
		want  bool
	}{
		{
			name: "verb does not match",
			attrs: &authorization.Attributes{
				Verb: "create",
			},
			rule: corev2.Rule{
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
			rule: corev2.Rule{
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
			rule: corev2.Rule{
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
			rule: corev2.Rule{
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

func TestVisitRulesFor(t *testing.T) {
	attrs := &authorization.Attributes{
		Namespace: "acme",
		User: corev2.User{
			Username: "foo",
		},
		Verb:         "create,delete",
		Resource:     "checks",
		ResourceName: "check-cpu",
	}
	stor := &mockstore.V2MockStore{}
	a := &Authorizer{
		Store: stor,
	}
	ctx := store.NamespaceContext(context.Background(), "acme")
	clusterRoleBindingListRequest := storev2.NewResourceRequest(ctx, "acme", "", new(corev2.ClusterRoleBinding).StoreName())
	roleBindingListRequest := storev2.NewResourceRequest(ctx, "acme", "", new(corev2.RoleBinding).StoreName())
	roleRequest := storev2.NewResourceRequest(store.NamespaceContext(ctx, "acme"), "acme", "admin", new(corev2.Role).StoreName())
	clusterRoleRequest := storev2.NewResourceRequest(ctx, "acme", "admin", new(corev2.ClusterRole).StoreName())

	stor.On("List", clusterRoleBindingListRequest, mock.Anything).
		Return(wrapList[*corev2.ClusterRoleBinding]{{
			ObjectMeta: corev2.ObjectMeta{
				Namespace: "acme",
			},
			RoleRef: corev2.RoleRef{
				Type: "ClusterRole",
				Name: "admin",
			},
			Subjects: []corev2.Subject{
				{Type: corev2.UserType, Name: "foo"},
			},
		}}, nil)

	stor.On("List", roleBindingListRequest, mock.Anything).
		Return(wrapList[*corev2.RoleBinding]{{
			ObjectMeta: corev2.ObjectMeta{
				Namespace: "acme",
			},
			RoleRef: corev2.RoleRef{
				Type: "Role",
				Name: "admin",
			},
			Subjects: []corev2.Subject{
				{Type: corev2.UserType, Name: "foo"},
			},
		}}, nil)
	stor.On("Get", roleRequest).
		Return(wrapper[*corev2.Role]{Value: &corev2.Role{Rules: []corev2.Rule{
			{
				Verbs:         []string{"create"},
				Resources:     []string{"checks"},
				ResourceNames: []string{"check-cpu"},
			},
		}}}, nil)
	stor.On("Get", clusterRoleRequest).
		Return(wrapper[*corev2.ClusterRole]{Value: &corev2.ClusterRole{Rules: []corev2.Rule{
			{
				Verbs:         []string{"delete"},
				Resources:     []string{"checks"},
				ResourceNames: []string{"check-cpu"},
			},
		}}}, nil)

	var rules []corev2.Rule

	a.VisitRulesFor(ctx, attrs, func(binding RoleBinding, rule corev2.Rule, err error) bool {
		if err != nil {
			t.Fatal(err)
			return false
		}
		rules = append(rules, rule)
		return true
	})

	if got, want := len(rules), 2; got != want {
		t.Fatalf("wrong number of rules: got %d, want %d", got, want)
	}
}
