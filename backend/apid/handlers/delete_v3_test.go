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

func TestHandlers_DeleteV3Resource(t *testing.T) {
	type storeFunc func(*mockstore.V2MockStore)
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
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("Delete", mock.Anything, mock.Anything).
					Return(&store.ErrNotFound{})
			},
			wantErr: true,
		},
		{
			name:    "store ErrInternal",
			urlVars: map[string]string{"id": "foo"},
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("Delete", mock.Anything, mock.Anything).
					Return(&store.ErrInternal{})
			},
			wantErr: true,
		},
		{
			name:    "successful delete",
			urlVars: map[string]string{"id": "foo"},
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("Delete", mock.Anything, mock.Anything).
					Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockstore.V2MockStore{}
			if tt.storeFunc != nil {
				tt.storeFunc(store)
			}

			h := Handlers{
				V3Resource: &fixture.V3Resource{},
				StoreV2:    store,
			}

			r, _ := http.NewRequest(http.MethodDelete, "/", nil)
			r = mux.SetURLVars(r, tt.urlVars)

			_, err := h.DeleteV3Resource(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.DeleteResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
