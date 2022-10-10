package graphql

import (
	"testing"

	"github.com/sensu/core/v2"
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
	assert.Contains(t, res, KVPairString{Key: "bob", Val: "builder"})
	assert.Contains(t, res, KVPairString{Key: "fort", Val: "knox"})
}

func TestObjectMetaTypeAnnotationsField(t *testing.T) {
	meta := &v2.ObjectMeta{
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
	assert.Contains(t, res, KVPairString{Key: "jeff", Val: "gertsman"})
	assert.Contains(t, res, KVPairString{Key: "brad", Val: "shoemaker"})
}
