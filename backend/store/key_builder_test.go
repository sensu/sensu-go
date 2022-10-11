package store

import (
	"context"
	"testing"

	corev2 "github.com/sensu/core/v2"
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
	ctx = context.WithValue(ctx, corev2.NamespaceKey, "acme")
	builder := NewKeyBuilder("checks").WithContext(ctx)
	assert.Equal(t, "/sensu.io/checks/acme/check_name", builder.Build("check_name"))

	// Querying the content of a whole namespace uses etcd prefixes, so we want
	// the key to end in '/' in order not to inadvertently retrieve objects from
	// a different namespace that would have the same prefix, eg: acme-devel
	assert.Equal(t, "/sensu.io/checks/acme/", builder.Build(""))
}

func TestExactMatchKeyBuilder(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = context.WithValue(ctx, corev2.NamespaceKey, "default")
	builder := NewKeyBuilder("events").WithContext(ctx).WithExactMatch()

	// Querying all events for an entity uses etcd prefixes, so we want
	// the key to end in '/' in order not to inadvertently retrieve events from
	// a different entity that would have the same prefix, eg: foobar
	assert.Equal(t, "/sensu.io/events/default/entity_name/", builder.Build("entity_name"))
}
