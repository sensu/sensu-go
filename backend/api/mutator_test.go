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

var defaultMutator = corev2.FixtureMutator("default")

func TestListMutators(t *testing.T) {
	tests := []struct {
		Name   string
		Ctx    func() context.Context
		Store  func() store.Store
		Auth   func() authorization.Authorizer
		Exp    []*corev2.Mutator
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
							Resource:   "mutators",
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
							Resource:   "mutators",
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
				store.On("ListResources", mock.Anything, (&corev2.Mutator{}).StorePrefix(), mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(2).(*[]*corev2.Mutator)
					*arg = []*corev2.Mutator{defaultMutator}
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
							Resource:   "mutators",
							UserName:   "legit",
							Verb:       "list",
						}: true,
					},
				}
				return auth
			},
			Exp: []*corev2.Mutator{defaultMutator},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			store := test.Store()
			auth := test.Auth()
			client := NewMutatorClient(store, auth)
			mutators, err := client.ListMutators(ctx)
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
			if got, want := mutators, test.Exp; !reflect.DeepEqual(got, want) {
				t.Fatalf("bad mutators: got %v, want %v", got, want)
			}
		})
	}
}

func TestGetMutator(t *testing.T) {
	tests := []struct {
		Name   string
		Ctx    func() context.Context
		Store  func() store.Store
		Auth   func() authorization.Authorizer
		Exp    *corev2.Mutator
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
							Resource:     "mutators",
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
							Resource:     "mutators",
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
					arg := args.Get(2).(*corev2.Mutator)
					*arg = *defaultMutator
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
							Resource:     "mutators",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "get",
						}: true,
					},
				}
				return auth
			},
			Exp: defaultMutator,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			store := test.Store()
			auth := test.Auth()
			client := NewMutatorClient(store, auth)
			mutators, err := client.FetchMutator(ctx, "default")
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
			if got, want := mutators, test.Exp; !reflect.DeepEqual(got, want) {
				t.Fatalf("bad mutators: got %v, want %v", got, want)
			}
		})
	}
}

func TestCreateMutator(t *testing.T) {
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
							Resource:     "mutators",
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
							Resource:     "mutators",
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
				store.On("CreateResource", mock.Anything, defaultMutator).Return(nil)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "mutators",
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
			client := NewMutatorClient(store, auth)
			err := client.CreateMutator(ctx, defaultMutator)
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
		})
	}
}

func TestUpdateMutator(t *testing.T) {
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
							Resource:     "mutators",
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
							Resource:     "mutators",
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
				store.On("CreateOrUpdateResource", mock.Anything, defaultMutator).Return(nil)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "mutators",
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
			client := NewMutatorClient(store, auth)
			err := client.UpdateMutator(ctx, defaultMutator)
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
		})
	}
}

func TestDeleteMutator(t *testing.T) {
	tests := []struct {
		Name   string
		Ctx    func() context.Context
		Store  func() store.Store
		Auth   func() authorization.Authorizer
		Exp    *corev2.Mutator
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
							Resource:     "mutators",
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
							Resource:     "mutators",
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
				store.On("DeleteResource", mock.Anything, "mutators", "default").Return(nil)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "mutators",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "delete",
						}: true,
					},
				}
				return auth
			},
			Exp: defaultMutator,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			store := test.Store()
			auth := test.Auth()
			client := NewMutatorClient(store, auth)
			err := client.DeleteMutator(ctx, "default")
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
		})
	}
}
