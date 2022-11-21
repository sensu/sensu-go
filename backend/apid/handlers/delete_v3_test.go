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
	type storeFunc func(*mockstore.ConfigStore)
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
			storeFunc: func(s *mockstore.ConfigStore) {
				s.On("Delete", mock.Anything, mock.Anything).
					Return(&store.ErrNotFound{})
			},
			wantErr: true,
		},
		{
			name:    "store ErrInternal",
			urlVars: map[string]string{"id": "foo"},
			storeFunc: func(s *mockstore.ConfigStore) {
				s.On("Delete", mock.Anything, mock.Anything).
					Return(&store.ErrInternal{})
			},
			wantErr: true,
		},
		{
			name:    "successful delete",
			urlVars: map[string]string{"id": "foo"},
			storeFunc: func(s *mockstore.ConfigStore) {
				s.On("Delete", mock.Anything, mock.Anything).
					Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockstore.V2MockStore{}
			cs := new(mockstore.ConfigStore)
			store.On("GetConfigStore").Return(cs)
			if tt.storeFunc != nil {
				tt.storeFunc(cs)
			}

			h := NewHandlers[*fixture.V3Resource](store)

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
