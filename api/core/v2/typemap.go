package v2

// automatically generated file, do not edit!

import (
	"fmt"
	"reflect"
)

// typeMap is used to dynamically look up data types from strings.
var typeMap = map[string]interface{}{
	"APIKey":                 &APIKey{},
	"api_key":                &APIKey{},
	"AdhocRequest":           &AdhocRequest{},
	"adhoc_request":          &AdhocRequest{},
	"Any":                    &Any{},
	"any":                    &Any{},
	"Asset":                  &Asset{},
	"asset":                  &Asset{},
	"AssetBuild":             &AssetBuild{},
	"asset_build":            &AssetBuild{},
	"AssetList":              &AssetList{},
	"asset_list":             &AssetList{},
	"AuthProviderClaims":     &AuthProviderClaims{},
	"auth_provider_claims":   &AuthProviderClaims{},
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
	"ClusterHealth":          &ClusterHealth{},
	"cluster_health":         &ClusterHealth{},
	"ClusterRole":            &ClusterRole{},
	"cluster_role":           &ClusterRole{},
	"ClusterRoleBinding":     &ClusterRoleBinding{},
	"cluster_role_binding":   &ClusterRoleBinding{},
	"Deregistration":         &Deregistration{},
	"deregistration":         &Deregistration{},
	"Entity":                 &Entity{},
	"entity":                 &Entity{},
	"Event":                  &Event{},
	"event":                  &Event{},
	"EventFilter":            &EventFilter{},
	"event_filter":           &EventFilter{},
	"Handler":                &Handler{},
	"handler":                &Handler{},
	"HandlerSocket":          &HandlerSocket{},
	"handler_socket":         &HandlerSocket{},
	"HealthResponse":         &HealthResponse{},
	"health_response":        &HealthResponse{},
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
	"Namespace":              &Namespace{},
	"namespace":              &Namespace{},
	"Network":                &Network{},
	"network":                &Network{},
	"NetworkInterface":       &NetworkInterface{},
	"network_interface":      &NetworkInterface{},
	"ObjectMeta":             &ObjectMeta{},
	"object_meta":            &ObjectMeta{},
	"Pipeline":               &Pipeline{},
	"pipeline":               &Pipeline{},
	"PipelineWorkflow":       &PipelineWorkflow{},
	"pipeline_workflow":      &PipelineWorkflow{},
	"PostgresHealth":         &PostgresHealth{},
	"postgres_health":        &PostgresHealth{},
	"Process":                &Process{},
	"process":                &Process{},
	"ProxyRequests":          &ProxyRequests{},
	"proxy_requests":         &ProxyRequests{},
	"ResourceReference":      &ResourceReference{},
	"resource_reference":     &ResourceReference{},
	"Role":                   &Role{},
	"role":                   &Role{},
	"RoleBinding":            &RoleBinding{},
	"role_binding":           &RoleBinding{},
	"RoleRef":                &RoleRef{},
	"role_ref":               &RoleRef{},
	"Rule":                   &Rule{},
	"rule":                   &Rule{},
	"Secret":                 &Secret{},
	"secret":                 &Secret{},
	"Silenced":               &Silenced{},
	"silenced":               &Silenced{},
	"Subject":                &Subject{},
	"subject":                &Subject{},
	"System":                 &System{},
	"system":                 &System{},
	"TLSOptions":             &TLSOptions{},
	"tls_options":            &TLSOptions{},
	"TessenConfig":           &TessenConfig{},
	"tessen_config":          &TessenConfig{},
	"TimeWindowDays":         &TimeWindowDays{},
	"time_window_days":       &TimeWindowDays{},
	"TimeWindowTimeRange":    &TimeWindowTimeRange{},
	"time_window_time_range": &TimeWindowTimeRange{},
	"TimeWindowWhen":         &TimeWindowWhen{},
	"time_window_when":       &TimeWindowWhen{},
	"Tokens":                 &Tokens{},
	"tokens":                 &Tokens{},
	"TypeMeta":               &TypeMeta{},
	"type_meta":              &TypeMeta{},
	"User":                   &User{},
	"user":                   &User{},
	"Version":                &Version{},
	"version":                &Version{},
}

// ResolveResource returns a zero-valued resource, given a name.
// If the named type does not exist, or if the type is not a Resource,
// then an error will be returned.
func ResolveResource(name string) (Resource, error) {
	t, ok := typeMap[name]
	if !ok {
		return nil, fmt.Errorf("type could not be found: %q", name)
	}
	if _, ok := t.(Resource); !ok {
		return nil, fmt.Errorf("%q is not a Resource", name)
	}
	return newResource(t), nil
}

// Make a new Resource to avoid aliasing problems with ResolveResource.
// don't use this function. no, seriously.
func newResource(r interface{}) Resource {
	return reflect.New(reflect.ValueOf(r).Elem().Type()).Interface().(Resource)
}
