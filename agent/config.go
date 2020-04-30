package agent

import (
	"io/ioutil"
	"os"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"golang.org/x/time/rate"
)

const (
	// MaxMessageBufferSize specifies the maximum number of messages of a given
	// type that an agent will queue before rejecting messages.
	MaxMessageBufferSize = 10

	// TCPSocketReadDeadline specifies the maximum time the TCP socket will wait
	// to receive data.
	TCPSocketReadDeadline = 500 * time.Millisecond

	// DefaultAPIHost specifies the default API Host
	DefaultAPIHost = "127.0.0.1"

	// DefaultAPIPort specifies the default API Port
	DefaultAPIPort = 3031

	// DefaultBackendURL specifies the default backend URL
	DefaultBackendURL = "ws://127.0.0.1:8081"

	// DefaultEventsAPIRateLimit defines the rate limit, in events per second,
	// for outgoing events.
	DefaultEventsAPIRateLimit rate.Limit = 10.0

	// DefaultEventsAPIBurstLimit defines the burst ceiling for a rate limited
	// events API. If DefaultEventsAPIRateLimit is 0, then the setting has no
	// effect.
	DefaultEventsAPIBurstLimit int = 10

	// DefaultKeepaliveInterval specifies the default keepalive interval
	DefaultKeepaliveInterval = 20

	// DefaultNamespace specifies the default namespace
	DefaultNamespace = "default"

	// DefaultPassword specifies the default password
	DefaultPassword = "P@ssw0rd!"

	// DefaultSocketHost specifies the default socket host
	DefaultSocketHost = "127.0.0.1"

	// DefaultSocketPort specifies the default socket port
	DefaultSocketPort = 3030

	// DefaultStatsdDisable specifies if the statsd listener is disabled
	DefaultStatsdDisable = false

	// DefaultStatsdFlushInterval specifies the default flush interval for statsd
	DefaultStatsdFlushInterval = 10

	// DefaultStatsdMetricsHost specifies the default metrics host for statsd server
	DefaultStatsdMetricsHost = "127.0.0.1"

	// DefaultStatsdMetricsPort specifies the default metrics port for statsd server
	DefaultStatsdMetricsPort = 8125

	// DefaultSystemInfoRefreshInterval specifies the default refresh interval
	// (in seconds) for the agent's cached system information.
	DefaultSystemInfoRefreshInterval = 20

	// DefaultUser specifies the default user
	DefaultUser = "agent"
)

// A Config specifies Agent configuration.
type Config struct {
	// AgentName is the entity name for the running agent. Default is hostname.
	AgentName string

	// AllowList is the path to agent execution allow list configuration file.
	AllowList string

	// API contains the Sensu client HTTP API configuration
	API *APIConfig

	// AssetsRateLimit is the maximum number of assets per second that will be fetched.
	AssetsRateLimit rate.Limit

	// AssetsBurstLimit is the maximum amount of burst allowed in a rate interval.
	AssetsBurstLimit int

	// BackendURLs is a list of URLs for the Sensu Backend. Default:
	// ws://127.0.0.1:8081
	BackendURLs []string

	// CacheDir path where cached data is stored
	CacheDir string

	// Deregister indicates whether the entity is ephemeral
	Deregister bool

	// DeregistrationHandler specifies a single deregistration handler
	DeregistrationHandler string

	// DetectCloudProvider enables cloud provider detection mechanisms.
	// When enabled, the agent will attempt to read files, resolve hostnames,
	// and make HTTP requests to determine what cloud environment it is running
	// in.
	DetectCloudProvider bool

	// DisableAPI disables the events API
	DisableAPI bool

	// DisableAssets stops the agent from downloading and deploying assets
	// in check execution.
	DisableAssets bool

	// DisableSockets disables the event sockets
	DisableSockets bool

	// EventsAPIRateLimit is the maximum number of events per second that will
	// be transmitted to the backend from the events API
	EventsAPIRateLimit rate.Limit

	// EventsAPIBurstLimit is the maximum amount of burst allowed in a rate
	// interval.
	EventsAPIBurstLimit int

	// KeepaliveHandlers contains the handlers to use for the agent's keepalive
	// events
	KeepaliveHandlers []string

	// KeepaliveInterval is the interval between keepalive events.
	KeepaliveInterval uint32

	// KeepaliveWarningTimeout is the time after which a sensu-agent is considered dead
	// by the backend to create a warning event. See DefaultKeepaliveTimeout in
	// corev2 package for default value.
	KeepaliveWarningTimeout uint32

	// KeepaliveCriticalTimeout is the time after which a sensu-agent is considered dead
	// by the backend to create a critical event.
	KeepaliveCriticalTimeout uint32

	// Labels are key-value pairs that users can provide to agent entities
	Labels map[string]string

	// Annotations are key-value pairs that users can provide to agent entities
	Annotations map[string]string

	// Namespace sets the Agent's RBAC namespace identifier
	Namespace string

	// Password sets Agent's password
	Password string

	// Redact contains the fields to redact when marshalling the agent's entity
	Redact []string

	// Socket contains the Sensu client socket configuration
	Socket *SocketConfig

	// StatsdServer contains the statsd server configuration
	StatsdServer *StatsdServerConfig

	// Subscriptions is an array of subscription names. Default: empty array.
	Subscriptions []string

	// TLS sets the TLSConfig for agent TLS options
	TLS *corev2.TLSOptions

	// User sets the Agent's username
	User string

	// BackendHandshakeTimeout specifies the maximum time (in seconds) to wait for
	// the handshake with the backend to complete when opening a connection. If a
	// timeout occurs, the agent will attempt to reconnect with exponential
	// backoff
	BackendHandshakeTimeout int

	// BackendHeartbeatInterval specifies the interval at which the agent must
	// send a heartbeat to the backend
	BackendHeartbeatInterval int

	// BackendHeartbeatTimeout specifies the maximum time (in seconds) to wait for
	// a response to a heartbeat from the backend.  If a timeout occurs, the agent
	// will close the existing connection with the backend and attempt to
	// reconnect with exponential backoff
	BackendHeartbeatTimeout int
}

// StatsdServerConfig contains the statsd server configuration
type StatsdServerConfig struct {
	Host          string
	Port          int
	FlushInterval int
	Handlers      []string
	Disable       bool
}

// SocketConfig contains the Socket configuration
type SocketConfig struct {
	Host string
	Port int
}

// FixtureConfig provides a new Config object initialized with defaults for use
// in tests, as well as a cleanup function to call at the end of the test.
func FixtureConfig() (*Config, func()) {
	cacheDir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}

	c := &Config{
		AgentName: GetDefaultAgentName(),
		API: &APIConfig{
			Host: DefaultAPIHost,
			Port: DefaultAPIPort,
		},
		AssetsRateLimit:         asset.DefaultAssetsRateLimit,
		AssetsBurstLimit:        asset.DefaultAssetsBurstLimit,
		BackendURLs:             []string{},
		CacheDir:                cacheDir,
		EventsAPIRateLimit:      DefaultEventsAPIRateLimit,
		EventsAPIBurstLimit:     DefaultEventsAPIBurstLimit,
		KeepaliveInterval:       DefaultKeepaliveInterval,
		KeepaliveWarningTimeout: corev2.DefaultKeepaliveTimeout,
		Namespace:               DefaultNamespace,
		Password:                DefaultPassword,
		Socket: &SocketConfig{
			Host: DefaultSocketHost,
			Port: DefaultSocketPort,
		},
		StatsdServer: &StatsdServerConfig{
			Host:          DefaultStatsdMetricsHost,
			Port:          DefaultStatsdMetricsPort,
			FlushInterval: DefaultStatsdFlushInterval,
			Handlers:      []string{},
			Disable:       DefaultStatsdDisable,
		},
		User: DefaultUser,
	}
	return c, func() {
		if err := os.RemoveAll(cacheDir); err != nil {
			logger.Debugf("Error removing test agent cache dir: %s", err)
		}
	}
}

// NewConfig provides a new empty Config object
func NewConfig() *Config {
	c := &Config{
		API:          &APIConfig{},
		Socket:       &SocketConfig{},
		StatsdServer: &StatsdServerConfig{},
	}
	return c
}
