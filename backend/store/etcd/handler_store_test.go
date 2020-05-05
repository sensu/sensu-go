// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestHandlerStorage(t *testing.T) {
	testWithEtcd(t, func(s store.Store) {
		handler := corev2.FixtureHandler("handler1")
		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, handler.Namespace)

		// We should receive an empty slice if no results were found
		pred := &store.SelectionPredicate{}
		handlers, err := s.GetHandlers(ctx, pred)
		assert.NoError(t, err)
		assert.NotNil(t, handlers)
		assert.Empty(t, pred.Continue)

		err = s.UpdateHandler(ctx, handler)
		assert.NoError(t, err)

		retrieved, err := s.GetHandlerByName(ctx, "handler1")
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, handler.Name, retrieved.Name)
		assert.Equal(t, handler.Type, retrieved.Type)
		assert.Equal(t, handler.Command, retrieved.Command)
		assert.Equal(t, handler.Timeout, retrieved.Timeout)

		dne, err := s.GetHandlerByName(ctx, "doesnotexist")
		require.NoError(t, err)
		require.Nil(t, dne)
		require.NoError(t, s.DeleteHandlerByName(ctx, "doesnotexist"))

		handlers, err = s.GetHandlers(ctx, pred)
		require.NoError(t, err)
		require.NotEmpty(t, handlers)
		assert.Equal(t, 1, len(handlers))
		assert.Empty(t, pred.Continue)

		// Updating a handler in a nonexistent org and env should not work
		handler.Namespace = "missing"
		err = s.UpdateHandler(ctx, handler)
		assert.Error(t, err)
	})
}
