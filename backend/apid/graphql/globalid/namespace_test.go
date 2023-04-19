package globalid

import (
	"context"
	"testing"

	v2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamespaceTranslator(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	ctx := context.Background()
	nsp := &v2.Namespace{Name: "myns"}

	// Encode
	gid := NamespaceTranslator.EncodeToString(ctx, nsp)
	assert.Equal("srn:namespaces:myns", gid)

	// Decode
	idComponents, err := Parse(gid)
	require.NoError(err)
	assert.Empty(idComponents.Namespace())
	assert.Equal("myns", idComponents.UniqueComponent())
}
