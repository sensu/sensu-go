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
		storeFunc func(*mockstore.V2MockStore)
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
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("Get", mock.Anything).
					Return((storev2.Wrapper)(nil), &store.ErrNotFound{})
			},
			wantErr: true,
		},
		{
			name:    "store ErrInternal",
			urlVars: map[string]string{"id": "foo"},
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("Get", mock.Anything).
					Return((storev2.Wrapper)(nil), &store.ErrInternal{})
			},
			wantErr: true,
		},
		{
			name:    "successful get",
			urlVars: map[string]string{"id": "foo"},
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("Get", mock.Anything).
					Return(wrapper, nil)
			},
			want: barResource,
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

			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			r = mux.SetURLVars(r, tt.urlVars)

			got, err := h.GetV3Resource(r)
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
