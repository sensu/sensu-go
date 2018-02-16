package types

import (
	"encoding/json"
	"errors"
	"regexp"
	"time"
)

// CheckHookRegexStr used to validate type of check hook
var CheckHookRegexStr = `([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])`

// CheckHookRegex used to validate type of check hook
var CheckHookRegex = regexp.MustCompile("^" + CheckHookRegexStr + "$")

// Severities used to validate type of check hook
var Severities = []string{"ok", "warning", "critical", "unknown", "non-zero"}

// HookRequestType is the message type string for hook request.
const HookRequestType = "hook_request"

// Validate returns an error if the hook does not pass validation tests.
func (c *Hook) Validate() error {
	if err := c.HookConfig.Validate(); err != nil {
		return err
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

// Validate returns an error if the check hook does not pass validation tests.
func (h *HookList) Validate() error {
	if h.Type == "" {
		return errors.New("type cannot be empty")
	}

	if !(CheckHookRegex.MatchString(h.Type) || isSeverity(h.Type)) {
		return errors.New(
			"valid check hook types are \"1\"-\"255\", \"ok\", \"warning\", \"critical\", \"unknown\", and \"non-zero\"",
		)
	}

	return nil
}

func isSeverity(name string) bool {
	for _, sev := range Severities {
		if sev == name {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface.
func (h *HookList) MarshalJSON() ([]byte, error) {
	result := make(map[string][]string)
	result[h.Type] = append(result[h.Type], h.Hooks...)
	return json.Marshal(result)
}

// UnmarshalJSON implements the json.Marshaler interface.
func (h *HookList) UnmarshalJSON(b []byte) error {
	result := map[string][]string{}
	if err := json.Unmarshal(b, &result); err != nil {
		return err
	}
	for key := range result {
		h.Type = key
		h.Hooks = result[key]
	}

	return nil
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
		Status:     0,
		Output:     "",
		Issued:     t,
		Executed:   t + 1,
		Duration:   1.0,
		HookConfig: *config,
	}
}

// FixtureHookList returns a fixture for a HookList object.
func FixtureHookList(hookName string) *HookList {
	return &HookList{
		Hooks: []string{hookName},
		Type:  "non-zero",
	}
}
