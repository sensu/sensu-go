package types

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/robfig/cron"
	"github.com/sensu/sensu-go/types/dynamic"
	"github.com/sensu/sensu-go/util/eval"
)

// CheckRequestType is the message type string for check request.
const CheckRequestType = "check_request"

// DefaultSplayCoverage is the default splay coverage for proxy check requests
const DefaultSplayCoverage = 90.0

// NewCheck creates a new Check. It copies the fields from CheckConfig that
// match with Check's fields.
//
// Because CheckConfig uses extended attributes, embedding CheckConfig was
// deemed to be too complicated, due to interactions between promoted methods
// and encoding/json.
func NewCheck(c *CheckConfig) *Check {
	check := &Check{
		Command:            c.Command,
		Environment:        c.Environment,
		Handlers:           c.Handlers,
		HighFlapThreshold:  c.HighFlapThreshold,
		Interval:           c.Interval,
		LowFlapThreshold:   c.LowFlapThreshold,
		Name:               c.Name,
		Organization:       c.Organization,
		Publish:            c.Publish,
		RuntimeAssets:      c.RuntimeAssets,
		Subscriptions:      c.Subscriptions,
		ExtendedAttributes: c.ExtendedAttributes,
		ProxyEntityID:      c.ProxyEntityID,
		CheckHooks:         c.CheckHooks,
		Stdin:              c.Stdin,
		Subdue:             c.Subdue,
		Cron:               c.Cron,
		Ttl:                c.Ttl,
		Timeout:            c.Timeout,
		ProxyRequests:      c.ProxyRequests,
		RoundRobin:         c.RoundRobin,
	}
	return check
}

// Validate returns an error if the check does not pass validation tests.
func (c *Check) Validate() error {
	if err := ValidateName(c.Name); err != nil {
		return errors.New("check name " + err.Error())
	}
	if c.Status < 0 {
		return errors.New("check status must be greater than or equal to 0")
	}
	if c.Cron != "" {
		if c.Interval > 0 {
			return errors.New("must only specify either an interval or a cron schedule")
		}

		if _, err := cron.ParseStandard(c.Cron); err != nil {
			return errors.New("check cron string is invalid")
		}
	} else {
		if c.Interval < 1 {
			return errors.New("check interval must be greater than or equal to 1")
		}
	}

	if c.Ttl > 0 && c.Ttl <= int64(c.Interval) {
		return errors.New("ttl must be greater than check interval")
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

	if c.ProxyRequests != nil {
		if err := c.ProxyRequests.Validate(); err != nil {
			return err
		}
	}

	return c.Subdue.Validate()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (c *Check) UnmarshalJSON(b []byte) error {
	return dynamic.Unmarshal(b, c)
}

// MarshalJSON implements the json.Marshaler interface.
func (c *Check) MarshalJSON() ([]byte, error) {
	return dynamic.Marshal(c)
}

// SetExtendedAttributes sets the serialized ExtendedAttributes of c.
func (c *Check) SetExtendedAttributes(e []byte) {
	c.ExtendedAttributes = e
}

// Get implements govaluate.Parameters
func (c *Check) Get(name string) (interface{}, error) {
	return dynamic.GetField(c, name)
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
		if c.Interval > 0 {
			return errors.New("must only specify either an interval or a cron schedule")
		}

		if _, err := cron.ParseStandard(c.Cron); err != nil {
			return errors.New("check cron string is invalid")
		}
	}

	if c.Interval == 0 && c.Cron == "" {
		return errors.New("check interval must be greater than 0 or a valid cron schedule must be provided")
	}

	if c.Environment == "" {
		return errors.New("environment cannot be empty")
	}

	if c.Organization == "" {
		return errors.New("organization must be set")
	}

	if c.Ttl > 0 && c.Ttl <= int64(c.Interval) {
		return errors.New("ttl must be greater than check interval")
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

	if c.ProxyRequests != nil {
		if err := c.ProxyRequests.Validate(); err != nil {
			return err
		}
	}

	return c.Subdue.Validate()
}

// Validate returns an error if the ProxyRequests does not pass validation tests
func (p *ProxyRequests) Validate() error {
	if p.SplayCoverage > 100 {
		return errors.New("proxy request splay coverage must be between 0 and 100")
	}

	if (p.Splay) && (p.SplayCoverage == 0) {
		return errors.New("proxy request splay coverage must be greater than 0 if splay is enabled")
	}

	return eval.ValidateStatements(p.EntityAttributes)
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
	c.LastOK = chk.LastOK
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
	timeout := uint32(0)

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
		Ttl:           0,
		Timeout:       timeout,
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

	c := NewCheck(config)
	c.Issued = t
	c.Executed = t + 1
	c.Duration = 1.0
	c.History = history

	return c
}

// FixtureProxyRequests returns a fixture for a ProxyRequests object.
func FixtureProxyRequests(splay bool) *ProxyRequests {
	splayCoverage := uint32(0)
	if splay {
		splayCoverage = DefaultSplayCoverage
	}
	return &ProxyRequests{
		Splay:         splay,
		SplayCoverage: splayCoverage,
	}
}
