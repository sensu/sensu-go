package filter

import (
	"testing"

	v2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {
	filters := map[string]Filter{
		"name": String(func(res v2.Resource, name string) bool {
			return res.GetObjectMeta().Name == name
		}),
	}

	m, err := Compile([]string{}, filters, nil)
	require.NoError(t, err)
	assert.NotNil(t, m)

	// statement doesn't match a filter
	m, err = Compile([]string{"unknown:test"}, filters, nil)
	require.Error(t, err, ErrKeylessStatement)
	assert.Nil(t, m)

	// statement is not valid
	m, err = Compile([]string{"badstatement"}, filters, nil)
	require.Error(t, err)
	assert.Nil(t, m)

	// valid statement
	m, err = Compile([]string{"name:disk-check"}, filters, nil)
	require.NoError(t, err)
	assert.NotNil(t, m)
}
