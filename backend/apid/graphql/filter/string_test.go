package filter

import (
	"testing"

	v2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestString(t *testing.T) {
	fn := func(res corev3.Resource, name string) bool {
		return res.GetMetadata().Name == name
	}

	f := String(fn)
	require.NotNil(t, f)

	matches, err := f("disk-check", nil)
	require.NoError(t, err)
	require.NotNil(t, matches)

	ok := matches(v2.FixtureCheck("disk-check"))
	assert.True(t, ok)
}
