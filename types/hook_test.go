package types

import (
	"encoding/json"
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
		ObjectMeta: ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Command: "yes",
		Timeout: 10,
	}
	assert.NoError(t, h.Validate())
}

func TestHookListValidate(t *testing.T) {
	var h HookList

	// Invalid hooks
	h.Hooks = nil
	assert.Error(t, h.Validate())

	// Invalid hooks
	h.Hooks = []string{}
	assert.Error(t, h.Validate())

	// Invalid without type
	h.Hooks = append(h.Hooks, "hook")
	assert.Error(t, h.Validate())

	// Invalid type
	h.Type = "invalid"
	assert.Error(t, h.Validate())

	// Valid
	h.Type = "0"
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

	// Invalid Namespace
	assert.Error(t, h.Validate())
	h.Namespace = "default"

	// Valid hook
	assert.NoError(t, h.Validate())
}

func TestFixtureHookIsValid(t *testing.T) {
	c := FixtureHook("hook")
	config := c.HookConfig

	assert.Equal(t, "hook", config.Name)
	assert.NoError(t, config.Validate())
}

func TestHookUnmarshal_GH1520(t *testing.T) {
	b := []byte(`{"metadata": {"name": "foo", "namespace":"default"},"command":"ps aux","timeout":60}`)
	var h Hook
	var err error
	if err := json.Unmarshal(b, &h); err != nil {
		t.Fatal(err)
	}
	if err := h.Validate(); err != nil {
		t.Fatal(err)
	}
	b, err = json.Marshal(&h)
	if err != nil {
		t.Fatal(err)
	}
	var hc HookConfig
	if err := json.Unmarshal(b, &hc); err != nil {
		t.Fatal(err)
	}
	if err := hc.Validate(); err != nil {
		t.Fatal(err)
	}
}
