package types

import "github.com/sensu/sensu-go/api/core/v2"

type (
	AdhocRequest        = v2.AdhocRequest
	Asset               = v2.Asset
	ByExecuted          = v2.ByExecuted
	Check               = v2.Check
	CheckConfig         = v2.CheckConfig
	CheckHistory        = v2.CheckHistory
	CheckRequest        = v2.CheckRequest
	Claims              = v2.Claims
	ClusterHealth       = v2.ClusterHealth
	ClusterRole         = v2.ClusterRole
	ClusterRoleBinding  = v2.ClusterRoleBinding
	Deregistration      = v2.Deregistration
	Entity              = v2.Entity
	Event               = v2.Event
	EventFilter         = v2.EventFilter
	Extension           = v2.Extension
	Handler             = v2.Handler
	HandlerSocket       = v2.HandlerSocket
	HealthResponse      = v2.HealthResponse
	Hook                = v2.Hook
	HookConfig          = v2.HookConfig
	HookList            = v2.HookList
	KeepaliveRecord     = v2.KeepaliveRecord
	MetricPoint         = v2.MetricPoint
	MetricTag           = v2.MetricTag
	Metrics             = v2.Metrics
	Mutator             = v2.Mutator
	Namespace           = v2.Namespace
	Network             = v2.Network
	NetworkInterface    = v2.NetworkInterface
	ObjectMeta          = v2.ObjectMeta
	ProxyRequests       = v2.ProxyRequests
	Resource            = v2.Resource
	Role                = v2.Role
	RoleBinding         = v2.RoleBinding
	RoleRef             = v2.RoleRef
	Rule                = v2.Rule
	Silenced            = v2.Silenced
	Subject             = v2.Subject
	System              = v2.System
	TLSOptions          = v2.TLSOptions
	TimeWindowDays      = v2.TimeWindowDays
	TimeWindowTimeRange = v2.TimeWindowTimeRange
	TimeWindowWhen      = v2.TimeWindowWhen
	Tokens              = v2.Tokens
	TypeMeta            = v2.TypeMeta
	User                = v2.User
)

type (
	ConstrainedResource = v2.ConstrainedResource
	MultitenantResource = v2.MultitenantResource
)

const (
	// AcceessTokenString is the key name used to retrieve the access token string
	AccessTokenString = v2.AccessTokenString

	// AccessTokenClaims contains the key name to retrieve the access token claims
	AccessTokenClaims = v2.AccessTokenClaims

	// ClaimsKey contains key name to retrieve the jwt claims from context
	ClaimsKey = v2.ClaimsKey

	// NamespaceKey contains the key name to retrieve the namespace from context
	NamespaceKey = v2.NamespaceKey

	// RefreshTokenClaims contains the key name to retrieve the refresh token claims
	RefreshTokenClaims = v2.RefreshTokenClaims

	// RefreshTokenString contains the key name to retrieve the refresh token string
	RefreshTokenString = v2.RefreshTokenString

	// AuthorizationAttributesKey is the key name used to store authorization
	// attributes extracted from a request = v2.//
	AuthorizationAttributesKey = v2.AuthorizationAttributesKey

	// StoreKey contains the key name to retrieve the etcd store from within a context = v2.//
	StoreKey = v2.StoreKey

	// ResourceAll represents all possible resources
	ResourceAll = v2.ResourceAll

	// VerbAll represents all possible verbs
	VerbAll = v2.VerbAll

	// GroupType represents a group object in a subject
	GroupType = v2.GroupType

	// UserType represents a user object in a subject
	UserType = v2.UserType

	// LocalSelfUserResource represents a local user trying to view itself
	// or change its password
	LocalSelfUserResource = v2.LocalSelfUserResource

	// HandlerPipeType represents handlers that pipes event data // into arbitrary
	// commands via STDIN
	HandlerPipeType = v2.HandlerPipeType

	// HandlerSetType represents handlers that groups event handlers, making it
	// easy to manage groups of actions that should be executed for certain v2
	// of events.
	HandlerSetType = v2.HandlerSetType

	// HandlerTCPType represents handlers that send event data to a remote TCP
	// socket
	HandlerTCPType = v2.HandlerTCPType

	// HandlerUDPType represents handlers that send event data to a remote UDP
	// socket
	HandlerUDPType = v2.HandlerUDPType

	// HandlerGRPCType is a special kind of handler that represents an extension
	HandlerGRPCType = v2.HandlerGRPCType

	// EventFilterActionAllow is an action to allow events to pass through to the pipeline
	EventFilterActionAllow = v2.EventFilterActionAllow

	// EventFilterActionDeny is an action to deny events from passing through to the pipeline
	EventFilterActionDeny = v2.EventFilterActionDeny

	// DefaultEventFilterAction is the default action for filters
	DefaultEventFilterAction = v2.DefaultEventFilterAction

	// EntityAgentClass is the name of the class given to agent entities.
	EntityAgentClass = v2.EntityAgentClass

	// EntityProxyClass is the name of the class given to proxy entities.
	EntityProxyClass = v2.EntityProxyClass

	// EntityBackendClass is the name of the class given to backend entities.
	EntityBackendClass = v2.EntityBackendClass

	// Redacted is filled in for fields that contain sensitive information
	Redacted = v2.Redacted

	EventFailingState = v2.EventFailingState

	// EventFlappingState indicates a rapid change in check result status
	EventFlappingState = v2.EventFlappingState

	// EventPassingState indicates successful check result status
	EventPassingState = v2.EventPassingState

	// CheckRequestType is the message type string for check request.
	CheckRequestType = v2.CheckRequestType

	// DefaultSplayCoverage is the default splay coverage for proxy check requests
	DefaultSplayCoverage = v2.DefaultSplayCoverage

	// NagiosOutputMetricFormat is the accepted string to represent the output metric format of
	// Nagios Perf Data
	NagiosOutputMetricFormat = v2.NagiosOutputMetricFormat

	// GraphiteOutputMetricFormat is the accepted string to represent the output metric format of
	// Graphite Plain Text
	GraphiteOutputMetricFormat = v2.GraphiteOutputMetricFormat

	// OpenTSDBOutputMetricFormat is the accepted string to represent the output metric format of
	// OpenTSDB Line
	OpenTSDBOutputMetricFormat = v2.OpenTSDBOutputMetricFormat

	// InfluxDBOutputMetricFormat is the accepted string to represent the output metric format of
	// InfluxDB Line
	InfluxDBOutputMetricFormat = v2.InfluxDBOutputMetricFormat

	// CoreEdition represents the Sensu Core Edition (CE)
	CoreEdition = v2.CoreEdition

	// EditionHeader represents the HTTP header containing the edition value
	EditionHeader = v2.EditionHeader

	// NamespaceTypeAll matches all actions
	NamespaceTypeAll = v2.NamespaceTypeAll

	DefaultKeepaliveTimeout = v2.DefaultKeepaliveTimeout
)

// Test fixture
var (
	FixtureCheckRequest       = v2.FixtureCheckRequest
	FixtureCheckConfig        = v2.FixtureCheckConfig
	FixtureCheck              = v2.FixtureCheck
	FixtureProxyRequests      = v2.FixtureProxyRequests
	FixtureNamespace          = v2.FixtureNamespace
	FixtureMetrics            = v2.FixtureMetrics
	FixtureMetricPoint        = v2.FixtureMetricPoint
	FixtureMetricTag          = v2.FixtureMetricTag
	FixtureHandler            = v2.FixtureHandler
	FixtureSocketHandler      = v2.FixtureSocketHandler
	FixtureSetHandler         = v2.FixtureSetHandler
	FixtureUser               = v2.FixtureUser
	FixtureHealthResponse     = v2.FixtureHealthResponse
	FixtureEvent              = v2.FixtureEvent
	FixtureEventFilter        = v2.FixtureEventFilter
	FixtureDenyEventFilter    = v2.FixtureDenyEventFilter
	FixtureExtension          = v2.FixtureExtension
	FixtureMutator            = v2.FixtureMutator
	FixtureAsset              = v2.FixtureAsset
	FixtureSubject            = v2.FixtureSubject
	FixtureRule               = v2.FixtureRule
	FixtureRole               = v2.FixtureRole
	FixtureRoleRef            = v2.FixtureRoleRef
	FixtureRoleBinding        = v2.FixtureRoleBinding
	FixtureClusterRole        = v2.FixtureClusterRole
	FixtureClusterRoleBinding = v2.FixtureClusterRoleBinding
	FixtureEntity             = v2.FixtureEntity
	FixtureHookConfig         = v2.FixtureHookConfig
	FixtureHook               = v2.FixtureHook
	FixtureHookList           = v2.FixtureHookList
	FixtureSilenced           = v2.FixtureSilenced
	FixtureAdhocRequest       = v2.FixtureAdhocRequest
	FixtureTokens             = v2.FixtureTokens
)

// Misc functions and vars
var (
	SetContextFromResource      = v2.SetContextFromResource
	NewKeepaliveRecord          = v2.NewKeepaliveRecord
	GetEntitySubscription       = v2.GetEntitySubscription
	OutputMetricFormats         = v2.OutputMetricFormats
	ContextNamespace            = v2.ContextNamespace
	NewCheck                    = v2.NewCheck
	CommonCoreResources         = v2.CommonCoreResources
	SilencedName                = v2.SilencedName
	FakeHandlerCommand          = v2.FakeHandlerCommand
	FakeMutatorCommand          = v2.FakeMutatorCommand
	ValidateName                = v2.ValidateName
	SortCheckConfigsByPredicate = v2.SortCheckConfigsByPredicate
	SortCheckConfigsByName      = v2.SortCheckConfigsByName
	SortEntitiesByPredicate     = v2.SortEntitiesByPredicate
	SortEntitiesByID            = v2.SortEntitiesByID
	SortEntitiesByLastSeen      = v2.SortEntitiesByLastSeen
	SortSilencedByPredicate     = v2.SortSilencedByPredicate
	SortSilencedByName          = v2.SortSilencedByName
	SortSilencedByBegin         = v2.SortSilencedByBegin
	DefaultRedactFields         = v2.DefaultRedactFields
	EventsBySeverity            = v2.EventsBySeverity
	EventsByTimestamp           = v2.EventsByTimestamp
	EventsByLastOk              = v2.EventsByLastOk
	EventFilterAllActions       = v2.EventFilterAllActions
	ValidateOutputMetricFormat  = v2.ValidateOutputMetricFormat
)
