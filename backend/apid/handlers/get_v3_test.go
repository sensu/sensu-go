package handlers

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/testing/fixture"
	"github.com/sensu/sensu-go/testing/mockstore"
)

func TestHandlers_GetV3Resource(t *testing.T) {
	meta := corev2.NewObjectMeta("default", "bar")
	barResource := &fixture.V3Resource{Metadata: &meta}
	wrapper, _ := storev2.WrapResource(barResource)
	tests := []struct {
		name      string
		urlVars   map[string]string
		storeFunc func(*mockstore.ConfigStore)
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
			storeFunc: func(s *mockstore.ConfigStore) {
				s.On("Get", mock.Anything, mock.Anything).
					Return((storev2.Wrapper)(nil), &store.ErrNotFound{})
			},
			wantErr: true,
		},
		{
			name:    "store ErrInternal",
			urlVars: map[string]string{"id": "foo"},
			storeFunc: func(s *mockstore.ConfigStore) {
				s.On("Get", mock.Anything, mock.Anything).
					Return((storev2.Wrapper)(nil), &store.ErrInternal{})
			},
			wantErr: true,
		},
		{
			name:    "successful get",
			urlVars: map[string]string{"id": "foo"},
			storeFunc: func(s *mockstore.ConfigStore) {
				s.On("Get", mock.Anything, mock.Anything).
					Return(wrapper, nil)
			},
			want: barResource,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sto := &mockstore.V2MockStore{}
			cs := new(mockstore.ConfigStore)
			sto.On("GetConfigStore").Return(cs)

			if tt.storeFunc != nil {
				tt.storeFunc(cs)
			}

			h := NewHandlers[*fixture.V3Resource](sto)

			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			r = mux.SetURLVars(r, tt.urlVars)

			got, err := h.GetResource(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.GetResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != nil {
				meta := got.GetMetadata()
				// delete these to facilitate comparison
				delete(meta.Labels, store.SensuCreatedAtKey)
				delete(meta.Labels, store.SensuUpdatedAtKey)
				delete(meta.Labels, store.SensuDeletedAtKey)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.GetResource() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
