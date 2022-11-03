package handlers

import (
	"context"
	"reflect"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/fixture"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

func TestHandlers_ListResources(t *testing.T) {
	type storeFunc func(*mockstore.MockStore)
	barResource := &fixture.Resource{Foo: "bar"}
	tests := []struct {
		name      string
		storeFunc storeFunc
		want      interface{}
		wantErr   bool
	}{
		{
			name: "store err",
			storeFunc: func(s *mockstore.MockStore) {
				s.On("ListResources", mock.Anything, "resource", mock.AnythingOfType("*[]*fixture.Resource"), mock.AnythingOfType("*store.SelectionPredicate")).
					Return(&store.ErrInternal{})
			},
			want:    []corev2.Resource(nil),
			wantErr: true,
		},
		{
			name: "sucessful list",
			storeFunc: func(s *mockstore.MockStore) {
				s.On("ListResources", mock.Anything, "resource", mock.AnythingOfType("*[]*fixture.Resource"), mock.AnythingOfType("*store.SelectionPredicate")).
					Return(nil).
					Run(func(args mock.Arguments) {
						resources := args[2].(*[]*fixture.Resource)
						*resources = append(*resources, barResource)
					})
			},
			want: []corev2.Resource{corev2.Resource(barResource)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &mockstore.MockStore{}
			if tt.storeFunc != nil {
				tt.storeFunc(s)
			}

			h := Handlers{
				Resource: &fixture.Resource{},
				Store:    s,
			}

			got, err := h.ListResources(context.Background(), &store.SelectionPredicate{})
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.ListResources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.ListResources() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
