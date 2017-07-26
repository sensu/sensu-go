package authorization

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestNewActorFromContext(t *testing.T) {
	var actor Actor
	ctx := context.TODO()

	t.Run("empty context", func(t *testing.T) {
		actor = NewActorFromContext(ctx)
		assert.NotNil(t, actor)
		assert.Empty(t, actor.Organization)
		assert.Empty(t, actor.Roles)
	})

	t.Run("given context w/ organization", func(t *testing.T) {
		orgCtx := context.WithValue(ctx, types.OrganizationKey, "default")
		actor = NewActorFromContext(orgCtx)
		assert.NotNil(t, actor)
		assert.Equal(t, actor.Organization, "default")
		assert.Empty(t, actor.Roles)
	})

	t.Run("given context w/ org & roles", func(t *testing.T) {
		tCtx := context.WithValue(ctx, types.OrganizationKey, "default")
		tCtx = context.WithValue(
			tCtx,
			types.AuthorizationRoleKey,
			[]*types.Role{types.FixtureRole("x", "y")},
		)
		actor = NewActorFromContext(tCtx)
		assert.NotNil(t, actor)
		assert.Equal(t, actor.Organization, "default")
		assert.NotEmpty(t, actor.Roles)
	})
}

func TestAbilityWithContext(t *testing.T) {
	ctx := context.TODO()
	ctx = context.WithValue(ctx, types.OrganizationKey, "default")
	ctx = context.WithValue(
		ctx,
		types.AuthorizationRoleKey,
		[]*types.Role{types.FixtureRole("x", "y")},
	)

	ability := Ability{}
	result := ability.WithContext(ctx)

	assert.Empty(t, ability.Actor.Organization)
	assert.Empty(t, ability.Actor.Roles)
	assert.NotEmpty(t, result.Actor)
	assert.NotEmpty(t, result.Actor.Roles)
}

func TestAbilityCanMethods(t *testing.T) {
	actor := Actor{
		Organization: "default",
		Roles: []*types.Role{{
			Name:  "test",
			Rules: []types.Rule{*types.FixtureRule("default")},
		}},
	}

	ability := Ability{}
	ability.Resource = "lemonade-stand"
	ability.Actor = actor

	rule := &actor.Roles[0].Rules[0]
	t.Run("no permissions", func(t *testing.T) {
		rule.Permissions = []string{}
		assert.False(t, ability.CanRead(), "cannot read")
		assert.False(t, ability.CanCreate(), "cannot create")
		assert.False(t, ability.CanUpdate(), "cannot update")
		assert.False(t, ability.CanDelete(), "cannot delete")
	})

	t.Run("with read & update permission", func(t *testing.T) {
		rule.Permissions = []string{"read", "update"}
		assert.True(t, ability.CanRead(), "can read")
		assert.False(t, ability.CanCreate(), "cannot create")
		assert.True(t, ability.CanUpdate(), "can update")
		assert.False(t, ability.CanDelete(), "cannot delete")
	})

	t.Run("with delete permission", func(t *testing.T) {
		rule.Permissions = []string{"delete"}
		assert.False(t, ability.CanRead(), "cannot read")
		assert.False(t, ability.CanCreate(), "cannot create")
		assert.False(t, ability.CanUpdate(), "cannot update")
		assert.True(t, ability.CanDelete(), "can delete")
	})

	t.Run("without access to resource", func(t *testing.T) {
		rule.Permissions = types.RuleAllPerms
		rule.Type = "blahlbahl"
		assert.False(t, ability.CanRead(), "cannot read")
		assert.False(t, ability.CanCreate(), "cannot create")
		assert.False(t, ability.CanUpdate(), "cannot update")
		assert.False(t, ability.CanDelete(), "cannot delete")
	})
}
