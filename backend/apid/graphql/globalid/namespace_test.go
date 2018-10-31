package globalid

import (
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamespaceTranslator(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	nsp := &types.Namespace{Name: "myns"}

	// Encode
	gid := NamespaceTranslator.EncodeToString(nsp)
	assert.Equal("srn:namespaces:myns", gid)

	// Decode
	idComponents, err := Parse(gid)
	require.NoError(err)
	assert.Empty(idComponents.Namespace())
	assert.Equal("myns", idComponents.UniqueComponent())
}
