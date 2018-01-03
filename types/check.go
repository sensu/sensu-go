package types

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/robfig/cron"
	"github.com/sensu/sensu-go/types/dynamic"
)

// CheckRequestType is the message type string for check request.
const CheckRequestType = "check_request"

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

// UnmarshalJSON implements the json.Unmarshaler interface.
func (c *CheckConfig) UnmarshalJSON(b []byte) error {
	return dynamic.Unmarshal(b, c)
}

// MarshalJSON implements the json.Marshaler interface.
func (c *CheckConfig) MarshalJSON() ([]byte, error) {
	return dynamic.Marshal(c)
}

// SetExtendedAttributes sets the serialized ExtendedAttributes of c.
func (c *CheckConfig) SetExtendedAttributes(e []byte) {
	c.ExtendedAttributes = e
}

// Get implements govaluate.Parameters
func (c *CheckConfig) Get(name string) (interface{}, error) {
	return dynamic.GetField(c, name)
}

// Validate returns an error if the check does not pass validation tests.
func (c *CheckConfig) Validate() error {
	if err := ValidateName(c.Name); err != nil {
		return errors.New("check name " + err.Error())
	}

	if c.Cron != "" {
		if _, err := cron.Parse(c.Cron); err != nil {
			return errors.New("check cron string is invalid")
		}
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

	// The entity can be empty but can't contain invalid characters (only
	// alphanumeric string)
	if c.ProxyEntityID != "" {
		if err := ValidateName(c.ProxyEntityID); err != nil {
			return errors.New("proxy entity id " + err.Error())
		}
	}

	return c.Subdue.Validate()
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
		Assets: []Asset{
			*FixtureAsset("ruby-2-4-2"),
		},
		Hooks: []HookConfig{
			*FixtureHookConfig("hook1"),
		},
	}
}

// FixtureCheckConfig returns a fixture for a CheckConfig object.
func FixtureCheckConfig(id string) *CheckConfig {
	interval := uint32(60)

	return &CheckConfig{
		Name:          id,
		Interval:      interval,
		Subscriptions: []string{"linux"},
		Command:       "command",
		Handlers:      []string{},
		RuntimeAssets: []string{"ruby-2-4-2"},
		CheckHooks:    []HookList{*FixtureHookList("hook1")},
		Environment:   "default",
		Organization:  "default",
		Publish:       true,
		Cron:          "",
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
