package api

import (
	"context"
	"reflect"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockbus"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

var defaultEvent = corev2.FixtureEvent("default", "default")

func TestListEvents(t *testing.T) {
	tests := []struct {
		Name       string
		Ctx        func() context.Context
		EventStore func() store.EventStore
		Bus        func() messaging.MessageBus
		Auth       func() authorization.Authorizer
		Exp        []*corev2.Event
		ExpErr     bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				return &rbac.Authorizer{}
			},
			Bus: func() messaging.MessageBus {
				return new(mockbus.MockBus)
			},
			ExpErr: true,
		},
		{
			Name: "wrong user",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Bus: func() messaging.MessageBus {
				return new(mockbus.MockBus)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:   "core",
							APIVersion: "v2",
							Namespace:  "default",
							Resource:   "events",
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Bus: func() messaging.MessageBus {
				return new(mockbus.MockBus)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:   "core",
							APIVersion: "v2",
							Namespace:  "default",
							Resource:   "events",
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				store.On("GetEvents", mock.Anything, mock.Anything).Return([]*corev2.Event{defaultEvent}, nil)
				return store
			},
			Bus: func() messaging.MessageBus {
				return new(mockbus.MockBus)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:   "core",
							APIVersion: "v2",
							Namespace:  "default",
							Resource:   "events",
							UserName:   "legit",
							Verb:       "list",
						}: true,
					},
				}
				return auth
			},
			Exp: []*corev2.Event{defaultEvent},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			eventStore := test.EventStore()
			auth := test.Auth()
			bus := test.Bus()
			client := NewEventClient(eventStore, auth, bus)
			events, err := client.ListEvents(ctx, &store.SelectionPredicate{})
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
			if got, want := events, test.Exp; !reflect.DeepEqual(got, want) {
				t.Fatalf("bad events: got %v, want %v", got, want)
			}
		})
	}
}

func TestGetEvent(t *testing.T) {
	tests := []struct {
		Name       string
		Ctx        func() context.Context
		EventStore func() store.EventStore
		Bus        func() messaging.MessageBus
		Auth       func() authorization.Authorizer
		Exp        *corev2.Event
		ExpErr     bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Bus: func() messaging.MessageBus {
				return new(mockbus.MockBus)
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Bus: func() messaging.MessageBus {
				return new(mockbus.MockBus)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "events",
							ResourceName: "default:default",
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Bus: func() messaging.MessageBus {
				return new(mockbus.MockBus)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "events",
							ResourceName: "default:default",
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				store.On("GetEventByEntityCheck", mock.Anything, "default", "default").Return(defaultEvent, nil)
				return store
			},
			Bus: func() messaging.MessageBus {
				return new(mockbus.MockBus)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "events",
							ResourceName: "default:default",
							UserName:     "legit",
							Verb:         "get",
						}: true,
					},
				}
				return auth
			},
			Exp: defaultEvent,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			eventStore := test.EventStore()
			auth := test.Auth()
			bus := test.Bus()
			client := NewEventClient(eventStore, auth, bus)
			events, err := client.FetchEvent(ctx, "default", "default")
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
			if got, want := events, test.Exp; !reflect.DeepEqual(got, want) {
				t.Fatalf("bad events: got %v, want %v", got, want)
			}
		})
	}
}

func TestUpdateEvent(t *testing.T) {
	tests := []struct {
		Name       string
		Ctx        func() context.Context
		EventStore func() store.EventStore
		Bus        func() messaging.MessageBus
		Auth       func() authorization.Authorizer
		ExpErr     bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Bus: func() messaging.MessageBus {
				return new(mockbus.MockBus)
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Bus: func() messaging.MessageBus {
				return new(mockbus.MockBus)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "events",
							ResourceName: "default:default",
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
				return contextWithUser(defaultContext(), "legit", nil)
			},
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Bus: func() messaging.MessageBus {
				bus := new(mockbus.MockBus)
				bus.On("Publish", messaging.TopicEventRaw, defaultEvent).Return(nil)
				return bus
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "events",
							ResourceName: "default:default",
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				store.On("CreateOrUpdateResource", mock.Anything, defaultEvent).Return(nil)
				return store
			},
			Bus: func() messaging.MessageBus {
				bus := new(mockbus.MockBus)
				bus.On("Publish", messaging.TopicEventRaw, defaultEvent).Return(nil)
				return bus
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:   "core",
							APIVersion: "v2",
							Namespace:  "default",
							Resource:   "events",
							UserName:   "legit",
							Verb:       "update",
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
			eventStore := test.EventStore()
			auth := test.Auth()
			bus := test.Bus()
			client := NewEventClient(eventStore, auth, bus)
			err := client.UpdateEvent(ctx, defaultEvent)
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
		})
	}
}

func TestDeleteEvent(t *testing.T) {
	tests := []struct {
		Name       string
		Ctx        func() context.Context
		EventStore func() store.EventStore
		Bus        func() messaging.MessageBus
		Auth       func() authorization.Authorizer
		ExpErr     bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Bus: func() messaging.MessageBus {
				return new(mockbus.MockBus)
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
			Bus: func() messaging.MessageBus {
				return new(mockbus.MockBus)
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
							Resource:     "events",
							ResourceName: "default:default",
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Bus: func() messaging.MessageBus {
				return new(mockbus.MockBus)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "events",
							ResourceName: "default:default",
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				store.On("DeleteEventByEntityCheck", mock.Anything, "default", "default").Return(nil)
				return store
			},
			Bus: func() messaging.MessageBus {
				return new(mockbus.MockBus)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:     "core",
							APIVersion:   "v2",
							Namespace:    "default",
							Resource:     "events",
							ResourceName: "default:default",
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
			eventStore := test.EventStore()
			auth := test.Auth()
			bus := test.Bus()
			client := NewEventClient(eventStore, auth, bus)
			err := client.DeleteEvent(ctx, "default", "default")
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
		})
	}
}

func TestListEventsByEntity(t *testing.T) {
	tests := []struct {
		Name       string
		Ctx        func() context.Context
		EventStore func() store.EventStore
		Bus        func() messaging.MessageBus
		Auth       func() authorization.Authorizer
		Exp        []*corev2.Event
		ExpErr     bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Auth: func() authorization.Authorizer {
				return &rbac.Authorizer{}
			},
			Bus: func() messaging.MessageBus {
				return new(mockbus.MockBus)
			},
			ExpErr: true,
		},
		{
			Name: "wrong user",
			Ctx: func() context.Context {
				return contextWithUser(defaultContext(), "haxor", nil)
			},
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Bus: func() messaging.MessageBus {
				return new(mockbus.MockBus)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:   "core",
							APIVersion: "v2",
							Namespace:  "default",
							Resource:   "events",
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				return store
			},
			Bus: func() messaging.MessageBus {
				return new(mockbus.MockBus)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:   "core",
							APIVersion: "v2",
							Namespace:  "default",
							Resource:   "events",
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
			EventStore: func() store.EventStore {
				store := new(mockstore.MockStore)
				store.On("GetEventsByEntity", mock.Anything, "ralphie", mock.Anything).Return([]*corev2.Event{defaultEvent}, nil)
				return store
			},
			Bus: func() messaging.MessageBus {
				return new(mockbus.MockBus)
			},
			Auth: func() authorization.Authorizer {
				auth := &mockAuth{
					attrs: map[authorization.AttributesKey]bool{
						authorization.AttributesKey{
							APIGroup:   "core",
							APIVersion: "v2",
							Namespace:  "default",
							Resource:   "events",
							UserName:   "legit",
							Verb:       "list",
						}: true,
					},
				}
				return auth
			},
			Exp: []*corev2.Event{defaultEvent},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			eventStore := test.EventStore()
			auth := test.Auth()
			bus := test.Bus()
			client := NewEventClient(eventStore, auth, bus)
			events, err := client.ListEventsByEntity(ctx, "ralphie", &store.SelectionPredicate{})
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
			if got, want := events, test.Exp; !reflect.DeepEqual(got, want) {
				t.Fatalf("bad events: got %v, want %v", got, want)
			}
		})
	}
}
