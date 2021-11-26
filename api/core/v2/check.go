package v2

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"time"

	jsoniter "github.com/json-iterator/go"
	cron "github.com/robfig/cron/v3"
	utilstrings "github.com/sensu/sensu-go/api/core/v2/internal/stringutil"
)

const (
	// CheckRequestType is the message type string for check request.
	CheckRequestType = "check_request"

	// ChecksResource is the name of this resource type
	ChecksResource = "checks"

	// DefaultSplayCoverage is the default splay coverage for proxy check requests
	DefaultSplayCoverage = 90.0

	// NagiosOutputMetricFormat is the accepted string to represent the output metric format of
	// Nagios Perf Data
	NagiosOutputMetricFormat = "nagios_perfdata"

	// GraphiteOutputMetricFormat is the accepted string to represent the output metric format of
	// Graphite Plain Text
	GraphiteOutputMetricFormat = "graphite_plaintext"

	// OpenTSDBOutputMetricFormat is the accepted string to represent the output metric format of
	// OpenTSDB Line
	OpenTSDBOutputMetricFormat = "opentsdb_line"

	// InfluxDBOutputMetricFormat is the accepted string to represent the output metric format of
	// InfluxDB Line
	InfluxDBOutputMetricFormat = "influxdb_line"

	// PrometheusOutputMetricFormat is the accepted string to represent the output metric format of
	// Prometheus Exposition Text Format
	PrometheusOutputMetricFormat = "prometheus_text"

	// KeepaliveCheckName is the name of the check that is created when a
	// keepalive timeout occurs.
	KeepaliveCheckName = "keepalive"

	// RegistrationCheckName is the name of the check that is created when an
	// entity sends a keepalive and the entity does not yet exist in the store.
	RegistrationCheckName = "registration"

	// MemoryScheduler indicates that a check is scheduled in-memory.
	MemoryScheduler = "memory"

	// EtcdScheduler indicates that a check is scheduled with etcd leases and
	// watchers.
	EtcdScheduler = "etcd"

	// PostgresScheduler indicates that a check is scheduled with postgresql,
	// using transactions and asynchronous notification (NOTIFY).
	PostgresScheduler = "postgres"
)

// OutputMetricFormats represents all the accepted output_metric_format's a check can have
var OutputMetricFormats = []string{NagiosOutputMetricFormat, GraphiteOutputMetricFormat, OpenTSDBOutputMetricFormat, InfluxDBOutputMetricFormat, PrometheusOutputMetricFormat}

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
	c.State = EventPassingState

	return c
}

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
		OutputMetricTags:     c.OutputMetricTags,
		EnvVars:              c.EnvVars,
		DiscardOutput:        c.DiscardOutput,
		MaxOutputSize:        c.MaxOutputSize,
		Scheduler:            c.Scheduler,
	}
	if check.Labels == nil {
		check.Labels = make(map[string]string)
	}
	if check.Annotations == nil {
		check.Annotations = make(map[string]string)
	}
	return check
}

// SetNamespace sets the namespace of the resource.
func (c *Check) SetNamespace(namespace string) {
	c.Namespace = namespace
}

// SetObjectMeta sets the meta of the resource.
func (c *Check) SetObjectMeta(meta ObjectMeta) {
	c.ObjectMeta = meta
}

// SetName sets the name of the resource.
func (c *Check) SetName(name string) {
	c.Name = name
}

// StorePrefix returns the path prefix to this resource in the store
func (c *Check) StorePrefix() string {
	return ChecksResource
}

// URIPath returns the path component of a check URI.
func (c *Check) URIPath() string {
	if c.Namespace == "" {
		return path.Join(URLPrefix, ChecksResource, url.PathEscape(c.Name))
	}
	return path.Join(URLPrefix, "namespaces", url.PathEscape(c.Namespace), ChecksResource, url.PathEscape(c.Name))
}

// Validate returns an error if the check does not pass validation tests.
func (c *Check) Validate() error {
	if err := ValidateName(c.Name); err != nil {
		return errors.New("check name " + err.Error())
	}
	if c.Publish {
		if c.Cron != "" {
			if c.Interval > 0 {
				return errors.New("must only specify either an interval or a cron schedule")
			}

			if _, err := cron.ParseStandard(c.Cron); err != nil {
				return fmt.Errorf("check cron string is invalid: %w", err)
			}
		} else {
			if c.Interval < 1 {
				return errors.New("check interval must be greater than or equal to 1")
			}
		}
	}

	if c.Ttl > 0 && c.Ttl <= int64(c.Interval) {
		return errors.New("ttl must be greater than check interval")
	}
	if c.Ttl > 0 && c.Ttl < 5 {
		return errors.New("minimum ttl is 5 seconds")
	}

	for _, assetName := range c.RuntimeAssets {
		if err := ValidateAssetName(assetName); err != nil {
			return fmt.Errorf("asset's %s", err)
		}
	}

	for _, subscription := range c.Subscriptions {
		if subscription == "" {
			return fmt.Errorf("subscriptions cannot be empty strings")
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

	if c.MaxOutputSize < 0 {
		return fmt.Errorf("MaxOutputSize must be >= 0")
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
	if c.Pipelines == nil {
		c.Pipelines = []*ResourceReference{}
	}

	type Clone Check
	clone := &Clone{}
	*clone = Clone(*c)

	return jsoniter.Marshal(clone)
}

// MergeWith updates the current Check with the history of the check given as
// an argument, updating the current check's history appropriately.
func (c *Check) MergeWith(prevCheck *Check) {
	history := prevCheck.History
	histEntry := CheckHistory{
		Status:   c.Status,
		Executed: c.Executed,
	}

	history = append(history, histEntry)
	if len(history) > 21 {
		history = history[1:]
	}

	c.History = history
	c.LastOK = prevCheck.LastOK
	c.Occurrences = prevCheck.Occurrences
	c.OccurrencesWatermark = prevCheck.OccurrencesWatermark
	updateCheckState(c)

	// This has to be done after the call to updateCheckState, as that function is what
	// sets the value for c.State that is used below, but the order can't be switched
	// around as updateCheckState relies on the latest item (specifically, its status)
	// being present in c.History.
	// NB! This has been disabled for 5.x releases.
	// c.History[len(c.History)-1].Flapping = c.State == EventFlappingState
}

// ValidateOutputMetricFormat returns an error if the string is not a valid metric
// format
func ValidateOutputMetricFormat(format string) error {
	if utilstrings.InArray(format, OutputMetricFormats) {
		return nil
	}
	return errors.New("output metric format is not valid")
}

// previousOccurrence returns the most recent CheckHistory item, excluding the current result.
func (c *Check) previousOccurrence() *CheckHistory {
	if len(c.History) < 2 {
		return nil
	}
	return &c.History[len(c.History)-2]
}

// DEPRECATED, DO NOT USE! Events should be ordered FIFO.
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

func (c *Check) RBACName() string {
	return "checks"
}

func (c *CheckConfig) RBACName() string {
	return "checks"
}
