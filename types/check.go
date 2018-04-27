package types

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/robfig/cron"
	"github.com/sensu/sensu-go/types/dynamic"
	"github.com/sensu/sensu-go/util/eval"
	utilstrings "github.com/sensu/sensu-go/util/strings"
)

// CheckRequestType is the message type string for check request.
const CheckRequestType = "check_request"

// DefaultSplayCoverage is the default splay coverage for proxy check requests
const DefaultSplayCoverage = 90.0

// NagiosMetricFormat is the accepted string to represent the metric format of
// Nagios Perf Data
const NagiosMetricFormat = "nagios_perfdata"

// GraphiteMetricFormat is the accepted string to represent the metric format of
// Graphite Plain Text
const GraphiteMetricFormat = "graphite_plaintext"

// OpenTSDBMetricFormat is the accepted string to represent the metric format of
// OpenTSDB Line
const OpenTSDBMetricFormat = "opentsdb_line"

// InfluxDBMetricFormat is the accepted string to represent the metric format of
// InfluxDB Line
const InfluxDBMetricFormat = "influxdb_line"

// MetricFormats represents all the accepted metric_format's a check can have
var MetricFormats = []string{NagiosMetricFormat, GraphiteMetricFormat, OpenTSDBMetricFormat, InfluxDBMetricFormat}

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
		MetricFormat:       c.MetricFormat,
		MetricHandlers:     c.MetricHandlers,
	}
	return check
}

// Validate returns an error if the check does not pass validation tests.
func (c *Check) Validate() error {
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

	if c.MetricFormat != "" {
		if err := ValidateMetricFormat(c.MetricFormat); err != nil {
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
	if c == nil {
		return []byte("null"), nil
	}
	if c.Subscriptions == nil {
		c.Subscriptions = []string{}
	}
	if c.Handlers == nil {
		c.Handlers = []string{}
	}

	// Only use dynamic marshaling if there are dynamic attributes.
	// Otherwise, use default json marshaling.
	if len(c.ExtendedAttributes) > 0 {
		return dynamic.Marshal(c)
	}

	type Clone Check
	clone := &Clone{}
	*clone = Clone(*c)

	return jsoniter.Marshal(clone)
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
	if c == nil {
		return []byte("null"), nil
	}
	if c.Subscriptions == nil {
		c.Subscriptions = []string{}
	}
	if c.Handlers == nil {
		c.Handlers = []string{}
	}

	// Only use dynamic marshaling if there are dynamic attributes.
	// Otherwise, use default json marshaling.
	if len(c.ExtendedAttributes) > 0 {
		return dynamic.Marshal(c)
	}

	type Clone CheckConfig
	clone := &Clone{}
	*clone = Clone(*c)

	return jsoniter.Marshal(clone)
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

	if c.MetricFormat != "" {
		if err := ValidateMetricFormat(c.MetricFormat); err != nil {
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

	return eval.ValidateStatements(p.EntityAttributes, false)
}

// ValidateMetricFormat returns an error if the string is not a valid metric
// format
func ValidateMetricFormat(format string) error {
	if utilstrings.InArray(format, MetricFormats) {
		return nil
	}
	return errors.New("metric format is not valid")
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
func (c *Check) MergeWith(prevCheck *Check) {
	history := prevCheck.History
	histEntry := CheckHistory{
		Status:   c.Status,
		Executed: c.Executed,
	}

	history = append(history, histEntry)
	sort.Sort(ByExecuted(history))
	if len(history) > 21 {
		history = history[1:]
	}

	c.History = history
	c.LastOK = prevCheck.LastOK
	c.Occurrences = prevCheck.Occurrences
	c.OccurrencesWatermark = prevCheck.OccurrencesWatermark
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

// URIPath returns the path component of a Check URI.
func (c *Check) URIPath() string {
	return fmt.Sprintf("/checks/%s", url.PathEscape(c.Name))
}

//
// Sorting

type cmpCheckConfig func(a, b *CheckConfig) bool

// SortCheckConfigsByPredicate can be used to sort a given collection using a given
// predicate.
func SortCheckConfigsByPredicate(cs []*CheckConfig, fn cmpCheckConfig) sort.Interface {
	return &checkSorter{checks: cs, byFn: fn}
}

// SortCheckConfigsByName can be used to sort a given collection of checks by their
// names.
func SortCheckConfigsByName(es []*CheckConfig, asc bool) sort.Interface {
	if asc {
		return SortCheckConfigsByPredicate(es, func(a, b *CheckConfig) bool {
			return a.Name < a.Name
		})
	}
	return SortCheckConfigsByPredicate(es, func(a, b *CheckConfig) bool {
		return a.Name > a.Name
	})
}

type checkSorter struct {
	checks []*CheckConfig
	byFn   cmpCheckConfig
}

// Len implements sort.Interface.
func (s *checkSorter) Len() int {
	return len(s.checks)
}

// Swap implements sort.Interface.
func (s *checkSorter) Swap(i, j int) {
	s.checks[i], s.checks[j] = s.checks[j], s.checks[i]
}

// Less implements sort.Interface.
func (s *checkSorter) Less(i, j int) bool {
	return s.byFn(s.checks[i], s.checks[j])
}
