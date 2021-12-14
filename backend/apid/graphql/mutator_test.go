package graphql

import (
	"context"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMutatorTypeToJSONField(t *testing.T) {
	src := corev2.FixtureMutator("name")
	imp := &mutatorImpl{}

	res, err := imp.ToJSON(graphql.ResolveParams{Source: src, Context: context.Background()})
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestMutatorTypeTypeField(t *testing.T) {
	tests := []struct {
		name   string
		source *corev2.Mutator
		want   string
	}{
		{
			name:   "default",
			source: corev2.FixtureMutator("name"),
			want:   "pipe",
		},
		{
			name:   "jabbascript",
			source: &corev2.Mutator{Type: "javascript"},
			want:   "javascript",
		},
		{
			name:   "backward compat",
			source: &corev2.Mutator{Type: ""},
			want:   "pipe",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imp := &mutatorImpl{}
			res, err := imp.Type(graphql.ResolveParams{Source: tt.source, Context: context.Background()})
			require.NoError(t, err)
			assert.Equal(t, tt.want, res)
		})
	}
}
