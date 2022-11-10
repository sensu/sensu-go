package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/fixture"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandlers_UpdateResourceV3(t *testing.T) {
	type storeFunc func(*mockstore.V2MockStore)
	tests := []struct {
		name      string
		body      []byte
		urlVars   map[string]string
		storeFunc storeFunc
		wantErr   bool
	}{
		{
			name:    "invalid request body",
			body:    []byte("foobar"),
			wantErr: true,
		},
		{
			name:    "invalid resource meta",
			body:    marshal(t, fixture.V3Resource{Metadata: corev2.NewObjectMetaP("foo", "acme")}),
			urlVars: map[string]string{"id": "bar", "namespace": "acme"},
			wantErr: true,
		},
		{
			name: "store err, not valid",
			body: marshal(t, fixture.V3Resource{Metadata: corev2.NewObjectMetaP("", "")}),
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("CreateOrUpdate", mock.Anything, mock.Anything).
					Return(&store.ErrNotValid{Err: errors.New("error")})
			},
			wantErr: true,
		},
		{
			name: "store err, default",
			body: marshal(t, fixture.V3Resource{Metadata: corev2.NewObjectMetaP("", "")}),
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("CreateOrUpdate", mock.Anything, mock.Anything).
					Return(&store.ErrInternal{})
			},
			wantErr: true,
		},
		{
			name: "successful create",
			body: marshal(t, fixture.V3Resource{Metadata: corev2.NewObjectMetaP("", "")}),
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("CreateOrUpdate", mock.Anything, mock.Anything).
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

			r, _ := http.NewRequest(http.MethodPut, "/", bytes.NewReader(tt.body))
			r = mux.SetURLVars(r, tt.urlVars)

			_, err := h.CreateOrUpdateV3Resource(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.CreateOrUpdateV3Resource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestCreatedByUpdateV3(t *testing.T) {
	claims, err := jwt.NewClaims(&corev2.User{Username: "admin"})
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), corev2.ClaimsKey, claims)
	body := marshal(t, fixture.V3Resource{Metadata: corev2.NewObjectMetaP("", "")})

	store := &mockstore.V2MockStore{}
	h := Handlers{
		V3Resource: &fixture.V3Resource{},
		StoreV2:    store,
	}

	store.On("CreateOrUpdate", mock.Anything, mock.Anything).Return(nil)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, "/", bytes.NewReader(body))
	assert.NoError(t, err)

	_, err = h.CreateOrUpdateV3Resource(req)
	assert.NoError(t, err)
}
