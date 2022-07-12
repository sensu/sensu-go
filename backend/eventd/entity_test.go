package eventd

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/storetest"
)

func TestCreateProxyEntity(t *testing.T) {
	type storeFunc func(*storetest.Store, *corev2.Event)
	var nilWrapper storev2.Wrapper

	tests := []struct {
		name           string
		event          *corev2.Event
		storeFunc      storeFunc
		wantEntityName string
		wantEntity     *corev2.Entity
		wantErr        bool
	}{
		{
			// We receive an event from entity "foo", already in the system.
			//
			// We expect to retrieve the EntityConfig and EntityState for "foo"
			// and the event to look like it came from "foo".
			name:  "entity exists",
			event: corev2.FixtureEvent("foo", "check-cpu"),
			storeFunc: func(s *storetest.Store, _ *corev2.Event) {
				state := corev3.FixtureEntityState("foo")
				config := corev3.FixtureEntityConfig("foo")

				stateReq := storev2.NewResourceRequestFromResource(state)
				configReq := storev2.NewResourceRequestFromResource(config)

				wConfig, err := storev2.WrapResource(config)
				if err != nil {
					t.Fatal(err)
				}

				wState, err := storev2.WrapResource(state)
				if err != nil {
					t.Fatal(err)
				}

				s.On("Get", mock.Anything, stateReq).Return(wState, nil)
				s.On("Get", mock.Anything, configReq).Return(wConfig, nil)
			},
			wantEntityName: "foo",
		},
		{
			// We receive an event from entity "foo", currently unknown in the
			// system.
			//
			// We expect an EntityConfig and an EntityState to be created for
			// "foo" and the event to be mutated appropriately.
			name: "entity does not exist",
			event: &corev2.Event{
				Check: corev2.FixtureCheck("check-cpu"),
				Entity: &corev2.Entity{
					ObjectMeta: corev2.ObjectMeta{
						Name:      "foo",
						Namespace: "default",
					},
					Subscriptions: []string{"linux"},
				},
			},
			storeFunc: func(s *storetest.Store, e *corev2.Event) {
				_, state := corev3.V2EntityToV3(e.Entity)
				stateReq := storev2.NewResourceRequestFromResource(state)

				config := corev3.FixtureEntityConfig("foo")
				configReq := storev2.NewResourceRequestFromResource(config)

				s.On("Get", mock.Anything, mock.AnythingOfType("v2.ResourceRequest")).
					Return(nilWrapper, &store.ErrNotFound{})

				// Assert that CreateOrUpdate() was called with the expected
				// request and wrapper
				s.On("CreateOrUpdate", mock.Anything, stateReq, mock.Anything).Return(nil)

				// Assert that CreateIfNotExists() was called with the expected
				// request and wrapper type
				// TODO(ccressent): can we do something more strict with the
				// matching?
				s.On("CreateIfNotExists", mock.Anything, configReq, mock.Anything).Return(nil)
			},
			wantEntity: &corev2.Entity{
				ObjectMeta: corev2.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				EntityClass:   "proxy",
				Subscriptions: []string{"linux", "entity:foo"},
			},
			wantEntityName: "foo",
		},
		{
			// We receive an event from entity "foo", but we encounter an error
			// while trying to find it in the store.
			//
			// We expect that error to be returned and the event to be unchanged.
			name:  "store error while getting the EntityConfig",
			event: corev2.FixtureEvent("foo", "check-cpu"),
			storeFunc: func(s *storetest.Store, _ *corev2.Event) {
				config := corev3.FixtureEntityConfig("foo")
				configReq := storev2.NewResourceRequestFromResource(config)

				s.On("Get", mock.Anything, configReq).Return(nilWrapper, errors.New("error"))
			},
			wantEntityName: "foo",
			wantErr:        true,
		},
		{
			// We receive an event from entity "foo", on behalf of a proxy
			// entity "bar" that already exists in the system.
			//
			// We expect to retrieve the EntityConfig and EntityState for "bar"
			// and mutate the event to make it look like it came from that proxy
			// entity.
			name: "proxy entity already exists",
			event: &corev2.Event{
				Check: &corev2.Check{
					ProxyEntityName: "bar",
				},
				Entity: corev2.FixtureEntity("foo"),
			},
			storeFunc: func(s *storetest.Store, _ *corev2.Event) {
				state := corev3.FixtureEntityState("bar")
				config := corev3.FixtureEntityConfig("bar")

				stateReq := storev2.NewResourceRequestFromResource(state)
				configReq := storev2.NewResourceRequestFromResource(config)

				wState, err := storev2.WrapResource(state)
				if err != nil {
					t.Fatal(err)
				}

				wConfig, err := storev2.WrapResource(config)
				if err != nil {
					t.Fatal(err)
				}

				s.On("Get", mock.Anything, stateReq).Return(wState, nil)
				s.On("Get", mock.Anything, configReq).Return(wConfig, nil)
			},
			wantEntityName: "bar",
		},
		{
			// We receive an event from entity "foo", on behalf of a proxy
			// entity "bar" that is currently unknown to the system.
			//
			// We expect to create and store a new EntityConfig and a new
			// EntityState for that new proxy entity "bar" and mutate the event
			// to make it look like it came from that proxy entity.
			name: "proxy entity does not exist yet",
			event: &corev2.Event{
				Check: &corev2.Check{
					ProxyEntityName: "bar",
				},
				Entity: corev2.FixtureEntity("foo"),
			},
			storeFunc: func(s *storetest.Store, _ *corev2.Event) {
				state := corev3.NewEntityState("default", "bar")
				config := corev3.FixtureEntityConfig("bar")

				configReq := storev2.NewResourceRequestFromResource(config)
				stateReq := storev2.NewResourceRequestFromResource(state)

				s.On("Get", mock.Anything, configReq).Return(nilWrapper, &store.ErrNotFound{})

				// Assert that CreateOrUpdate() was called with the expected
				// request and wrapper
				s.On("CreateOrUpdate", mock.Anything, stateReq, mock.Anything).Return(nil)

				// Assert that CreateIfNotExists() was called with the expected
				// request and wrapper type
				// TODO(ccressent): can we do something more strict with the
				// matching?
				s.On("CreateIfNotExists", mock.Anything, configReq, mock.Anything).Return(nil)
			},
			wantEntityName: "bar",
		},
		{
			// We receive an event from entity "foo", on behalf of a proxy
			// entity "bar" that is currently unknown to the system and
			// encounter an error while creating that new proxy entity "bar".
			//
			// We expect that error to be returned and the event mutated to make
			// it look like it came from that proxy entity.
			name: "store error while creating new proxy entity",
			event: &corev2.Event{
				Check: &corev2.Check{
					ProxyEntityName: "bar",
				},
				Entity: corev2.FixtureEntity("foo"),
			},
			storeFunc: func(s *storetest.Store, e *corev2.Event) {
				e.Entity.ObjectMeta.Name = "bar"
				_, state := corev3.V2EntityToV3(e.Entity)
				stateReq := storev2.NewResourceRequestFromResource(state)

				config := corev3.FixtureEntityConfig("bar")
				configReq := storev2.NewResourceRequestFromResource(config)

				s.On("Get", mock.Anything, configReq).
					Return(nilWrapper, &store.ErrNotFound{})

				s.On("CreateOrUpdate", mock.Anything, stateReq, mock.Anything).
					Return(errors.New("error"))
			},
			wantEntityName: "bar",
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &storetest.Store{}
			if tt.storeFunc != nil {
				tt.storeFunc(store, tt.event)
			}
			defer store.AssertExpectations(t)

			if err := createProxyEntity(tt.event, store); (err != nil) != tt.wantErr {
				t.Errorf("createProxyEntity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.event.Entity.Name, tt.wantEntityName) {
				t.Errorf("createProxyEntity() entity name = %v, want %v", tt.event.Entity.Name, tt.wantEntityName)
			}
			if tt.wantEntity != nil && !tt.event.Entity.Equal(tt.wantEntity) {
				t.Errorf("createProxyEntity() entity = %v, want %v", tt.event.Entity, tt.wantEntity)
			}
		})
	}
}
