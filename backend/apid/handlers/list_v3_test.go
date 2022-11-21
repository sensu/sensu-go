package handlers

import (
	"context"
	"testing"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/sensu/sensu-go/testing/fixture"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

func TestHandlers_ListV3Resources(t *testing.T) {
	meta := corev2.NewObjectMeta("default", "bar")
	barResource := &fixture.V3Resource{Metadata: &meta}
	wrapper, _ := storev2.WrapResource(barResource)
	tests := []struct {
		name      string
		storeFunc func(*mockstore.ConfigStore)
		want      []corev3.Resource
		wantErr   bool
	}{
		{
			name: "store err",
			storeFunc: func(s *mockstore.ConfigStore) {
				s.On("List", mock.Anything, mock.Anything, mock.Anything).
					Return((storev2.WrapList)(nil), &store.ErrInternal{})
			},
			want:    []corev3.Resource(nil),
			wantErr: true,
		},
		{
			name: "sucessful list",
			storeFunc: func(s *mockstore.ConfigStore) {
				s.On("List", mock.Anything, mock.Anything, mock.Anything).
					Return(storev2.WrapList(wrap.List{wrapper.(*wrap.Wrapper)}), nil)
			},
			want: []corev3.Resource{corev3.Resource(barResource)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &mockstore.V2MockStore{}
			cs := new(mockstore.ConfigStore)
			s.On("GetConfigStore").Return(cs)
			if tt.storeFunc != nil {
				tt.storeFunc(cs)
			}

			h := NewHandlers[*fixture.V3Resource](s)

			_, err := h.ListResources(context.Background(), &store.SelectionPredicate{})
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.ListResources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
