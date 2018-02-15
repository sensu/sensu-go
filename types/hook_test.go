package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHookValidate(t *testing.T) {
	var h Hook

	// Invalid status
	h.Status = -1
	assert.Error(t, h.Validate())

	// Invalid without config
	h.Status = 0
	assert.Error(t, h.Validate())

	// Valid with valid config
	h.HookConfig = HookConfig{
		Name:         "test",
		Command:      "yes",
		Timeout:      10,
		Environment:  "default",
		Organization: "default",
	}
	assert.NoError(t, h.Validate())
}

func TestHookConfig(t *testing.T) {
	var h HookConfig

	// Invalid name
	assert.Error(t, h.Validate())
	h.Name = "foo"

	// Invalid timeout
	assert.Error(t, h.Validate())
	h.Timeout = 60

	// Invalid command
	assert.Error(t, h.Validate())
	h.Command = "echo 'foo'"

	// Invalid organization
	assert.Error(t, h.Validate())
	h.Organization = "default"

	// Invalid environment
	assert.Error(t, h.Validate())
	h.Environment = "default"

	// Valid hook
	assert.NoError(t, h.Validate())
}

func TestFixtureHookIsValid(t *testing.T) {
	c := FixtureHook("hook")
	config := c.HookConfig

	assert.Equal(t, "hook", config.Name)
	assert.NoError(t, config.Validate())
}
