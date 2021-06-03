package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

func TestStoreProxy(t *testing.T) {
	storeA := new(mockstore.MockStore)
	storeB := new(mockstore.MockStore)

	proxy := store.NewStoreProxy(storeA)
	ctx := context.Background()

	storeA.On("DeleteEventByEntityCheck", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("a"))
	storeB.On("DeleteEventByEntityCheck", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("b"))

	err := proxy.DeleteEventByEntityCheck(ctx, "", "")
	if got, want := err.Error(), "a"; got != want {
		t.Fatalf("bad response from event store: got %s, want %s", got, want)
	}

	proxy.UpdateStore(storeB)

	err = proxy.DeleteEventByEntityCheck(ctx, "", "")
	if got, want := err.Error(), "b"; got != want {
		t.Fatalf("bad response from event store: got %s, want %s", got, want)
	}
}
