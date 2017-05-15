package cli

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("api-url", "", "")
	flags.String("profile", "", "")
	c := New(flags)

	// Ensure that given SensuCli instance is complete
	assert.NotNil(c, "New should not return nil")
	assert.NotNil(c.Config, "New should include Config")
	assert.NotNil(c.Client, "New should include Client")
	assert.NotNil(c.Logger, "New should include Logger")

	// Ensure that flags are correctly set
	flags.Set("api-url", "http://localhost:8080")
	flags.Set("profile", "sensu")

	assert.Equal(c.Config.GetString("api-url"), "http://localhost:8080")
	assert.Equal(c.Config.GetString("profile"), "sensu")
}
