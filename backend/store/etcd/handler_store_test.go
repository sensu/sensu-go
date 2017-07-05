package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestHandlerStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		handler := types.FixtureHandler("handler1")
		ctx := context.WithValue(context.Background(), types.OrganizationKey, handler.Organization)

		// We should receive an empty slice if no results were found
		handlers, err := store.GetHandlers(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, handlers)

		err = store.UpdateHandler(handler)
		assert.NoError(t, err)

		retrieved, err := store.GetHandlerByName(ctx, "handler1")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)

		assert.Equal(t, handler.Name, retrieved.Name)
		assert.Equal(t, handler.Type, retrieved.Type)
		assert.Equal(t, handler.Command, retrieved.Command)
		assert.Equal(t, handler.Timeout, retrieved.Timeout)

		handlers, err = store.GetHandlers(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, handlers)
		assert.Equal(t, 1, len(handlers))
	})
}
