package graphql

import (
	"testing"

	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestEnvColourID(t *testing.T) {
	handler := &envImpl{}
	env := types.Environment{Name: "pink"}

	colour, err := handler.ColourID(graphql.ResolveParams{Source: &env})
	assert.NoError(t, err)
	assert.Equal(t, string(colour), "BLUE")
}
