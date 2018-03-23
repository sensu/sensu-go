package types

// automatically generated file, do not edit!

import "errors"

var ErrNoType = errors.New("the named type could not be found")
var ErrInvalidResource = errors.New("the named type is not a Resource")

// typeMap is used to dynamically look up data types from strings.
var typeMap = map[string]interface{}{
	"AdhocRequest":           &AdhocRequest{},
	"adhoc_request":          &AdhocRequest{},
	"Any":                    &Any{},
	"any":                    &Any{},
	"Asset":                  &Asset{},
	"asset":                  &Asset{},
	"Check":                  &Check{},
	"check":                  &Check{},
	"CheckConfig":            &CheckConfig{},
	"check_config":           &CheckConfig{},
	"CheckHistory":           &CheckHistory{},
	"check_history":          &CheckHistory{},
	"CheckRequest":           &CheckRequest{},
	"check_request":          &CheckRequest{},
	"Claims":                 &Claims{},
	"claims":                 &Claims{},
	"Deregistration":         &Deregistration{},
	"deregistration":         &Deregistration{},
	"Entity":                 &Entity{},
	"entity":                 &Entity{},
	"Environment":            &Environment{},
	"environment":            &Environment{},
	"Error":                  &Error{},
	"error":                  &Error{},
	"Event":                  &Event{},
	"event":                  &Event{},
	"EventFilter":            &EventFilter{},
	"event_filter":           &EventFilter{},
	"Handler":                &Handler{},
	"handler":                &Handler{},
	"HandlerSocket":          &HandlerSocket{},
	"handler_socket":         &HandlerSocket{},
	"Hook":                   &Hook{},
	"hook":                   &Hook{},
	"HookConfig":             &HookConfig{},
	"hook_config":            &HookConfig{},
	"HookList":               &HookList{},
	"hook_list":              &HookList{},
	"KeepaliveRecord":        &KeepaliveRecord{},
	"keepalive_record":       &KeepaliveRecord{},
	"MetricPoint":            &MetricPoint{},
	"metric_point":           &MetricPoint{},
	"MetricTag":              &MetricTag{},
	"metric_tag":             &MetricTag{},
	"Metrics":                &Metrics{},
	"metrics":                &Metrics{},
	"Mutator":                &Mutator{},
	"mutator":                &Mutator{},
	"Network":                &Network{},
	"network":                &Network{},
	"NetworkInterface":       &NetworkInterface{},
	"network_interface":      &NetworkInterface{},
	"Organization":           &Organization{},
	"organization":           &Organization{},
	"ProxyRequests":          &ProxyRequests{},
	"proxy_requests":         &ProxyRequests{},
	"Role":                   &Role{},
	"role":                   &Role{},
	"Rule":                   &Rule{},
	"rule":                   &Rule{},
	"Silenced":               &Silenced{},
	"silenced":               &Silenced{},
	"System":                 &System{},
	"system":                 &System{},
	"TLSOptions":             &TLSOptions{},
	"t_l_s_options":          &TLSOptions{},
	"TimeWindowDays":         &TimeWindowDays{},
	"time_window_days":       &TimeWindowDays{},
	"TimeWindowTimeRange":    &TimeWindowTimeRange{},
	"time_window_time_range": &TimeWindowTimeRange{},
	"TimeWindowWhen":         &TimeWindowWhen{},
	"time_window_when":       &TimeWindowWhen{},
	"Tokens":                 &Tokens{},
	"tokens":                 &Tokens{},
	"User":                   &User{},
	"user":                   &User{},
	"Wrapper":                &Wrapper{},
	"wrapper":                &Wrapper{},
}

// ResolveResource returns a zero-valued resource, given a name.
// If the named type does not exist, or if the type is not a Resource,
// then an error will be returned.
func ResolveResource(name string) (Resource, error) {
	t, ok := typeMap[name]
	if !ok {
		return nil, ErrNoType
	}
	r, ok := t.(Resource)
	if !ok {
		return nil, ErrInvalidResource
	}
	return r, nil
}
