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
	ctx = context.WithValue(ctx, types.NamespaceKey, "acme")
	builder := NewKeyBuilder("checks").WithContext(ctx)
	assert.Equal(t, "/sensu.io/checks/acme/check_name", builder.Build("check_name"))

	// Querying the content of a whole namespace uses etcd prefixes, so we want
	// the key to end in '/' in order not to inadvertently retrieve objects from
	// a different namespace that would have the same prefix, eg: acme-devel
	assert.Equal(t, "/sensu.io/checks/acme/", builder.Build(""))
}
