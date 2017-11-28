package types

import (
	"errors"
	"time"
)

// HookRequestType is the message type string for hook request.
const HookRequestType = "hook_request"

// Validate returns an error if the hook does not pass validation tests.
func (c *Hook) Validate() error {
	if config := c.Config; config != nil {
		if err := config.Validate(); err != nil {
			return err
		}
	}

	if c.Status < 0 {
		return errors.New("hook status must be greater than or equal to 0")
	}

	return nil
}

// Validate returns an error if the hook does not pass validation tests.
func (c *HookConfig) Validate() error {
	if err := ValidateName(c.Name); err != nil {
		return errors.New("hook name " + err.Error())
	}

	if c.Command == "" {
		return errors.New("command cannot be empty")
	}

	if c.Timeout <= 0 {
		return errors.New("hook timeout must be greater than 0")
	}

	if c.Environment == "" {
		return errors.New("environment cannot be empty")
	}

	if c.Organization == "" {
		return errors.New("organization must be set")
	}

	return nil
}

// FixtureHookRequest returns a fixture for a HookRequest object.
func FixtureHookRequest(id string) *HookRequest {
	config := FixtureHookConfig(id)

	return &HookRequest{
		Config: config,
	}
}

// FixtureHookConfig returns a fixture for a HookConfig object.
func FixtureHookConfig(id string) *HookConfig {
	timeout := uint32(10)

	return &HookConfig{
		Name:         id,
		Command:      "true",
		Timeout:      timeout,
		Stdin:        false,
		Environment:  "default",
		Organization: "default",
	}
}

// FixtureHook returns a fixture for a Hook object.
func FixtureHook(id string) *Hook {
	t := time.Now().Unix()
	config := FixtureHookConfig(id)

	return &Hook{
		Status:   0,
		Output:   "",
		Issued:   t,
		Executed: t + 1,
		Duration: 1.0,
		Config:   config,
	}
}
