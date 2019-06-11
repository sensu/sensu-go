package handlers

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/fixture"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

func TestHandlers_DeleteResource(t *testing.T) {
	type storeFunc func(*mockstore.MockStore)
	tests := []struct {
		name      string
		urlVars   map[string]string
		storeFunc storeFunc
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
				s.On("DeleteResource", mock.Anything, "resource", "foo").
					Return(&store.ErrNotFound{})
			},
			wantErr: true,
		},
		{
			name:    "store ErrInternal",
			urlVars: map[string]string{"id": "foo"},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("DeleteResource", mock.Anything, "resource", "foo").
					Return(&store.ErrInternal{})
			},
			wantErr: true,
		},
		{
			name:    "successful delete",
			urlVars: map[string]string{"id": "foo"},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("DeleteResource", mock.Anything, "resource", "foo").
					Return(nil)
			},
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

			_, err := h.DeleteResource(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.DeleteResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
