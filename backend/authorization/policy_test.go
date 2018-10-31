package authorization

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthContext(t *testing.T) {
	var authCtx Context
	ctx := context.TODO()

	t.Run("empty context", func(t *testing.T) {
		authCtx = ExtractValueFromContext(ctx)
		assert.NotNil(t, authCtx)
		assert.Empty(t, authCtx.Namespace)
		assert.Empty(t, authCtx.Actor.Rules)
	})

	t.Run("given context w/ namespace", func(t *testing.T) {
		namespaceCtx := context.WithValue(ctx, types.NamespaceKey, "default")
		authCtx = ExtractValueFromContext(namespaceCtx)
		assert.NotNil(t, authCtx)
		assert.Equal(t, authCtx.Namespace, "default")
		assert.Empty(t, authCtx.Actor.Rules)
	})

	t.Run("given context w/ namespace & roles", func(t *testing.T) {
		tCtx := context.WithValue(ctx, types.NamespaceKey, "default")
		tCtx = context.WithValue(
			tCtx,
			types.AuthorizationRoleKey,
			[]*types.Role{types.FixtureRole("x", "y")},
		)
		authCtx = ExtractValueFromContext(tCtx)
		assert.NotNil(t, authCtx)
		assert.Equal(t, "default", authCtx.Namespace)
	})
}
