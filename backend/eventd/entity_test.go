package eventd

import (
	"errors"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
)

func TestCreateProxyEntity(t *testing.T) {
	type storeFunc func(*mockstore.MockStore)
	var nilEntity *types.Entity

	tests := []struct {
		name           string
		event          *corev2.Event
		storeFunc      storeFunc
		wantEntityName string
		wantEntity     *corev2.Entity
		wantErr        bool
	}{
		{
			name:  "entity exists",
			event: corev2.FixtureEvent("foo", "check-cpu"),
			storeFunc: func(store *mockstore.MockStore) {
				store.On("GetEntityByName", mock.Anything, "foo").
					Return(corev2.FixtureEntity("foo"), nil)
			},
			wantEntityName: "foo",
		},
		{
			name:  "entity does not exist",
			event: corev2.FixtureEvent("foo", "check-cpu"),
			storeFunc: func(store *mockstore.MockStore) {
				store.On("GetEntityByName", mock.Anything, "foo").
					Return(nilEntity, nil)
				store.On("UpdateEntity", mock.Anything, mock.AnythingOfType("*v2.Entity")).
					Return(nil)
			},
			wantEntityName: "foo",
		},
		{
			name:  "store error while getting an entity",
			event: corev2.FixtureEvent("foo", "check-cpu"),
			storeFunc: func(store *mockstore.MockStore) {
				store.On("GetEntityByName", mock.Anything, "foo").
					Return(&corev2.Entity{}, errors.New("error"))
			},
			wantEntityName: "foo",
			wantErr:        true,
		},
		{
			name: "proxy entity exists",
			event: &corev2.Event{
				Check: &corev2.Check{
					ProxyEntityName: "bar",
				},
				Entity: corev2.FixtureEntity("foo"),
			},
			storeFunc: func(store *mockstore.MockStore) {
				store.On("GetEntityByName", mock.Anything, "bar").
					Return(corev2.FixtureEntity("bar"), nil)
			},
			wantEntityName: "bar",
		},
		{
			name: "proxy entity does not exist",
			event: &corev2.Event{
				Check: &corev2.Check{
					ProxyEntityName: "bar",
				},
				Entity: corev2.FixtureEntity("foo"),
			},
			storeFunc: func(store *mockstore.MockStore) {
				store.On("GetEntityByName", mock.Anything, "bar").
					Return(nilEntity, nil)
				store.On("UpdateEntity", mock.Anything, mock.AnythingOfType("*v2.Entity")).
					Return(nil)
			},
			wantEntityName: "bar",
		},
		{
			name: "store error while updating entity",
			event: &corev2.Event{
				Check: &corev2.Check{
					ProxyEntityName: "bar",
				},
				Entity: corev2.FixtureEntity("foo"),
			},
			storeFunc: func(store *mockstore.MockStore) {
				store.On("GetEntityByName", mock.Anything, "bar").
					Return(nilEntity, nil)
				store.On("UpdateEntity", mock.Anything, mock.AnythingOfType("*v2.Entity")).
					Return(errors.New("error"))
			},
			wantEntityName: "foo",
			wantErr:        true,
		},
		{
			name: "entity gets created as proxy entity with provided definition",
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
			storeFunc: func(store *mockstore.MockStore) {
				store.On("GetEntityByName", mock.Anything, "foo").
					Return(nilEntity, nil)
				store.On("UpdateEntity", mock.Anything, mock.AnythingOfType("*v2.Entity")).
					Return(nil)
			},
			wantEntityName: "foo",
			wantEntity: &corev2.Entity{
				ObjectMeta: corev2.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				EntityClass:   "proxy",
				Subscriptions: []string{"linux", "entity:foo"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockstore.MockStore{}
			if tt.storeFunc != nil {
				tt.storeFunc(store)
			}
			defer store.AssertExpectations(t)

			if err := createProxyEntity(tt.event, store); (err != nil) != tt.wantErr {
				t.Errorf("createProxyEntity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.event.Entity.Name, tt.wantEntityName) {
				t.Errorf("createProxyEntity() entity name = %v, want %v", tt.event.Entity.Name, tt.wantEntityName)
			}
			if tt.wantEntity != nil && !reflect.DeepEqual(tt.event.Entity, tt.wantEntity) {
				t.Errorf("createProxyEntity() entity = %v, want %v", tt.event.Entity, tt.wantEntity)
			}
		})
	}
}
