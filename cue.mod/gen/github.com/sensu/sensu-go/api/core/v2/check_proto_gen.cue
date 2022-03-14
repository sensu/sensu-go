package v2

// A CheckRequest represents a request to execute a check
#CheckRequest: {
	// Config is the specification of a check.
	config?: #CheckConfig @protobuf(1,CheckConfig,"(gogoproto.nullable)")

	// Assets are a list of assets required to execute check.
	assets?: [...#Asset] @protobuf(2,Asset,"(gogoproto.nullable)=false")

	// Hooks are a list of hooks to be executed after a check.
	hooks?: [...#HookConfig] @protobuf(3,HookConfig,"(gogoproto.nullable)=false")

	// Issued describes the time in which the check request was issued
	issued?: int64 @protobuf(4,int64,#"(gogoproto.jsontag)="issued""#,name=Issued)

	// HookAssets is a map of assets required to execute hooks.
	hookAssets?: {
		[string]: #AssetList
	} @protobuf(5,map[string]AssetList,hook_assets,#"(gogoproto.jsontag)="hook_assets""#)

	// Secrets is a list of kv to be added to the env vars of a check.
	secrets?: [...string] @protobuf(6,string)
}

// An AssetList represents a list of assets for a CheckRequest.
#AssetList: {
	// Assets are a list of assets required to execute check or hook.
	assets?: [...#Asset] @protobuf(1,Asset,#"(gogoproto.jsontag)="assets""#,"(gogoproto.nullable)=false")
}

// A ProxyRequests represents a request to execute a proxy check
#ProxyRequests: {
	// EntityAttributes store serialized arbitrary JSON-encoded data to match
	// entities in the registry.
	entity_attributes?: [...string] @protobuf(1,string,#"(gogoproto.jsontag)="entity_attributes""#)

	// Splay indicates if proxy check requests should be splayed, published
	// evenly over a window of time.
	splay?: bool @protobuf(2,bool,#"(gogoproto.jsontag)="splay""#)

	// SplayCoverage is the percentage used for proxy check request splay
	// calculation.
	splay_coverage?: uint32 @protobuf(3,uint32,#"(gogoproto.jsontag)="splay_coverage""#)
}

// CheckConfig is the specification of a check.
#CheckConfig: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)

	// Command is the command to be executed.
	command?: string @protobuf(1,string)

	// Handlers are the event handler for the check (incidents and/or metrics).
	handlers?: [...string] @protobuf(3,string,#"(gogoproto.jsontag)="handlers""#)

	// HighFlapThreshold is the flap detection high threshold (% state change)
	// for the check. Sensu uses the same flap detection algorithm as Nagios.
	high_flap_threshold?: uint32 @protobuf(4,uint32,#"(gogoproto.jsontag)="high_flap_threshold""#)

	// Interval is the interval, in seconds, at which the check should be run.
	interval?: uint32 @protobuf(5,uint32,#"(gogoproto.jsontag)="interval""#)

	// LowFlapThreshold is the flap detection low threshold (% state change) for
	// the check. Sensu uses the same flap detection algorithm as Nagios.
	low_flap_threshold?: uint32 @protobuf(6,uint32,#"(gogoproto.jsontag)="low_flap_threshold""#)

	// Publish indicates if check requests are published for the check
	publish?: bool @protobuf(9,bool,#"(gogoproto.jsontag)="publish""#)

	// RuntimeAssets are a list of assets required to execute check.
	runtime_assets?: [...string] @protobuf(10,string,#"(gogoproto.jsontag)="runtime_assets""#)

	// Subscriptions is the list of subscribers for the check.
	subscriptions?: [...string] @protobuf(11,string,#"(gogoproto.jsontag)="subscriptions""#)

	// ExtendedAttributes store serialized arbitrary JSON-encoded data
	ExtendedAttributes?: bytes @protobuf(12,bytes,#"(gogoproto.jsontag)="-""#)

	// Sources indicates the name of the entity representing an external
	// resource
	proxy_entity_name?: string @protobuf(13,string,#"(gogoproto.jsontag)="proxy_entity_name""#,#"(gogoproto.customname)="ProxyEntityName""#)

	// CheckHooks is the list of check hooks for the check
	check_hooks?: [...#HookList] @protobuf(14,HookList,#"(gogoproto.jsontag)="check_hooks""#,"(gogoproto.nullable)=false")

	// STDIN indicates if the check command accepts JSON via stdin from the
	// agent
	stdin?: bool @protobuf(15,bool,#"(gogoproto.jsontag)="stdin""#)

	// Subdue represents one or more time windows when the check should be
	// subdued.
	subdue?: #TimeWindowWhen @protobuf(16,TimeWindowWhen,#"(gogoproto.jsontag)="subdue""#)

	// Cron is the cron string at which the check should be run.
	cron?: string @protobuf(17,string)

	// TTL represents the length of time in seconds for which a check result is
	// valid.
	ttl?: int64 @protobuf(18,int64,#"(gogoproto.jsontag)="ttl""#)

	// Timeout is the timeout, in seconds, at which the check has to run
	timeout?: uint32 @protobuf(19,uint32,#"(gogoproto.jsontag)="timeout""#)

	// ProxyRequests represents a request to execute a proxy check
	proxyRequests?: #ProxyRequests @protobuf(20,ProxyRequests,name=proxy_requests)

	// RoundRobin enables round-robin scheduling if set true.
	round_robin?: bool @protobuf(21,bool,#"(gogoproto.jsontag)="round_robin""#)

	// OutputOutputMetricFormat is the metric protocol that the check's output
	// will be expected to follow in order to be extracted.
	output_metric_format?: string @protobuf(22,string,#"(gogoproto.jsontag)="output_metric_format""#)

	// OutputOutputMetricHandlers is the list of event handlers that will
	// respond to metrics that have been extracted from the check.
	output_metric_handlers?: [...string] @protobuf(23,string,#"(gogoproto.jsontag)="output_metric_handlers""#)

	// EnvVars is the list of environment variables to set for the check's
	// execution environment.
	env_vars?: [...string] @protobuf(24,string,#"(gogoproto.jsontag)="env_vars""#)

	// Metadata contains the name, namespace, labels and annotations of the
	// check
	metadata?: #ObjectMeta @protobuf(26,ObjectMeta,#"(gogoproto.jsontag)="metadata,omitempty""#,"(gogoproto.embed)","(gogoproto.nullable)=false")

	// MaxOutputSize is the maximum size in bytes that will be stored for check
	// output. If check output is larger than MaxOutputSize, it will be
	// truncated when stored. Filters, mutators, and handlers will still have
	// access to the full check output at the time the event occurs.
	maxOutputSize?: int64 @protobuf(27,int64,name=max_output_size)

	// DiscardOutput causes agents to discard check output. No check output is
	// written to the backend, but metrics extraction is still performed.
	discardOutput?: bool @protobuf(28,bool,name=discard_output)

	// Secrets is the list of Sensu secrets to set for the check's
	// execution environment.
	secrets?: [...#Secret] @protobuf(29,Secret,#"(gogoproto.jsontag)="secrets""#)

	// OutputMetricTags is list of metric tags to apply to metrics extracted from check output.
	output_metric_tags?: [...#MetricTag] @protobuf(30,MetricTag,#"(gogoproto.jsontag)="output_metric_tags,omitempty""#,#"(gogoproto.moretags)="yaml: \"output_metric_tags,omitempty\"""#)

	// Scheduler is the type of scheduler the check is scheduled by. The scheduler
	// can be "memory", "etcd", or "postgres". Scheduler is set by Sensu - any
	// setting by the user will be overridden.
	scheduler?: string @protobuf(31,string,#"(gogoproto.jsontag)="-""#,#"(gogoproto.moretags)="yaml: \"-\"""#)

	// Pipelines are the pipelines this check will use to process its events.
	pipelines?: [...#ResourceReference] @protobuf(32,ResourceReference,#"(gogoproto.jsontag)="pipelines""#)
	output_metric_thresholds?: [...#MetricThreshold] @protobuf(33,MetricThreshold,#"(gogoproto.jsontag)="output_metric_thresholds,omitempty""#,#"(gogoproto.moretags)="yaml: \"output_metric_thresholds,omitempty\"""#)
}

// A Check is a check specification and optionally the results of the check's
// execution.
#Check: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)

	// Command is the command to be executed.
	command?: string @protobuf(1,string)

	// Handlers are the event handler for the check (incidents and/or metrics).
	handlers?: [...string] @protobuf(3,string,#"(gogoproto.jsontag)="handlers""#)

	// HighFlapThreshold is the flap detection high threshold (% state change)
	// for the check. Sensu uses the same flap detection algorithm as Nagios.
	high_flap_threshold?: uint32 @protobuf(4,uint32,#"(gogoproto.jsontag)="high_flap_threshold""#)

	// Interval is the interval, in seconds, at which the check should be run.
	interval?: uint32 @protobuf(5,uint32,#"(gogoproto.jsontag)="interval""#)

	// LowFlapThreshold is the flap detection low threshold (% state change) for
	// the check. Sensu uses the same flap detection algorithm as Nagios.
	low_flap_threshold?: uint32 @protobuf(6,uint32,#"(gogoproto.jsontag)="low_flap_threshold""#)

	// Publish indicates if check requests are published for the check
	publish?: bool @protobuf(9,bool,#"(gogoproto.jsontag)="publish""#)

	// RuntimeAssets are a list of assets required to execute check.
	runtime_assets?: [...string] @protobuf(10,string,#"(gogoproto.jsontag)="runtime_assets""#)

	// Subscriptions is the list of subscribers for the check.
	subscriptions?: [...string] @protobuf(11,string,#"(gogoproto.jsontag)="subscriptions""#)

	// Sources indicates the name of the entity representing an external
	// resource
	proxy_entity_name?: string @protobuf(13,string,#"(gogoproto.jsontag)="proxy_entity_name""#,#"(gogoproto.customname)="ProxyEntityName""#)

	// CheckHooks is the list of check hooks for the check
	check_hooks?: [...#HookList] @protobuf(14,HookList,#"(gogoproto.jsontag)="check_hooks""#,"(gogoproto.nullable)=false")

	// STDIN indicates if the check command accepts JSON via stdin from the
	// agent
	stdin?: bool @protobuf(15,bool,#"(gogoproto.jsontag)="stdin""#)

	// Subdue represents one or more time windows when the check should be
	// subdued.
	subdue?: #TimeWindowWhen @protobuf(16,TimeWindowWhen,#"(gogoproto.jsontag)="subdue""#)

	// Cron is the cron string at which the check should be run.
	cron?: string @protobuf(17,string)

	// TTL represents the length of time in seconds for which a check result is
	// valid.
	ttl?: int64 @protobuf(18,int64,#"(gogoproto.jsontag)="ttl""#)

	// Timeout is the timeout, in seconds, at which the check has to run
	timeout?: uint32 @protobuf(19,uint32,#"(gogoproto.jsontag)="timeout""#)

	// ProxyRequests represents a request to execute a proxy check
	proxyRequests?: #ProxyRequests @protobuf(20,ProxyRequests,name=proxy_requests)

	// RoundRobin enables round-robin scheduling if set true.
	round_robin?: bool @protobuf(21,bool,#"(gogoproto.jsontag)="round_robin""#)

	// Duration of execution
	duration?: float64 @protobuf(22,double)

	// Executed describes the time in which the check request was executed
	executed?: int64 @protobuf(23,int64,#"(gogoproto.jsontag)="executed""#)

	// History is the check state history.
	history?: [...#CheckHistory] @protobuf(24,CheckHistory,#"(gogoproto.jsontag)="history""#,"(gogoproto.nullable)=false")

	// Issued describes the time in which the check request was issued
	issued?: int64 @protobuf(25,int64,#"(gogoproto.jsontag)="issued""#)

	// Output from the execution of Command
	output?: string @protobuf(26,string,#"(gogoproto.jsontag)="output""#)

	// State provides handlers with more information about the state change
	state?: string @protobuf(27,string)

	// Status is the exit status code produced by the check
	status?: uint32 @protobuf(28,uint32,#"(gogoproto.jsontag)="status""#)

	// TotalStateChange indicates the total state change percentage for the
	// check's history
	total_state_change?: uint32 @protobuf(29,uint32,#"(gogoproto.jsontag)="total_state_change""#)

	// LastOK displays last time this check was ok; if event status is 0 this is
	// set to timestamp
	last_ok?: int64 @protobuf(30,int64,#"(gogoproto.customname)="LastOK""#,#"(gogoproto.jsontag)="last_ok""#)

	// Occurrences indicates the number of times an event has occurred for a
	// client/check pair with the same check status
	occurrences?: int64 @protobuf(31,int64,#"(gogoproto.jsontag)="occurrences""#)

	// OccurrencesWatermark indicates the high water mark tracking number of
	// occurrences at the current severity
	occurrences_watermark?: int64 @protobuf(32,int64,#"(gogoproto.jsontag)="occurrences_watermark""#)

	// Silenced is a list of silenced entry ids (subscription and check name)
	silenced?: [...string] @protobuf(33,string,"(gogoproto.nullable)")

	// Hooks describes the results of multiple hooks; if event is associated to
	// hook execution.
	hooks?: [...#Hook] @protobuf(34,Hook,"(gogoproto.nullable)")

	// OutputMetricFormat is the metric protocol that the check's output
	// will be expected to follow in order to be extracted.
	output_metric_format?: string @protobuf(35,string,#"(gogoproto.jsontag)="output_metric_format""#)

	// OutputMetricHandlers is the list of event handlers that will
	// respond to metrics that have been extracted from the check.
	output_metric_handlers?: [...string] @protobuf(36,string,#"(gogoproto.jsontag)="output_metric_handlers""#)

	// EnvVars is the list of environment variables to set for the check's
	// execution environment.
	env_vars?: [...string] @protobuf(37,string,#"(gogoproto.jsontag)="env_vars""#)

	// Metadata contains the name, namespace, labels and annotations of the
	// check
	metadata?: #ObjectMeta @protobuf(38,ObjectMeta,#"(gogoproto.jsontag)="metadata,omitempty""#,"(gogoproto.embed)","(gogoproto.nullable)=false")

	// MaxOutputSize is the maximum size in bytes that will be stored for check
	// output. If check output is larger than MaxOutputSize, it will be
	// truncated when stored. Filters, mutators, and handlers will still have
	// access to the full check output at the time the event occurs.
	maxOutputSize?: int64 @protobuf(39,int64,name=max_output_size)

	// DiscardOutput causes agents to discard check output. No check output is
	// written to the backend, but metrics extraction is still performed.
	discardOutput?: bool @protobuf(40,bool,name=discard_output)

	// Secrets is the list of Sensu secrets to set for the check's
	// execution environment.
	secrets?: [...#Secret] @protobuf(41,Secret,#"(gogoproto.jsontag)="secrets""#)

	// IsSilenced indicates whether the check is silenced or not
	is_silenced?: bool @protobuf(42,bool,#"(gogoproto.jsontag)="is_silenced""#)

	// OutputMetricTags is list of metric tags to apply to metrics extracted from check output.
	output_metric_tags?: [...#MetricTag] @protobuf(43,MetricTag,#"(gogoproto.jsontag)="output_metric_tags,omitempty""#,#"(gogoproto.moretags)="yaml: \"output_metric_tags,omitempty\"""#)

	// Scheduler is the type of scheduler the check is scheduled by. The scheduler
	// can be "memory", "etcd", or "postgres". Scheduler is set by Sensu - any
	// setting by the user will be overridden.
	scheduler?: string @protobuf(44,string,#"(gogoproto.jsontag)="scheduler""#)

	// ProcessedBy indicates the name of the agent that processed the event. This
	// is mainly useful for determining which agent executed a proxy check request.
	processed_by?: string @protobuf(45,string,#"(gogoproto.jsontag)="processed_by,omitempty""#,#"(gogoproto.moretags)="yaml: \"processed_by\"""#,name=ProcessedBy)

	// Pipelines are the pipelines this check will use to process its events.
	pipelines?: [...#ResourceReference] @protobuf(46,ResourceReference,#"(gogoproto.jsontag)="pipelines""#)

	// MetricThresholds are a list of thresholds to apply to metrics in order to determine
	// the check status.
	output_metric_thresholds?: [...#MetricThreshold] @protobuf(47,MetricThreshold,#"(gogoproto.jsontag)="output_metric_thresholds,omitempty""#,#"(gogoproto.moretags)="yaml: \"output_metric_thresholds,omitempty\"""#)

	// ExtendedAttributes store serialized arbitrary JSON-encoded data
	ExtendedAttributes?: bytes @protobuf(99,bytes,#"(gogoproto.jsontag)="-""#)
}

// CheckHistory is a record of a check execution and its status
#CheckHistory: {
	// Status is the exit status code produced by the check.
	status?: uint32 @protobuf(1,uint32,#"(gogoproto.jsontag)="status""#)

	// Executed describes the time in which the check request was executed
	executed?: int64 @protobuf(2,int64,#"(gogoproto.jsontag)="executed""#)

	// Flapping describes whether the check was flapping at this particular
	// point in time. Comparing this value to the current flapping status allows
	// filters to trigger only on start and end of flapping. NB! This has been
	// disabled for 5.x releases.
	flapping?: bool @protobuf(3,bool,#"(gogoproto.jsontag)="-""#)
}
