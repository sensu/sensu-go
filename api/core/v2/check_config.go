package v2

import (
	"errors"
	fmt "fmt"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	cron "github.com/robfig/cron/v3"
	stringsutil "github.com/sensu/sensu-go/api/core/v2/internal/stringutil"
)

// FixtureCheckConfig returns a fixture for a CheckConfig object.
func FixtureCheckConfig(id string) *CheckConfig {
	interval := uint32(60)
	timeout := uint32(0)

	check := &CheckConfig{
		ObjectMeta:        NewObjectMeta(id, "default"),
		Interval:          interval,
		Subscriptions:     []string{"linux"},
		Command:           "command",
		RuntimeAssets:     []string{"ruby-2-4-2"},
		CheckHooks:        []HookList{*FixtureHookList("hook1")},
		Publish:           true,
		Ttl:               0,
		Timeout:           timeout,
		LowFlapThreshold:  20,
		HighFlapThreshold: 60,
	}
	return check
}

// NewCheckConfig creates a new CheckConfig.
func NewCheckConfig(meta ObjectMeta) *CheckConfig {
	return &CheckConfig{ObjectMeta: meta}
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
	if c.Pipelines == nil {
		c.Pipelines = []*ResourceReference{}
	}

	for _, ref := range c.Pipelines {
		// default to core/v2 for APIVersion when not set
		if ref.APIVersion == "" {
			ref.APIVersion = "core/v2"
		}
		// default to Pipeline for Type when not set
		if ref.Type == "" {
			ref.Type = "Pipeline"
		}
	}

	type Clone CheckConfig
	clone := &Clone{}
	*clone = Clone(*c)

	return jsoniter.Marshal(clone)
}

// SetNamespace sets the namespace of the resource.
func (c *CheckConfig) SetNamespace(namespace string) {
	c.Namespace = namespace
}

// SetObjectMeta sets the meta of the resource.
func (c *CheckConfig) SetObjectMeta(meta ObjectMeta) {
	c.ObjectMeta = meta
}

// StorePrefix returns the path prefix to this resource in the store
func (c *CheckConfig) StorePrefix() string {
	return ChecksResource
}

// URIPath returns the path component of a CheckConfig URI.
func (c *CheckConfig) URIPath() string {
	if c.Namespace == "" {
		return path.Join(URLPrefix, ChecksResource, url.PathEscape(c.Name))
	}
	return path.Join(URLPrefix, "namespaces", url.PathEscape(c.Namespace), ChecksResource, url.PathEscape(c.Name))
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
			return fmt.Errorf("check cron string is invalid: %w", err)
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

	return c.Subdue.Validate()
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

//
// Sorting
//

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

// CheckConfigFields returns a set of fields that represent that resource
func CheckConfigFields(r Resource) map[string]string {
	resource := r.(*CheckConfig)
	fields := map[string]string{
		"check.name":           resource.ObjectMeta.Name,
		"check.namespace":      resource.ObjectMeta.Namespace,
		"check.handlers":       strings.Join(resource.Handlers, ","),
		"check.publish":        strconv.FormatBool(resource.Publish),
		"check.round_robin":    strconv.FormatBool(resource.RoundRobin),
		"check.runtime_assets": strings.Join(resource.RuntimeAssets, ","),
		"check.subscriptions":  strings.Join(resource.Subscriptions, ","),
	}
	stringsutil.MergeMapWithPrefix(fields, resource.ObjectMeta.Labels, "check.labels.")

	pipelineIDs := []string{}
	for _, pipeline := range resource.Pipelines {
		pipelineIDs = append(pipelineIDs, pipeline.ResourceID())
	}
	fields["check.pipelines"] = strings.Join(pipelineIDs, ",")

	return fields
}
