package types

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/robfig/cron"
	"github.com/sensu/sensu-go/js"
	utilstrings "github.com/sensu/sensu-go/util/strings"
)

// CheckRequestType is the message type string for check request.
const CheckRequestType = "check_request"

// DefaultSplayCoverage is the default splay coverage for proxy check requests
const DefaultSplayCoverage = 90.0

// NagiosOutputMetricFormat is the accepted string to represent the output metric format of
// Nagios Perf Data
const NagiosOutputMetricFormat = "nagios_perfdata"

// GraphiteOutputMetricFormat is the accepted string to represent the output metric format of
// Graphite Plain Text
const GraphiteOutputMetricFormat = "graphite_plaintext"

// OpenTSDBOutputMetricFormat is the accepted string to represent the output metric format of
// OpenTSDB Line
const OpenTSDBOutputMetricFormat = "opentsdb_line"

// InfluxDBOutputMetricFormat is the accepted string to represent the output metric format of
// InfluxDB Line
const InfluxDBOutputMetricFormat = "influxdb_line"

// OutputMetricFormats represents all the accepted output_metric_format's a check can have
var OutputMetricFormats = []string{NagiosOutputMetricFormat, GraphiteOutputMetricFormat, OpenTSDBOutputMetricFormat, InfluxDBOutputMetricFormat}

// NewCheck creates a new Check. It copies the fields from CheckConfig that
// match with Check's fields.
//
// Because CheckConfig uses extended attributes, embedding CheckConfig was
// deemed to be too complicated, due to interactions between promoted methods
// and encoding/json.
func NewCheck(c *CheckConfig) *Check {
	check := &Check{
		ObjectMeta: ObjectMeta{
			Name:        c.Name,
			Namespace:   c.Namespace,
			Labels:      c.Labels,
			Annotations: c.Annotations,
		},
		Command:              c.Command,
		Handlers:             c.Handlers,
		HighFlapThreshold:    c.HighFlapThreshold,
		Interval:             c.Interval,
		LowFlapThreshold:     c.LowFlapThreshold,
		Publish:              c.Publish,
		RuntimeAssets:        c.RuntimeAssets,
		Subscriptions:        c.Subscriptions,
		ProxyEntityName:      c.ProxyEntityName,
		CheckHooks:           c.CheckHooks,
		Stdin:                c.Stdin,
		Subdue:               c.Subdue,
		Cron:                 c.Cron,
		Ttl:                  c.Ttl,
		Timeout:              c.Timeout,
		ProxyRequests:        c.ProxyRequests,
		RoundRobin:           c.RoundRobin,
		OutputMetricFormat:   c.OutputMetricFormat,
		OutputMetricHandlers: c.OutputMetricHandlers,
		EnvVars:              c.EnvVars,
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
	if c.ProxyEntityName != "" {
		if err := ValidateName(c.ProxyEntityName); err != nil {
			return errors.New("proxy entity name " + err.Error())
		}
	}

	if c.ProxyRequests != nil {
		if err := c.ProxyRequests.Validate(); err != nil {
			return err
		}
	}

	if c.OutputMetricFormat != "" {
		if err := ValidateOutputMetricFormat(c.OutputMetricFormat); err != nil {
			return err
		}
	}

	if c.LowFlapThreshold != 0 && c.HighFlapThreshold != 0 && c.LowFlapThreshold >= c.HighFlapThreshold {
		return errors.New("invalid flap thresholds")
	}

	if err := ValidateEnvVars(c.EnvVars); err != nil {
		return err
	}

	return c.Subdue.Validate()
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

	type Clone Check
	clone := &Clone{}
	*clone = Clone(*c)

	return jsoniter.Marshal(clone)
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

	type Clone CheckConfig
	clone := &Clone{}
	*clone = Clone(*c)

	return jsoniter.Marshal(clone)
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

	if c.Namespace == "" {
		return errors.New("namespace must be set")
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
	if c.ProxyEntityName != "" {
		if err := ValidateName(c.ProxyEntityName); err != nil {
			return errors.New("proxy entity name " + err.Error())
		}
	}

	if c.ProxyRequests != nil {
		if err := c.ProxyRequests.Validate(); err != nil {
			return err
		}
	}

	if c.OutputMetricFormat != "" {
		if err := ValidateOutputMetricFormat(c.OutputMetricFormat); err != nil {
			return err
		}
	}

	if c.LowFlapThreshold != 0 && c.HighFlapThreshold != 0 && c.LowFlapThreshold >= c.HighFlapThreshold {
		return errors.New("invalid flap thresholds")
	}

	if err := ValidateEnvVars(c.EnvVars); err != nil {
		return err
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

	return js.ParseExpressions(p.EntityAttributes)
}

// ValidateOutputMetricFormat returns an error if the string is not a valid metric
// format
func ValidateOutputMetricFormat(format string) error {
	if utilstrings.InArray(format, OutputMetricFormats) {
		return nil
	}
	return errors.New("output metric format is not valid")
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

	check := &CheckConfig{
		ObjectMeta: ObjectMeta{
			Name:      id,
			Namespace: "default",
		},
		Interval:             interval,
		Subscriptions:        []string{"linux"},
		Command:              "command",
		Handlers:             []string{},
		RuntimeAssets:        []string{"ruby-2-4-2"},
		CheckHooks:           []HookList{*FixtureHookList("hook1")},
		Publish:              true,
		Cron:                 "",
		Ttl:                  0,
		Timeout:              timeout,
		OutputMetricHandlers: []string{},
		OutputMetricFormat:   "",
	}
	return check
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

// URIPath returns the path component of a CheckConfig URI.
func (c *CheckConfig) URIPath() string {
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
			return a.Name < b.Name
		})
	}
	return SortCheckConfigsByPredicate(es, func(a, b *CheckConfig) bool {
		return a.Name > b.Name
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

// IsSubdued returns true if the check is subdued at the current time.
// It returns false otherwise.
func (c *CheckConfig) IsSubdued() bool {
	subdue := c.GetSubdue()
	if subdue == nil {
		return false
	}
	subdued, err := subdue.InWindows(time.Now())
	if err != nil {
		return false
	}
	return subdued
}
