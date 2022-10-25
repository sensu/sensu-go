package api

import (
	"context"
	"reflect"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

var defaultCheck = corev2.FixtureCheckConfig("default")

type mockCheckController struct {
	mock.Mock
}

func (m *mockCheckController) AddCheckHook(ctx context.Context, check string, hook corev2.HookList) error {
	return m.Called(ctx, check, hook).Error(0)
}

func (m *mockCheckController) RemoveCheckHook(ctx context.Context, check string, hookType string, hookName string) error {
	return m.Called(ctx, check, hookType, hookName).Error(0)
}

func (m *mockCheckController) QueueAdhocRequest(ctx context.Context, check string, req *corev2.AdhocRequest) error {
	return m.Called(ctx, check, req).Error(0)
}

func TestListChecks(t *testing.T) {
	tests := []struct {
		Name       string
		Ctx        func() context.Context
		Store      func() storev2.Interface
		Auth       func() authorization.Authorizer
		Controller func() CheckController
		Exp        []*corev2.CheckConfig
		ExpErr     bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			Store: func() storev2.Interface {
				store := new(mockstore.V2MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				return &rbac.Authorizer{}
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			ExpErr: true,
		},
		{
			Name: "wrong user",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() storev2.Interface {
				return new(mockstore.V2MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:   "core",
							APIVersion: "v2",
							Namespace:  "default",
							Resource:   "checks",
							UserName:   "legit",
							Verb:       "list",
						}: true,
					},
				}
				return auth
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			ExpErr: true,
		},
		{
			Name: "right user, wrong perms",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() storev2.Interface {
				return new(mockstore.V2MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:   "core",
							APIVersion: "v2",
							Namespace:  "default",
							Resource:   "checks",
							UserName:   "legit",
							Verb:       "create",
						}: true,
					},
				}
				return auth
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			ExpErr: true,
		},
		{
			Name: "good auth",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "legit", nil)
			},
			Store: func() storev2.Interface {
				store := new(mockstore.V2MockStore)
				wrapper, _ := wrap.Resource(defaultCheck)
				list := wrap.List{wrapper}
				store.On("List", mock.Anything, mock.Anything, mock.Anything).Return(list, nil)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:   "core",
							APIVersion: "v2",
							Namespace:  "default",
							Resource:   "checks",
							UserName:   "legit",
							Verb:       "list",
						}: true,
					},
				}
				return auth
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			Exp: []*corev2.CheckConfig{defaultCheck},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			store := test.Store()
			auth := test.Auth()
			ctrl := test.Controller()
			client := NewCheckClient(store, ctrl, auth)
			checks, err := client.ListChecks(ctx)
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
			if got, want := len(checks), len(test.Exp); got != want {
				t.Fatalf("bad checks: got %v, want %v", got, want)
			}
		})
	}
}

func TestGetCheck(t *testing.T) {
	tests := []struct {
		Name       string
		Ctx        func() context.Context
		Store      func() storev2.Interface
		Auth       func() authorization.Authorizer
		Controller func() CheckController
		Exp        *corev2.CheckConfig
		ExpErr     bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			Store: func() storev2.Interface {
				store := new(mockstore.V2MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				return &rbac.Authorizer{}
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			ExpErr: true,
		},
		{
			Name: "wrong user",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() storev2.Interface {
				return new(mockstore.V2MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "checks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "get",
						}: true,
					},
				}
				return auth
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			ExpErr: true,
		},
		{
			Name: "right user, wrong perms",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() storev2.Interface {
				return new(mockstore.V2MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "checks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "create",
						}: true,
					},
				}
				return auth
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			ExpErr: true,
		},
		{
			Name: "good auth",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "legit", nil)
			},
			Store: func() storev2.Interface {
				store := new(mockstore.V2MockStore)
				store.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.CheckConfig]{Value: defaultCheck}, nil)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "checks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "get",
						}: true,
					},
				}
				return auth
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			Exp: defaultCheck,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			store := test.Store()
			auth := test.Auth()
			ctrl := test.Controller()
			client := NewCheckClient(store, ctrl, auth)
			checks, err := client.FetchCheck(ctx, "default")
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
			if got, want := checks, test.Exp; !reflect.DeepEqual(got, want) {
				t.Fatalf("bad checks: got %v, want %v", got, want)
			}
		})
	}
}

func TestCreateCheck(t *testing.T) {
	tests := []struct {
		Name       string
		Ctx        func() context.Context
		Store      func() storev2.Interface
		Controller func() CheckController
		Auth       func() authorization.Authorizer
		ExpErr     bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			Store: func() storev2.Interface {
				store := new(mockstore.V2MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				return &rbac.Authorizer{}
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			ExpErr: true,
		},
		{
			Name: "wrong user",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() storev2.Interface {
				return new(mockstore.V2MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "checks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "create",
						}: true,
					},
				}
				return auth
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			ExpErr: true,
		},
		{
			Name: "right user, wrong perms",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() storev2.Interface {
				return new(mockstore.V2MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "checks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "get",
						}: true,
					},
				}
				return auth
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			ExpErr: true,
		},
		{
			Name: "good auth",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "legit", nil)
			},
			Store: func() storev2.Interface {
				store := new(mockstore.V2MockStore)
				store.On("CreateOrUpdate", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "checks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "create",
						}: true,
					},
				}
				return auth
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			store := test.Store()
			auth := test.Auth()
			ctrl := test.Controller()
			client := NewCheckClient(store, ctrl, auth)
			err := client.CreateCheck(ctx, defaultCheck)
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
		})
	}
}

func TestUpdateCheck(t *testing.T) {
	tests := []struct {
		Name       string
		Ctx        func() context.Context
		Store      func() storev2.Interface
		Controller func() CheckController
		Auth       func() authorization.Authorizer
		ExpErr     bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			Store: func() storev2.Interface {
				store := new(mockstore.V2MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				return &rbac.Authorizer{}
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			ExpErr: true,
		},
		{
			Name: "wrong user",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() storev2.Interface {
				return new(mockstore.V2MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "checks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "update",
						}: true,
					},
				}
				return auth
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			ExpErr: true,
		},
		{
			Name: "right user, wrong perms",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() storev2.Interface {
				return new(mockstore.V2MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "checks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "get",
						}: true,
					},
				}
				return auth
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			ExpErr: true,
		},
		{
			Name: "good auth",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "legit", nil)
			},
			Store: func() storev2.Interface {
				store := new(mockstore.V2MockStore)
				store.On("CreateOrUpdate", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "checks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "update",
						}: true,
					},
				}
				return auth
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			store := test.Store()
			auth := test.Auth()
			ctrl := test.Controller()
			client := NewCheckClient(store, ctrl, auth)
			err := client.UpdateCheck(ctx, defaultCheck)
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
		})
	}
}

func TestDeleteCheck(t *testing.T) {
	tests := []struct {
		Name       string
		Ctx        func() context.Context
		Store      func() storev2.Interface
		Controller func() CheckController
		Auth       func() authorization.Authorizer
		ExpErr     bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			Store: func() storev2.Interface {
				store := new(mockstore.V2MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				return &rbac.Authorizer{}
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			ExpErr: true,
		},
		{
			Name: "wrong user",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() storev2.Interface {
				return new(mockstore.V2MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "checks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "delete",
						}: true,
					},
				}
				return auth
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			ExpErr: true,
		},
		{
			Name: "right user, wrong perms",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() storev2.Interface {
				return new(mockstore.V2MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "checks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "get",
						}: true,
					},
				}
				return auth
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			ExpErr: true,
		},
		{
			Name: "good auth",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "legit", nil)
			},
			Store: func() storev2.Interface {
				store := new(mockstore.V2MockStore)
				store.On("Delete", mock.Anything, mock.Anything).Return(nil)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "checks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "delete",
						}: true,
					},
				}
				return auth
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			store := test.Store()
			auth := test.Auth()
			ctrl := test.Controller()
			client := NewCheckClient(store, ctrl, auth)
			err := client.DeleteCheck(ctx, "default")
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
		})
	}
}

func TestExecuteCheck(t *testing.T) {
	tests := []struct {
		Name       string
		Ctx        func() context.Context
		Store      func() storev2.Interface
		Controller func() CheckController
		Auth       func() authorization.Authorizer
		ExpErr     bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			Store: func() storev2.Interface {
				store := new(mockstore.V2MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				return &rbac.Authorizer{}
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			ExpErr: true,
		},
		{
			Name: "wrong user",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() storev2.Interface {
				return new(mockstore.V2MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "checks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "delete",
						}: true,
					},
				}
				return auth
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			ExpErr: true,
		},
		{
			Name: "right user, wrong perms",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() storev2.Interface {
				return new(mockstore.V2MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "checks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "get",
						}: true,
					},
				}
				return auth
			},
			Controller: func() CheckController {
				return new(mockCheckController)
			},
			ExpErr: true,
		},
		{
			Name: "good auth",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "legit", nil)
			},
			Store: func() storev2.Interface {
				store := new(mockstore.V2MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "checks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "create",
						}: true,
					},
				}
				return auth
			},
			Controller: func() CheckController {
				ctrl := new(mockCheckController)
				ctrl.On("QueueAdhocRequest", mock.Anything, "default", mock.Anything).Return(nil)
				return ctrl
			},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			store := test.Store()
			auth := test.Auth()
			ctrl := test.Controller()
			client := NewCheckClient(store, ctrl, auth)
			err := client.ExecuteCheck(ctx, "default", &corev2.AdhocRequest{})
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
		})
	}
}
