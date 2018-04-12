package globalid

import (
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironmentTranslator(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	env := &types.Environment{Name: "myenv", Organization: "myorg"}

	// Encode
	gid := EnvironmentTranslator.EncodeToString(env)
	assert.Equal("srn:environments:myorg:myenv", gid)

	// Decode
	idComponents, err := Parse(gid)
	require.NoError(err)
	assert.Empty(idComponents.Environment())
	assert.Equal("myorg", idComponents.Organization())
	assert.Equal("myenv", idComponents.UniqueComponent())
}
