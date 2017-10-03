package types

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

// CheckRequestType is the message type string for check request.
const CheckRequestType = "check_request"

// A CheckRequest represents a request to execute a check
type CheckRequest struct {
	// Config is the specification of a check.
	Config *CheckConfig `json:"config,omitempty"`

	// ExpandedAssets are a list of assets required to execute check.
	ExpandedAssets []Asset `json:"assets,omitempty"`
}

// A Check is a check specification and optionally the results of the check's
// execution.
type Check struct {
	// Config is the specification of a check.
	Config *CheckConfig `json:"config,omitempty"`

	// Duration of execution.
	Duration float64 `json:"duration,omitempty"`

	// Time check request was executed
	Executed int64 `json:"executed,omitempty"`

	// History is the check state history.
	History []CheckHistory `json:"history,omitempty"`

	// Time check request was issued.
	Issued int64 `json:"issued,omitempty"`

	// Output from the execution of Command.
	Output string `json:"output,omitempty"`

	// Status is the exit status code produced by the check.
	Status int `json:"status,omitempty"`
}

// CheckConfig is the specification of a check.
type CheckConfig struct {
	// Command is the command to be executed.
	Command string `json:"command"`

	// Environment indicates to which env a check belongs to
	Environment string `json:"environment"`

	// Handlers are the event handler for the check (incidents
	// and/or metrics).
	Handlers []string `json:"handlers"`

	// HighFlapThreshold is the flap detection high threshold (% state change) for
	// the check. Sensu uses the same flap detection algorithm as Nagios.
	HighFlapThreshold uint `json:"high_flap_threshold"`

	// Interval is the interval, in seconds, at which the check should be run.
	Interval uint `json:"interval"`

	// LowFlapThreshold is the flap detection low threshold (% state change) for
	// the check. Sensu uses the same flap detection algorithm as Nagios.
	LowFlapThreshold uint `json:"low_flap_threshold"`

	// Name is the unique identifier for a check.
	Name string `json:"name"`

	// Organization indicates to which org a check belongs to
	Organization string `json:"organization"`

	// Publish indicates if check requests are published for the check
	Publish bool `json:"publish"`

	// RuntimeAssets are a list of assets required to execute check.
	RuntimeAssets []string `json:"runtime_assets"`

	// Subscriptions is the list of subscribers for the check.
	Subscriptions []string `json:"subscriptions"`
}

// Validate returns an error if the check does not pass validation tests.
func (c *Check) Validate() error {
	if config := c.Config; config != nil {
		if err := config.Validate(); err != nil {
			return err
		}
	}

	if c.Status < 0 {
		return errors.New("check status must be greater than or equal to 0")
	}

	return nil
}

// Validate returns an error if the check does not pass validation tests.
func (c *CheckConfig) Validate() error {
	if err := ValidateName(c.Name); err != nil {
		return errors.New("check name " + err.Error())
	}

	if c.Interval == 0 {
		return errors.New("check interval must be greater than 0")
	}

	if c.Environment == "" {
		return errors.New("environment cannot be empty")
	}

	if c.Organization == "" {
		return errors.New("organization must be set")
	}

	for _, assetName := range c.RuntimeAssets {
		if err := ValidateAssetName(assetName); err != nil {
			return fmt.Errorf("asset's %s", err)
		}
	}

	return nil
}

// GetOrg refers to the organization the check belongs to
func (c *CheckConfig) GetOrg() string {
	return c.Organization
}

// GetEnv refers to the organization the check belongs to
func (c *CheckConfig) GetEnv() string {
	return c.Environment
}

// CheckHistory is a record of a check execution and its status.
type CheckHistory struct {
	Status   int   `json:"status"`
	Executed int64 `json:"executed"`
}

// ByExecuted implements the sort.Interface for []CheckHistory based on the
// Executed field.
//
// Example:
//
// sort.Sort(ByExecuted(check.History))
type ByExecuted []CheckHistory

func (b ByExecuted) Len() int           { return len(b) }
func (b ByExecuted) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByExecuted) Less(i, j int) bool { return b[i].Executed < b[j].Executed }

// MergeWith updates the current Check with the history of the check given as
// an argument, updating the current check's history appropriately.
func (c *Check) MergeWith(chk *Check) {
	history := chk.History
	histEntry := CheckHistory{
		Status:   chk.Status,
		Executed: chk.Executed,
	}

	history = append([]CheckHistory{histEntry}, history...)
	sort.Sort(ByExecuted(history))
	if len(history) > 21 {
		history = history[1:]
	}

	c.History = history
}

// FixtureCheckRequest returns a fixture for a CheckRequest object.
func FixtureCheckRequest(id string) *CheckRequest {
	config := FixtureCheckConfig(id)

	return &CheckRequest{
		Config: config,
		ExpandedAssets: []Asset{
			*FixtureAsset("ruby-2-4-2"),
		},
	}
}

// FixtureCheckConfig returns a fixture for a CheckConfig object.
func FixtureCheckConfig(id string) *CheckConfig {
	interval := uint(60)

	return &CheckConfig{
		Name:          id,
		Interval:      interval,
		Subscriptions: []string{},
		Command:       "command",
		Handlers:      []string{},
		RuntimeAssets: []string{"ruby-2-4-2"},
		Environment:   "default",
		Organization:  "default",
		Publish:       true,
	}
}

// FixtureCheck returns a fixture for a Check object.
func FixtureCheck(id string) *Check {
	t := time.Now().Unix()
	config := FixtureCheckConfig(id)
	history := make([]CheckHistory, 21)
	for i := 0; i < 21; i++ {
		history[i] = CheckHistory{
			Status:   0,
			Executed: t - (60 * int64(i+1)),
		}
	}

	return &Check{
		Status:   0,
		Output:   "",
		Issued:   t,
		Executed: t + 1,
		Duration: 1.0,
		History:  history,
		Config:   config,
	}
}
