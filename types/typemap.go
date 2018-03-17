// automatically generated file, do not edit!
package types

import "errors"

var ErrNoType = errors.New("the named type could not be found")
var ErrInvalidResource = errors.New("the named type is not a Resource")

// typeMap is used to dynamically look up data types from strings.
var typeMap = map[string]interface{}{
	"AdhocRequest":        &AdhocRequest{},
	"Any":                 &Any{},
	"Asset":               &Asset{},
	"Check":               &Check{},
	"CheckConfig":         &CheckConfig{},
	"CheckHistory":        &CheckHistory{},
	"CheckRequest":        &CheckRequest{},
	"Claims":              &Claims{},
	"Deregistration":      &Deregistration{},
	"Entity":              &Entity{},
	"Environment":         &Environment{},
	"Error":               &Error{},
	"Event":               &Event{},
	"EventFilter":         &EventFilter{},
	"Handler":             &Handler{},
	"HandlerSocket":       &HandlerSocket{},
	"Hook":                &Hook{},
	"HookConfig":          &HookConfig{},
	"HookList":            &HookList{},
	"KeepaliveRecord":     &KeepaliveRecord{},
	"MetricPoint":         &MetricPoint{},
	"MetricTag":           &MetricTag{},
	"Metrics":             &Metrics{},
	"Mutator":             &Mutator{},
	"Network":             &Network{},
	"NetworkInterface":    &NetworkInterface{},
	"Organization":        &Organization{},
	"ProxyRequests":       &ProxyRequests{},
	"Role":                &Role{},
	"Rule":                &Rule{},
	"Silenced":            &Silenced{},
	"System":              &System{},
	"TLSOptions":          &TLSOptions{},
	"TimeWindowDays":      &TimeWindowDays{},
	"TimeWindowTimeRange": &TimeWindowTimeRange{},
	"TimeWindowWhen":      &TimeWindowWhen{},
	"Tokens":              &Tokens{},
	"User":                &User{},
	"Wrapper":             &Wrapper{},
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
