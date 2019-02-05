package graphql

import (
	"testing"

	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObjectMetaTypeLabelsField(t *testing.T) {
	meta := v2.ObjectMeta{
		Labels: map[string]string{
			"bob":  "builder",
			"fort": "knox",
		},
	}

	impl := objectMetaImpl{}
	params := graphql.ResolveParams{Source: meta}

	res, err := impl.Labels(params)
	require.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Contains(t, res, map[string]string{"key": "bob", "val": "builder"})
	assert.Contains(t, res, map[string]string{"key": "fort", "val": "knox"})
}

func TestObjectMetaTypeAnnotationsField(t *testing.T) {
	meta := v2.ObjectMeta{
		Annotations: map[string]string{
			"jeff": "gertsman",
			"brad": "shoemaker",
		},
	}

	impl := objectMetaImpl{}
	params := graphql.ResolveParams{Source: meta}

	res, err := impl.Annotations(params)
	require.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Contains(t, res, map[string]string{"key": "jeff", "val": "gertsman"})
	assert.Contains(t, res, map[string]string{"key": "brad", "val": "shoemaker"})
}
