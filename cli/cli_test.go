package cli

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("api-url", "", "")
	require.NoError(t, flags.Set("api-url", "http://localhost:8080"))

	c := New(flags)

	// Ensure that given SensuCli instance is complete
	assert.NotNil(c, "New should not return nil")
	assert.NotNil(c.Config, "New should include Config")
	assert.NotNil(c.Client, "New should include Client")
	assert.NotNil(c.Logger, "New should include Logger")

	assert.Equal(c.Config.APIUrl(), "http://localhost:8080")
}
