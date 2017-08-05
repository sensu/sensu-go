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
		assert.Empty(t, authCtx.Organization)
		assert.Empty(t, authCtx.Actor.Rules)
	})

	t.Run("given context w/ organization", func(t *testing.T) {
		orgCtx := context.WithValue(ctx, types.OrganizationKey, "default")
		authCtx = ExtractValueFromContext(orgCtx)
		assert.NotNil(t, authCtx)
		assert.Equal(t, authCtx.Organization, "default")
		assert.Empty(t, authCtx.Actor.Rules)
	})

	t.Run("given context w/ org & roles", func(t *testing.T) {
		tCtx := context.WithValue(ctx, types.OrganizationKey, "default")
		tCtx = context.WithValue(tCtx, types.EnvironmentKey, "dev")
		tCtx = context.WithValue(
			tCtx,
			types.AuthorizationRoleKey,
			[]*types.Role{types.FixtureRole("x", "y", "z")},
		)
		authCtx = ExtractValueFromContext(tCtx)
		assert.NotNil(t, authCtx)
		assert.Equal(t, "default", authCtx.Organization)
		assert.NotEmpty(t, authCtx.Actor)
	})
}
