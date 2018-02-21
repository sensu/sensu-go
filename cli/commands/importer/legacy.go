package importer

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/elements/report"
	"github.com/sensu/sensu-go/types"
)

// NewSensuV1SettingsImporter returns a new importer configured to import
// settings for Sensu v1.
func NewSensuV1SettingsImporter(org, env string, client client.APIClient) *Importer {
	return NewImporter(
		&LegacyCheckImporter{
			Org:      org,
			Env:      env,
			SaveFunc: client.CreateCheck,
		},
		&LegacyHandlerImporter{
			Org:      org,
			Env:      env,
			SaveFunc: client.CreateHandler,
		},
		&LegacyMutatorImporter{
			Org:      org,
			Env:      env,
			SaveFunc: client.CreateMutator,
		},
		&LegacyFilterImporter{
			Org:      org,
			Env:      env,
			SaveFunc: client.CreateFilter,
		},
		&LegacyExtensionImporter{Org: org, Env: env},
		&LegacyEntityImporter{Org: org, Env: env},
		&LegacyAPIImporter,
		&LegacySensuImporter,
		&LegacyTransportImporter,
	)
}

//
// Mutators
//

// LegacyMutatorImporter provides utility to import Sensu v1 mutator definitiions
type LegacyMutatorImporter struct {
	Org      string
	Env      string
	SaveFunc func(*types.Mutator) error

	reporter *report.Writer
	mutators []*types.Mutator
}

// Name of the importer
func (i *LegacyMutatorImporter) Name() string {
	return "mutators"
}

// SetReporter ...
func (i *LegacyMutatorImporter) SetReporter(r *report.Writer) {
	reporter := r.WithValue("component", i.Name())
	i.reporter = &reporter
}

// Import given data
func (i *LegacyMutatorImporter) Import(data map[string]interface{}) error {
	if vals, ok := data["mutators"].(map[string]interface{}); ok {
		for name, cfg := range vals {
			mutator := i.newMutator(name)
			i.applyCfg(&mutator, cfg.(map[string]interface{}))

			i.mutators = append(i.mutators, &mutator)
		}
	} else if _, ok = data["mutators"]; ok {
		i.reporter.Warn("mutators present but do not appear to be a hash; please mutator format")
	}

	return nil
}

// Validate the given mutators
func (i *LegacyMutatorImporter) Validate() error {
	for _, mutator := range i.mutators {
		if err := mutator.Validate(); err != nil {
			i.reporter.Errorf(
				"mutator '%s' failed validation w/ '%s'",
				mutator.Name,
				err,
			)
		}
	}

	return nil
}

// Save calls API with mutators
func (i *LegacyMutatorImporter) Save() (int, error) {
	for _, mutator := range i.mutators {
		if err := i.SaveFunc(mutator); err != nil {
			i.reporter.Fatalf(
				"unable to persist mutator '%s' w/ error '%s'",
				mutator.Name,
				err,
			)
		} else {
			i.reporter.Debugf("mutator '%s' imported", mutator.Name)
		}
	}

	return len(i.mutators), nil
}

func (i *LegacyMutatorImporter) newMutator(name string) types.Mutator {
	return types.Mutator{
		Name:         name,
		Organization: i.Org,
		Environment:  i.Env,
	}
}

//
// example #1:
//
// {
//   "mutators": {
//     "example_mutator": {
//       "command": "example_mutator.rb"
//     }
//   }
// }
//
// NOTE: Any of these fields should cause failure:
//
// NOTE: Fields that are currently ignored
//
// NOTE: Valid keys
//
//   - `command`
//   - `timeout`
//
func (i *LegacyMutatorImporter) applyCfg(mutator *types.Mutator, cfg map[string]interface{}) {

	//
	// Apply values

	if val, ok := cfg["command"].(string); ok {
		mutator.Command = val
	}

	if val, ok := cfg["timeout"].(float64); ok {
		mutator.Timeout = uint32(val)
	}
}

//
// Checks
//

// LegacyCheckImporter provides utility to import Sensu v1 check definitiions
type LegacyCheckImporter struct {
	Org      string
	Env      string
	SaveFunc func(*types.CheckConfig) error

	reporter *report.Writer
	checks   []*types.CheckConfig
}

// Name of the importer
func (i *LegacyCheckImporter) Name() string {
	return "checks"
}

// SetReporter ...
func (i *LegacyCheckImporter) SetReporter(r *report.Writer) {
	reporter := r.WithValue("component", i.Name())
	i.reporter = &reporter
}

// Import given data
func (i *LegacyCheckImporter) Import(data map[string]interface{}) error {
	if vals, ok := data["checks"].(map[string]interface{}); ok {
		for name, cfg := range vals {
			check := i.newCheck(name)
			i.applyCfg(&check, cfg.(map[string]interface{}))

			check.Name = name
			check.Organization = i.Org
			check.Environment = i.Env
			i.checks = append(i.checks, &check)
		}
	} else if _, ok = data["checks"]; ok {
		i.reporter.Warn("checks present but do not appear to be a hash; please check format")
	}

	return nil
}

// Validate the given checks
func (i *LegacyCheckImporter) Validate() error {
	for _, check := range i.checks {
		if err := check.Validate(); err != nil {
			i.reporter.Errorf(
				"check '%s' failed validation w/ '%s'",
				check.Name,
				err,
			)
		}
	}

	return nil
}

// Save calls API with checks
func (i *LegacyCheckImporter) Save() (int, error) {
	for _, check := range i.checks {
		if err := i.SaveFunc(check); err != nil {
			i.reporter.
				WithValue("name", check.Name).
				WithValue("error", err).
				Fatalf(
					"unable to persist check '%s' w/ error '%s'",
					check.Name,
					err,
				)
		} else {
			i.reporter.Debugf("check '%s' imported", check.Name)
		}
	}

	return len(i.checks), nil
}

func (i *LegacyCheckImporter) newCheck(name string) types.CheckConfig {
	return types.CheckConfig{
		Name:         name,
		Organization: i.Org,
		Environment:  i.Env,
	}
}

//
// example #1:
//
// {
//   "checks": {
//     "sensu-website": {
//       "command": "check-http.rb -u https://sensuapp.org",
//       "subscribers": [
//         "production"
//       ],
//       "interval": 60,
//       "handler": "slack",
//     }
//   }
// }
//
// example #2:
//
// {
//   "checks": {
//     "disk-check": {
//       "command": "check-disk.rb",
//       "subscribers": [
//         "production"
//       ],
//       "cron": "0 0 * * *",
//       "handlers": ["slack", "pagerduty"],
//     }
//   }
// }
//
// NOTE: Any of these fields should cause failure:
//
//   - `type` if value is metric
//   - `extension`
//   - `standalone`; we'll have to figure out what to do with this to avoid conflicts
//   - `auto_resolve` if `false`
//   - `force_resolve`
//   - `handle` if false
//   - `source`
//   - `subdue`
//   - `contact`
//   - `contacts`
//   - `proxy_requests`
//
// NOTE: Fields that are currently ignored
//
//   - `aggregate`
//   - `aggregates`
//
// NOTE: Valid keys
//
//   - `type` if `standard` or empty
//   - `command`
//   - `cron`
//   - `subscribers`
//   - `handle` if true
//   - `handler`
//   - `handlers`
//   - `high_flap_threshold`
//   - `interval`
//   - `low_flap_threshold`
//   - `publish`
//   - `timeout`
//   - `ttl`
//
func (i *LegacyCheckImporter) applyCfg(check *types.CheckConfig, cfg map[string]interface{}) {
	reporter := i.reporter.WithValue("name", check.Name)

	//
	// Capture critical unsupported attributes and fail

	// "type": "metric"
	if val, ok := cfg["type"].(string); ok && val == "metric" {
		reporter.Error("metric checks have not been implemented at this time")
	}

	// "extension": true
	if val, ok := cfg["extension"].(bool); ok && val {
		reporter.Error("extension are not yet supported at this time")
	}

	// "standalone": true
	if val, ok := cfg["standalone"].(bool); ok && val {
		reporter.Error("standalone are not supported at this time")
	}

	// "auto_resolve": false
	if val, ok := cfg["auto_resolve"].(bool); ok && !val {
		reporter.Error(unsupportedAttr("checks", "auto_resolve"))
	}

	// "force_resolve": true|false
	if _, ok := cfg["force_resolve"]; ok {
		reporter.Error(unsupportedAttr("checks", "force_resolve"))
	}

	// "handle": false
	if val, ok := cfg["handle"].(bool); ok && !val {
		reporter.Error("unhandled checks are not supported at this time")
	}

	// "source": string
	if _, ok := cfg["source"]; ok {
		reporter.Error(unsupportedAttr("checks", "source"))
	}

	// "subdue": map
	if _, ok := cfg["subdue"]; ok {
		reporter.Error(unsupportedAttr("checks", "subdue"))
	}

	// "contact": string
	if _, ok := cfg["contact"]; ok {
		reporter.Error(unsupportedAttr("checks", "contact"))
	}

	// "contacts": []string
	if _, ok := cfg["contacts"]; ok {
		reporter.Error(unsupportedAttr("checks", "contacts"))
	}

	// "proxy_requests": map
	if _, ok := cfg["proxy_requests"]; ok {
		reporter.Error(unsupportedAttr("checks", "proxy_requests"))
	}

	//
	// Capture unsupported attributes and warn user

	// "aggregate": string
	if _, ok := cfg["aggregate"]; ok {
		reporter.Warn(unsupportedAttr("checks", "aggregate"))
	}

	// "aggregates": []string
	if _, ok := cfg["aggregate"]; ok {
		reporter.Warn(unsupportedAttr("checks", "aggregate"))
	}

	//
	// Apply values

	if val, ok := cfg["command"].(string); ok {
		check.Command = val
	}

	if val, ok := cfg["cron"].(string); ok {
		check.Cron = val
	}

	if vals, ok := cfg["subscribers"].([]interface{}); ok {
		for _, val := range vals {
			if val, ok := val.(string); ok {
				check.Subscriptions = append(check.Subscriptions, val)
			}
		}
	}

	if val, ok := cfg["interval"].(float64); ok {
		check.Interval = uint32(val)
	}

	if val, ok := cfg["handler"].(string); ok {
		check.Handlers = []string{val}
	}

	if val, ok := cfg["handlers"].([]string); ok {
		check.Handlers = append(check.Handlers, val...)
	}

	check.Publish = true
	if val, ok := cfg["publish"].(bool); ok {
		check.Publish = val
	}

	if val, ok := cfg["ttl"].(int64); ok {
		check.Ttl = val
	}

	if val, ok := cfg["timeout"].(int64); ok {
		check.Timeout = uint32(val)
	}

	if val, ok := cfg["low_flap_threshold"].(int64); ok {
		check.LowFlapThreshold = uint32(val)
	}

	if val, ok := cfg["high_flap_threshold"].(int64); ok {
		check.HighFlapThreshold = uint32(val)
	}
}

//
// Filters
//

// LegacyFilterImporter provides utility to import Sensu v1 filter definitiions
type LegacyFilterImporter struct {
	Org      string
	Env      string
	SaveFunc func(*types.EventFilter) error

	filters  []*types.EventFilter
	reporter *report.Writer
}

// Name of the importer
func (i *LegacyFilterImporter) Name() string {
	return "filters"
}

// SetReporter ...
func (i *LegacyFilterImporter) SetReporter(r *report.Writer) {
	reporter := r.WithValue("component", i.Name())
	i.reporter = &reporter
}

// Import given data
func (i *LegacyFilterImporter) Import(data map[string]interface{}) error {
	if vals, ok := data["filters"].(map[string]interface{}); ok {
		for name, cfg := range vals {
			filter := i.newFilter(name)
			i.applyCfg(&filter, cfg.(map[string]interface{}))

			i.filters = append(i.filters, &filter)
		}
	} else if _, ok = data["filters"]; ok {
		i.reporter.Warn("filters present but do not appear to be a hash; please properly format the filters")
	}

	return nil
}

// Validate the given filters
func (i *LegacyFilterImporter) Validate() error {
	for _, filter := range i.filters {
		if err := filter.Validate(); err != nil {
			i.reporter.Errorf(
				"filter '%s' failed validation w/ '%s'",
				filter.Name,
				err,
			)
		}
	}

	return nil
}

// Save calls API with filters
func (i *LegacyFilterImporter) Save() (int, error) {
	for _, filter := range i.filters {
		if err := i.SaveFunc(filter); err != nil {
			i.reporter.Fatalf(
				"unable to persist filter '%s' w/ error '%s'",
				filter.Name,
				err,
			)
		} else {
			i.reporter.Debugf("filter '%s' imported", filter.Name)
		}
	}

	return len(i.filters), nil
}

func (i *LegacyFilterImporter) newFilter(name string) types.EventFilter {
	return types.EventFilter{
		Name:         name,
		Organization: i.Org,
		Environment:  i.Env,
	}
}

//
// example #1:
//
// {
//   "filters": {
//     "state_change_only": {
//       "negate": false,
//       "attributes": {
//         "occurrences": "eval: value == 1 || ':::action:::' == 'resolve'"
//       }
//     }
//   }
// }
//
// example #2:
//
// {
//   "filters": {
//     "occurrences": {
//       "negate": true,
//       "attributes": {
//         "occurrences": "eval: value > :::check.occurrences|60:::"
//       }
//     }
//   }
// }
//
// NOTE: Any of these fields should cause failure:
//
// NOTE: Fields that are currently ignored
//
// NOTE: Valid keys
//
func (i *LegacyFilterImporter) applyCfg(filter *types.EventFilter, cfg map[string]interface{}) {
	// Determine the action
	filter.Action = types.EventFilterActionAllow
	if negate, ok := cfg["negate"].(bool); ok && negate {
		filter.Action = types.EventFilterActionDeny
	}

	// Determine the statements
	i.getStatements(filter, cfg["attributes"], "event")
}

func (i *LegacyFilterImporter) getStatements(filter *types.EventFilter, attributes interface{}, keyPrefix string) {
	// reporter := i.reporter.WithValue("name", filter.Name)

	if attributes, ok := attributes.(map[string]interface{}); ok {
		for key, value := range attributes {
			if key == "check" {
				i.getStatements(filter, value, "event.Check")
			} else if key == "client" {
				i.getStatements(filter, value, "event.Entity")
			} else {
				// If we have a string, we should put it within single quotes
				if reflect.TypeOf(value).Kind() == reflect.String {
					// Does the filters contains attribute evaluation using Ruby expressions?
					if strings.Contains(value.(string), "eval:") {
						i.reporter.Error("attribute evaluation in filters is not supported at this time")
						continue
					}

					value = fmt.Sprintf("'%s'", value)
				}

				// Add this statement to the list of statements
				filter.Statements = append(
					filter.Statements,
					fmt.Sprintf("%s.%s == %v", keyPrefix, strings.Title(key), value),
				)
			}
		}
	}
}

//
// Handlers
//

// LegacyHandlerImporter provides utility to import Sensu v1 handler definitiions
type LegacyHandlerImporter struct {
	Org      string
	Env      string
	SaveFunc func(*types.Handler) error

	reporter *report.Writer
	handlers []*types.Handler
}

// Name of the importer
func (i *LegacyHandlerImporter) Name() string {
	return "handlers"
}

// SetReporter ...
func (i *LegacyHandlerImporter) SetReporter(r *report.Writer) {
	reporter := r.WithValue("component", i.Name())
	i.reporter = &reporter
}

// Import given data
func (i *LegacyHandlerImporter) Import(data map[string]interface{}) error {
	if vals, ok := data["handlers"].(map[string]interface{}); ok {
		for name, cfg := range vals {
			handler := i.newHandler(name)
			i.applyCfg(&handler, cfg.(map[string]interface{}))

			handler.Name = name
			handler.Organization = i.Org
			handler.Environment = i.Env
			i.handlers = append(i.handlers, &handler)
		}
	} else if _, ok = data["handlers"]; ok {
		i.reporter.Warn("handlers present but do not appear to be a hash; please handler format")
	}

	return nil
}

// Validate the given handlers
func (i *LegacyHandlerImporter) Validate() error {
	for _, handler := range i.handlers {
		if err := handler.Validate(); err != nil {
			i.reporter.Error("handler '" + handler.Name + "' failed validation w/ '" + err.Error() + "'")
		}
	}

	return nil
}

// Save calls API with handlers
func (i *LegacyHandlerImporter) Save() (int, error) {
	for _, handler := range i.handlers {
		if err := i.SaveFunc(handler); err != nil {
			i.reporter.
				WithValue("name", handler.Name).
				WithValue("error", err).
				Fatalf(
					"unable to persist handler '%s' w/ error '%s'",
					handler.Name,
					err,
				)
		} else {
			i.reporter.Debugf("handler '%s' imported", handler.Name)
		}
	}

	return len(i.handlers), nil
}

func (i *LegacyHandlerImporter) newHandler(name string) types.Handler {
	return types.Handler{
		Name:         name,
		Organization: i.Org,
		Environment:  i.Env,
	}
}

//
// example #1:
//
// {
//   "handlers": {
//     "file": {
//       "type": "pipe",
//       "command": "/etc/sensu/plugins/event-file.rb",
//       "timeout": 10,
//       "severities": ["critical", "unknown"]
//     }
//   }
// }
//
// example #2:
//
// {
//   "handlers": {
//     "example_udp_handler": {
//       "type": "udp",
//       "socket": {
//         "host": "10.0.1.99",
//         "port": 4444
//       }
//     }
//   }
// }
//
// NOTE: Any of these fields should cause failure:
//
//   - `filter`
//   - `filters`
//   - `severities`
//   - `pipe`
//
// NOTE: Fields that are currently ignored
//
//   - `handle_silenced`
//   - `handle_flapped`
//
// NOTE: Valid keys
//
//   - `type`
//   - `timeout`
//   - `mutator`
//   - `command`
//   - `socket`
//   - `handlers`
//
func (i *LegacyHandlerImporter) applyCfg(handler *types.Handler, cfg map[string]interface{}) {
	reporter := i.reporter.WithValue("name", handler.Name)

	//
	// Capture critical unsupported attributes and fail

	// "filter": "..."
	if _, ok := cfg["filter"]; ok {
		reporter.Error(unsupportedAttr("handlers", "filter"))
	}

	// "filters": ["...", "..."],
	if _, ok := cfg["filters"]; ok {
		reporter.Error(unsupportedAttr("handlers", "filters"))
	}

	// "severities": ["...", "..."],
	if _, ok := cfg["severities"]; ok {
		reporter.Error(unsupportedAttr("handlers", "severities"))
	}

	// "pipe": {}
	if _, ok := cfg["pipe"]; ok {
		reporter.Error(unsupportedAttr("handlers", "pipe"))
	}

	//
	// Capture unsupported attributes and warn user

	// "type": string
	if val, ok := cfg["type"].(string); ok && val == "set" {
		reporter.Info("handler sets will not work at this time")
	} else if val, ok := cfg["type"].(string); ok && (val == "udp" || val == "tcp") {
		reporter.Info("socket handlers will not work at this time")
	} else if val, ok := cfg["type"].(string); ok && val == "transport" {
		reporter.Info("transport handlers will not work at this time")
	}

	// "handle_silenced": bool
	if _, ok := cfg["handle_silenced"]; ok {
		reporter.Warn(unsupportedAttr("handlers", "handle_silenced"))
	}

	// "handle_flapped": bool
	if _, ok := cfg["handle_flapped"]; ok {
		reporter.Warn(unsupportedAttr("handlers", "handle_flapped"))
	}

	//
	// Apply values

	if val, ok := cfg["type"].(string); ok {
		handler.Type = val
	}

	if val, ok := cfg["mutator"].(string); ok {
		handler.Mutator = val
	}

	if val, ok := cfg["timeout"].(float64); ok {
		handler.Timeout = uint32(val)
	} else {
		handler.Timeout = 10
	}

	if val, ok := cfg["command"].(string); ok {
		handler.Command = val
	}

	if val, ok := cfg["socket"].(map[string]interface{}); ok {
		handler.Socket = &types.HandlerSocket{
			Host: val["host"].(string),
			// https://github.com/sensu/sensu-go/pull/494#discussion_r147844977
			Port: uint32(float32(val["port"].(float64))),
		}
	} else if _, ok := cfg["socket"]; ok {
		reporter.Error("handler's 'socket' attribute does not appear to be a JSON object")
	}

	if val, ok := cfg["handlers"].([]string); ok {
		handler.Handlers = val
	}
}

//
// Entities
//

// LegacyEntityImporter provides utility to import Sensu v1 entity definitiions
type LegacyEntityImporter struct {
	Org      string
	Env      string
	SaveFunc func(*types.Entity) error

	reporter *report.Writer
}

// Name of the importer
func (i *LegacyEntityImporter) Name() string {
	return "clients"
}

// SetReporter ...
func (i *LegacyEntityImporter) SetReporter(r *report.Writer) {
	reporter := r.WithValue("component", i.Name())
	i.reporter = &reporter
}

// Import given data
func (i *LegacyEntityImporter) Import(data map[string]interface{}) error {
	if _, ok := data["client"]; ok {
		i.reporter.Info("Sensu v1 clients cannot be imported; given 'client' settings will be ignored. Sensu v2 agents are configured directly through CLI arguments.")
	}

	return nil
}

// Validate the given entitys
func (i *LegacyEntityImporter) Validate() error {
	return nil
}

// Save calls API with entitys
func (i *LegacyEntityImporter) Save() (int, error) {
	return 0, nil
}

// NewEntity creates a new Entity with the importer's Organization and
// Environment.
func (i *LegacyEntityImporter) NewEntity(name string) types.Entity {
	return types.Entity{
		ID:           name,
		Organization: i.Org,
		Environment:  i.Env,
	}
}

// ApplyCfg applies a configuration to an entity.
// example #1:
//
// {
//   "client": {
//     "name": "i-424242",
//     "address": "8.8.8.8",
//     "subscriptions": [
//       "production",
//       "webserver",
//       "mysql"
//     ],
//     "socket": {
//       "bind": "127.0.0.1",
//       "port": 3030
//     }
//   }
// }
//
// example #2:
//
// {
//   "client": {
//     "name": "i-424242",
//     "address": "8.8.8.8",
//     "subscriptions": [
//       "production",
//       "webserver",
//       "roundrobin:webserver"
//     ]
//   }
// }
//
// NOTE: Any of these fields should cause failure:
//
// NOTE: Fields that are currently ignored
//
// NOTE: Valid keys
//
func (i *LegacyEntityImporter) ApplyCfg(entity *types.Entity, cfg map[string]interface{}) {
	// TODO
}

//
// Extensions
//

// LegacyExtensionImporter provides utility to import Sensu v1 extension definitiions
type LegacyExtensionImporter struct {
	Org string
	Env string

	reporter *report.Writer
}

// Name of the importer
func (i *LegacyExtensionImporter) Name() string {
	return "extensions"
}

// SetReporter ...
func (i *LegacyExtensionImporter) SetReporter(r *report.Writer) {
	reporter := r.WithValue("component", i.Name())
	i.reporter = &reporter
}

// Import given data
func (i *LegacyExtensionImporter) Import(data map[string]interface{}) error {
	if _, ok := data["extensions"]; ok {
		i.reporter.Warn("Sensu v1 extensions cannot be emulated in v2 at this time.")
	}

	return nil
}

// Validate the given extensions
func (i *LegacyExtensionImporter) Validate() error {
	return nil
}

// Save calls API with extensions
func (i *LegacyExtensionImporter) Save() (int, error) {
	return 0, nil
}

// ApplyCfg applies a configuration to an entity.
// example #1:
//
// {
//   "extensions": {
//     "state_change_only": {
//       "negate": false,
//       "attributes": {
//         "occurrences": "eval: value == 1 || ':::action:::' == 'resolve'"
//       }
//     }
//   }
// }
//
// example #2:
//
// {
//   "extensions": {
//     "occurrences": {
//       "negate": true,
//       "attributes": {
//         "occurrences": "eval: value > :::check.occurrences|60:::"
//       }
//     }
//   }
// }
//
// NOTE: Any of these fields should cause failure:
//
// NOTE: Fields that are currently ignored
//
// NOTE: Valid keys
//
func (i *LegacyExtensionImporter) ApplyCfg(_ *types.Entity, cfg map[string]interface{}) {
	// TODO
}

//
// Unsupported Settings
//

var (
	// LegacyTransportImporter logs message if transport configuration is present
	LegacyTransportImporter = UnsupportedLegacyResource{
		Key:          "transport",
		ResourceName: "transport",
		Message:      "Sensu v2 no longer requires a separate transport mechanism",
	}

	// LegacyAPIImporter logs message if api configuration is present
	LegacyAPIImporter = UnsupportedLegacyResource{
		Key:          "api",
		ResourceName: "Sensu API",
	}

	// LegacySensuImporter logs message if api configuration is present
	LegacySensuImporter = UnsupportedLegacyResource{
		Key:          "sensu",
		ResourceName: "Sensu server",
	}
)

//
// Misc.
//

// UnsupportedLegacyResource logs a message when the given key is found
type UnsupportedLegacyResource struct {
	Key          string
	ResourceName string
	Message      string

	reporter *report.Writer
}

// Name of the importer
func (i *UnsupportedLegacyResource) Name() string {
	return i.ResourceName
}

// SetReporter ...
func (i *UnsupportedLegacyResource) SetReporter(r *report.Writer) {
	reporter := r.WithValue("component", i.Name())
	i.reporter = &reporter
}

// Import given data
func (i *UnsupportedLegacyResource) Import(data map[string]interface{}) error {
	if _, ok := data[i.Key]; ok && i.Message == "" {
		i.reporter.Infof("Sensu v2 no longer requires a separate '%s' configuration", i.Key)
	} else if ok {
		i.reporter.Info(i.Message)
	}

	return nil
}

// Validate the given transports
func (i *UnsupportedLegacyResource) Validate() error {
	return nil
}

// Save calls API with transports
func (i *UnsupportedLegacyResource) Save() (int, error) {
	return 0, nil
}
