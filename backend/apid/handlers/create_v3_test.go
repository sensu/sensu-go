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

func TestHandlers_CreateResource(t *testing.T) {
	tests := []struct {
		name      string
		body      []byte
		urlVars   map[string]string
		storeFunc func(s *mockstore.V2MockStore)
		wantErr   bool
	}{
		{
			name:    "invalid request body",
			body:    []byte("foobar"),
			wantErr: true,
		},
		{
			name: "invalid resource meta",
			body: marshal(t, fixture.V3Resource{Metadata: &corev2.ObjectMeta{
				Name:      "foo",
				Namespace: "acme",
			}}),
			urlVars: map[string]string{"id": "bar", "namespace": "acme"},
			wantErr: true,
		},
		{
			name: "store err, already exists",
			body: marshal(t, fixture.V3Resource{Metadata: &corev2.ObjectMeta{}}),
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("CreateIfNotExists", mock.Anything, mock.Anything, mock.Anything).
					Return(&store.ErrAlreadyExists{})
			},
			wantErr: true,
		},
		{
			name: "store err, not valid",
			body: marshal(t, fixture.V3Resource{Metadata: &corev2.ObjectMeta{}}),
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("CreateIfNotExists", mock.Anything, mock.Anything, mock.Anything).
					Return(&store.ErrNotValid{Err: errors.New("error")})
			},
			wantErr: true,
		},
		{
			name: "store err, default",
			body: marshal(t, fixture.V3Resource{Metadata: &corev2.ObjectMeta{}}),
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("CreateIfNotExists", mock.Anything, mock.Anything, mock.Anything).
					Return(&store.ErrInternal{})
			},
			wantErr: true,
		},
		{
			name: "successful create",
			body: marshal(t, fixture.V3Resource{Metadata: &corev2.ObjectMeta{}}),
			storeFunc: func(s *mockstore.V2MockStore) {
				s.On("CreateIfNotExists", mock.Anything, mock.Anything, mock.Anything).
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
				Resource: &fixture.V3Resource{},
				Store:    store,
			}

			r, _ := http.NewRequest(http.MethodPost, "/", bytes.NewReader(tt.body))
			r = mux.SetURLVars(r, tt.urlVars)

			_, err := h.CreateResource(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.CreateResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestV3CreatedByCreate(t *testing.T) {
	claims, err := jwt.NewClaims(&corev2.User{Username: "admin"})
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), corev2.ClaimsKey, claims)
	body := marshal(t, fixture.V3Resource{Metadata: &corev2.ObjectMeta{}})

	store := &mockstore.V2MockStore{}
	h := Handlers{
		Resource: &fixture.V3Resource{},
		Store:    store,
	}

	store.On("CreateIfNotExists", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/", bytes.NewReader(body))
	assert.NoError(t, err)

	_, err = h.CreateResource(req)
	assert.NoError(t, err)
}
