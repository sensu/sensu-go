package helpers

import (
	"bytes"
	"testing"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/core/v3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestPrintYAML(t *testing.T) {
	assert := assert.New(t)

	resources := []corev3.Resource{
		corev2.FixtureCheckConfig("foo"),
		corev2.FixtureCheckConfig("bar"),
		corev3.FixtureEntityConfig("host"),
		corev3.FixtureNamespace("default"),
	}

	var expected, actual bytes.Buffer
	enc := yaml.NewEncoder(&expected)
	for _, resource := range resources {
		require.NoError(t, enc.Encode(types.WrapResource(resource)))
	}

	assert.NoError(PrintYAML(resources, &actual))
	assert.Equal(expected.String(), actual.String())
}
