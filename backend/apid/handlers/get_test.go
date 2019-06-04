package handlers

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/fixture"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

func TestHandlers_GetResource(t *testing.T) {
	type storeFunc func(*mockstore.MockStore)
	barResource := &fixture.Resource{Foo: "bar"}
	tests := []struct {
		name      string
		urlVars   map[string]string
		storeFunc storeFunc
		want      interface{}
		wantErr   bool
	}{
		{
			name:    "invalid URL parameter",
			urlVars: map[string]string{"id": "%"},
			wantErr: true,
		},
		{
			name:    "store ErrNotFound",
			urlVars: map[string]string{"id": "foo"},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("GetResource", mock.Anything, "foo", mock.AnythingOfType("*fixture.Resource")).
					Return(&store.ErrNotFound{})
			},
			wantErr: true,
		},
		{
			name:    "store ErrInternal",
			urlVars: map[string]string{"id": "foo"},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("GetResource", mock.Anything, "foo", mock.AnythingOfType("*fixture.Resource")).
					Return(&store.ErrInternal{})
			},
			wantErr: true,
		},
		{
			name:    "successful get",
			urlVars: map[string]string{"id": "foo"},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("GetResource", mock.Anything, "foo", mock.AnythingOfType("*fixture.Resource")).
					Return(nil).
					Run(func(args mock.Arguments) {
						resource := args[2].(*fixture.Resource)
						*resource = *barResource
					})
			},
			want: barResource,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockstore.MockStore{}
			if tt.storeFunc != nil {
				tt.storeFunc(store)
			}

			h := Handlers{
				Resource: &fixture.Resource{},
				Store:    store,
			}

			r, _ := http.NewRequest(http.MethodDelete, "/", nil)
			r = mux.SetURLVars(r, tt.urlVars)

			got, err := h.GetResource(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetResource() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
