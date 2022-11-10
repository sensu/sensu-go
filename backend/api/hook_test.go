package api

import (
	"context"
	"reflect"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

var defaultHookConfig = corev2.FixtureHookConfig("default")

func TestListHookConfigs(t *testing.T) {
	tests := []struct {
		Name   string
		Ctx    func() context.Context
		Store  func() store.Store
		Auth   func() authorization.Authorizer
		Exp    []*corev2.HookConfig
		ExpErr bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			Store: func() store.Store {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				return &rbac.Authorizer{}
			},
			ExpErr: true,
		},
		{
			Name: "wrong user",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() store.Store {
				return new(mockstore.MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:   "core",
							APIVersion: "v2",
							Namespace:  "default",
							Resource:   "hooks",
							UserName:   "legit",
							Verb:       "list",
						}: true,
					},
				}
				return auth
			},
			ExpErr: true,
		},
		{
			Name: "right user, wrong perms",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() store.Store {
				return new(mockstore.MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:   "core",
							APIVersion: "v2",
							Namespace:  "default",
							Resource:   "hooks",
							UserName:   "legit",
							Verb:       "create",
						}: true,
					},
				}
				return auth
			},
			ExpErr: true,
		},
		{
			Name: "good auth",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "legit", nil)
			},
			Store: func() store.Store {
				store := new(mockstore.MockStore)
				store.On("ListResources", mock.Anything, (&corev2.HookConfig{}).StorePrefix(), mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(2).(*[]*corev2.HookConfig)
					*arg = []*corev2.HookConfig{defaultHookConfig}
				}).Return(nil)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:   "core",
							APIVersion: "v2",
							Namespace:  "default",
							Resource:   "hooks",
							UserName:   "legit",
							Verb:       "list",
						}: true,
					},
				}
				return auth
			},
			Exp: []*corev2.HookConfig{defaultHookConfig},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			store := test.Store()
			auth := test.Auth()
			client := NewHookConfigClient(store, auth)
			hooks, err := client.ListHookConfigs(ctx)
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
			if got, want := hooks, test.Exp; !reflect.DeepEqual(got, want) {
				t.Fatalf("bad hooks: got %v, want %v", got, want)
			}
		})
	}
}

func TestGetHookConfig(t *testing.T) {
	tests := []struct {
		Name   string
		Ctx    func() context.Context
		Store  func() store.Store
		Auth   func() authorization.Authorizer
		Exp    *corev2.HookConfig
		ExpErr bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			Store: func() store.Store {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				return &rbac.Authorizer{}
			},
			ExpErr: true,
		},
		{
			Name: "wrong user",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() store.Store {
				return new(mockstore.MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "hooks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "get",
						}: true,
					},
				}
				return auth
			},
			ExpErr: true,
		},
		{
			Name: "right user, wrong perms",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() store.Store {
				return new(mockstore.MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "hooks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "create",
						}: true,
					},
				}
				return auth
			},
			ExpErr: true,
		},
		{
			Name: "good auth",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "legit", nil)
			},
			Store: func() store.Store {
				store := new(mockstore.MockStore)
				store.On("GetResource", mock.Anything, "default", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(2).(*corev2.HookConfig)
					*arg = *defaultHookConfig
				}).Return(nil)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "hooks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "get",
						}: true,
					},
				}
				return auth
			},
			Exp: defaultHookConfig,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			store := test.Store()
			auth := test.Auth()
			client := NewHookConfigClient(store, auth)
			hooks, err := client.FetchHookConfig(ctx, "default")
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
			if got, want := hooks, test.Exp; !reflect.DeepEqual(got, want) {
				t.Fatalf("bad hooks: got %v, want %v", got, want)
			}
		})
	}
}

func TestCreateHookConfig(t *testing.T) {
	tests := []struct {
		Name   string
		Ctx    func() context.Context
		Store  func() store.Store
		Auth   func() authorization.Authorizer
		ExpErr bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			Store: func() store.Store {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				return &rbac.Authorizer{}
			},
			ExpErr: true,
		},
		{
			Name: "wrong user",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() store.Store {
				return new(mockstore.MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "hooks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "create",
						}: true,
					},
				}
				return auth
			},
			ExpErr: true,
		},
		{
			Name: "right user, wrong perms",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() store.Store {
				return new(mockstore.MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "hooks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "get",
						}: true,
					},
				}
				return auth
			},
			ExpErr: true,
		},
		{
			Name: "good auth",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "legit", nil)
			},
			Store: func() store.Store {
				store := new(mockstore.MockStore)
				store.On("CreateResource", mock.Anything, defaultHookConfig).Return(nil)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "hooks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "create",
						}: true,
					},
				}
				return auth
			},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			store := test.Store()
			auth := test.Auth()
			client := NewHookConfigClient(store, auth)
			err := client.CreateHookConfig(ctx, defaultHookConfig)
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
		})
	}
}

func TestUpdateHookConfig(t *testing.T) {
	tests := []struct {
		Name   string
		Ctx    func() context.Context
		Store  func() store.Store
		Auth   func() authorization.Authorizer
		ExpErr bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			Store: func() store.Store {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				return &rbac.Authorizer{}
			},
			ExpErr: true,
		},
		{
			Name: "wrong user",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() store.Store {
				return new(mockstore.MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "hooks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "update",
						}: true,
					},
				}
				return auth
			},
			ExpErr: true,
		},
		{
			Name: "right user, wrong perms",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() store.Store {
				return new(mockstore.MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "hooks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "get",
						}: true,
					},
				}
				return auth
			},
			ExpErr: true,
		},
		{
			Name: "good auth",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "legit", nil)
			},
			Store: func() store.Store {
				store := new(mockstore.MockStore)
				store.On("CreateOrUpdateResource", mock.Anything, defaultHookConfig).Return(nil)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "hooks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "update",
						}: true,
					},
				}
				return auth
			},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			store := test.Store()
			auth := test.Auth()
			client := NewHookConfigClient(store, auth)
			err := client.UpdateHookConfig(ctx, defaultHookConfig)
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
		})
	}
}

func TestDeleteHookConfig(t *testing.T) {
	tests := []struct {
		Name   string
		Ctx    func() context.Context
		Store  func() store.Store
		Auth   func() authorization.Authorizer
		Exp    *corev2.HookConfig
		ExpErr bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			Store: func() store.Store {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				return &rbac.Authorizer{}
			},
			ExpErr: true,
		},
		{
			Name: "wrong user",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() store.Store {
				return new(mockstore.MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "hooks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "delete",
						}: true,
					},
				}
				return auth
			},
			ExpErr: true,
		},
		{
			Name: "right user, wrong perms",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			Store: func() store.Store {
				return new(mockstore.MockStore)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "hooks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "create",
						}: true,
					},
				}
				return auth
			},
			ExpErr: true,
		},
		{
			Name: "good auth",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "legit", nil)
			},
			Store: func() store.Store {
				store := new(mockstore.MockStore)
				store.On("DeleteResource", mock.Anything, "hooks", "default").Return(nil)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "hooks",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "delete",
						}: true,
					},
				}
				return auth
			},
			Exp: defaultHookConfig,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			store := test.Store()
			auth := test.Auth()
			client := NewHookConfigClient(store, auth)
			err := client.DeleteHookConfig(ctx, "default")
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
		})
	}
}
