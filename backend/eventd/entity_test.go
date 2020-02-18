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
		name       string
		event      *corev2.Event
		storeFunc  storeFunc
		wantEntity string
		wantErr    bool
	}{
		{
			name:  "entity exists",
			event: corev2.FixtureEvent("foo", "check-cpu"),
			storeFunc: func(store *mockstore.MockStore) {
				store.On("GetEntityByName", mock.Anything, "foo").
					Return(corev2.FixtureEntity("foo"), nil)
			},
			wantEntity: "foo",
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
			wantEntity: "foo",
		},
		{
			name:  "store error while getting an entity",
			event: corev2.FixtureEvent("foo", "check-cpu"),
			storeFunc: func(store *mockstore.MockStore) {
				store.On("GetEntityByName", mock.Anything, "foo").
					Return(&corev2.Entity{}, errors.New("error"))
			},
			wantEntity: "foo",
			wantErr:    true,
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
			wantEntity: "bar",
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
			wantEntity: "bar",
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
			wantEntity: "foo",
			wantErr:    true,
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
			if !reflect.DeepEqual(tt.event.Entity.Name, tt.wantEntity) {
				t.Errorf("createProxyEntity() entity name = %v, want %v", tt.event.Entity.Name, tt.wantEntity)
			}
		})
	}
}
