package api

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

func TestJWTGetSecretNotFoundCreateIfNotExists(t *testing.T) {
	s := new(mockstore.V2MockStore)
	cs := new(mockstore.ConfigStore)
	s.On("GetConfigStore").Return(cs)
	cs.On("Get", mock.Anything, mock.Anything).Return(nil, &store.ErrNotFound{})
	cs.On("CreateIfNotExists", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	jwtclient := JWT{Store: s}
	value, err := jwtclient.GetSecret(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(value), 32; got != want {
		t.Errorf("value len wrong: got %d, want %d", got, want)
	}
	cs.AssertCalled(t, "Get", mock.Anything, mock.Anything)
	cs.AssertCalled(t, "CreateIfNotExists", mock.Anything, mock.Anything, mock.Anything)
}
