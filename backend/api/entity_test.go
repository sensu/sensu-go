package api

import (
	"context"
	"errors"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

var defaultEntity = corev2.FixtureEntity("default")

func TestListEntities(t *testing.T) {
	tests := []struct {
		Name       string
		Ctx        func() context.Context
		Store      func() store.Store
		EventStore func() store.EventStore
		Auth       func() authorization.Authorizer
		Exp        []*corev2.Entity
		ExpErr     bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			Store: func() store.Store {
				store := new(mockstore.MockStore)
				return store
			},
			EventStore: func() store.EventStore {
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:   "core",
							APIVersion: "v2",
							Namespace:  "default",
							Resource:   "entities",
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:   "core",
							APIVersion: "v2",
							Namespace:  "default",
							Resource:   "entities",
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
				store.On("ListResources", mock.Anything, (&corev2.Entity{}).StorePrefix(), mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(2).(*[]*corev2.Entity)
					*arg = []*corev2.Entity{defaultEntity}
				}).Return(nil)
				return store
			},
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:   "core",
							APIVersion: "v2",
							Namespace:  "default",
							Resource:   "entities",
							UserName:   "legit",
							Verb:       "list",
						}: true,
					},
				}
				return auth
			},
			Exp: []*corev2.Entity{defaultEntity},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			store := test.Store()
			eventStore := test.EventStore()
			auth := test.Auth()
			client := NewEntityClient(store, eventStore, auth)
			entities, err := client.ListEntities(ctx)
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
			if got, want := entities, test.Exp; !reflect.DeepEqual(got, want) {
				t.Fatalf("bad entities: got %v, want %v", got, want)
			}
		})
	}
}

func TestGetEntity(t *testing.T) {
	tests := []struct {
		Name       string
		Ctx        func() context.Context
		Store      func() store.Store
		EventStore func() store.EventStore
		Auth       func() authorization.Authorizer
		Exp        *corev2.Entity
		ExpErr     bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			Store: func() store.Store {
				store := new(mockstore.MockStore)
				return store
			},
			EventStore: func() store.EventStore {
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "entities",
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "entities",
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
					arg := args.Get(2).(*corev2.Entity)
					*arg = *defaultEntity
				}).Return(nil)
				return store
			},
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "entities",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "get",
						}: true,
					},
				}
				return auth
			},
			Exp: defaultEntity,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			store := test.Store()
			eventStore := test.EventStore()
			auth := test.Auth()
			client := NewEntityClient(store, eventStore, auth)
			entities, err := client.FetchEntity(ctx, "default")
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
			if got, want := entities, test.Exp; !reflect.DeepEqual(got, want) {
				t.Fatalf("bad entities: got %v, want %v", got, want)
			}
		})
	}
}

func TestCreateEntity(t *testing.T) {
	tests := []struct {
		Name       string
		Ctx        func() context.Context
		Store      func() store.Store
		EventStore func() store.EventStore
		Auth       func() authorization.Authorizer
		ExpErr     bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			Store: func() store.Store {
				store := new(mockstore.MockStore)
				return store
			},
			EventStore: func() store.EventStore {
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "entities",
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "entities",
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
				store.On("CreateResource", mock.Anything, defaultEntity).Return(nil)
				return store
			},
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "entities",
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
			eventStore := test.EventStore()
			auth := test.Auth()
			client := NewEntityClient(store, eventStore, auth)
			err := client.CreateEntity(ctx, defaultEntity)
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
		})
	}
}

func TestUpdateEntity(t *testing.T) {
	tests := []struct {
		Name       string
		Ctx        func() context.Context
		Store      func() store.Store
		EventStore func() store.EventStore
		Auth       func() authorization.Authorizer
		ExpErr     bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			Store: func() store.Store {
				store := new(mockstore.MockStore)
				return store
			},
			EventStore: func() store.EventStore {
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "entities",
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "entities",
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
				store.On("CreateOrUpdateResource", mock.Anything, defaultEntity).Return(nil)
				return store
			},
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "entities",
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
			eventStore := test.EventStore()
			auth := test.Auth()
			client := NewEntityClient(store, eventStore, auth)
			err := client.UpdateEntity(ctx, defaultEntity)
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
		})
	}
}

func TestDeleteEntity(t *testing.T) {
	tests := []struct {
		Name       string
		Ctx        func() context.Context
		Store      func() store.Store
		EventStore func() store.EventStore
		Auth       func() authorization.Authorizer
		ExpErr     bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			Store: func() store.Store {
				store := new(mockstore.MockStore)
				return store
			},
			EventStore: func() store.EventStore {
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "entities",
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "entities",
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
				store.On("DeleteResource", mock.Anything, "entities", "default").Return(nil)
				return store
			},
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				store.On("GetEventsByEntity", mock.Anything, "default", mock.Anything).Return([]*corev2.Event{corev2.FixtureEvent("default", "default")}, nil)
				store.On("DeleteEventByEntityCheck", mock.Anything, "default", "default").Return(nil)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "entities",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "delete",
						}: true,
					},
				}
				return auth
			},
		},
		{
			Name: "eventstore error",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "legit", nil)
			},
			Store: func() store.Store {
				store := new(mockstore.MockStore)
				store.On("DeleteResource", mock.Anything, "entities", "default").Return(nil)
				return store
			},
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				store.On("GetEventsByEntity", mock.Anything, "default", mock.Anything).Return([]*corev2.Event{}, errors.New("error"))
				store.On("DeleteEventByEntityCheck", mock.Anything, "default", "default").Return(nil)
				return store
			},
			ExpErr: true,
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "entities",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "delete",
						}: true,
					},
				}
				return auth
			},
		},
		{
			Name: "event store error 2 (ignore error)",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "legit", nil)
			},
			Store: func() store.Store {
				store := new(mockstore.MockStore)
				store.On("DeleteResource", mock.Anything, "entities", "default").Return(nil)
				return store
			},
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				store.On("GetEventsByEntity", mock.Anything, "default", mock.Anything).Return([]*corev2.Event{corev2.FixtureEvent("default", "default")}, nil)
				store.On("DeleteEventByEntityCheck", mock.Anything, "default", "default").Return(errors.New("error"))
				return store
			},
			ExpErr: false,
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "entities",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "delete",
						}: true,
					},
				}
				return auth
			},
		},
		{
			Name: "event without check",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "legit", nil)
			},
			Store: func() store.Store {
				store := new(mockstore.MockStore)
				store.On("DeleteResource", mock.Anything, "entities", "default").Return(nil)
				return store
			},
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				store.On("GetEventsByEntity", mock.Anything, "default", mock.Anything).Return([]*corev2.Event{&corev2.Event{}}, nil)
				return store
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "entities",
							ResourceName: "default",
							UserName:     "legit",
							Verb:         "delete",
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
			eventStore := test.EventStore()
			auth := test.Auth()
			client := NewEntityClient(store, eventStore, auth)
			err := client.DeleteEntity(ctx, "default")
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
		})
	}
}
