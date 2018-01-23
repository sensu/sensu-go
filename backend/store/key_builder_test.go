package store

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestSimpleKeyBuilder(t *testing.T) {
	t.Parallel()

	builder := NewKeyBuilder("checks")
	assert.Equal(t, "/sensu.io/checks", builder.Build(""))
}

func TestContextKeyBuilder(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = context.WithValue(ctx, types.OrganizationKey, "org")
	ctx = context.WithValue(ctx, types.EnvironmentKey, "env")
	builder := NewKeyBuilder("checks").WithContext(ctx)
	assert.Equal(t, "/sensu.io/checks/org/env/check_name", builder.Build("check_name"))
}
