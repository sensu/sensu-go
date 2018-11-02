package agent

import (
	"os"
	"path/filepath"
	"time"

	"github.com/sensu/sensu-go/types"
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
	// AgentID is the entity ID for the running agent. Default is hostname.
	AgentID string
	// API contains the Sensu client HTTP API configuration
	API *APIConfig
	// BackendURLs is a list of URLs for the Sensu Backend. Default:
	// ws://127.0.0.1:8081
	BackendURLs []string
	// CacheDir path where cached data is stored
	CacheDir string
	// Deregister indicates whether the entity is ephemeral
	Deregister bool
	// DeregistrationHandler specifies a single deregistration handler
	DeregistrationHandler string
	// ExtendedAttributes contains any extended attributes passed to the agent on
	// start
	ExtendedAttributes []byte
	// KeepaliveInterval is the interval, in seconds, when agents will send a
	// keepalive to sensu-backend.
	KeepaliveInterval uint32
	// KeepaliveTimeout is the time after which a sensu-agent is considered dead
	// by the backend. See DefaultKeepaliveTimeout in types package for default
	// value.
	KeepaliveTimeout uint32
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
	TLS *types.TLSOptions
	// User sets the Agent's username
	User string
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
	cacheDir := filepath.Join(os.TempDir(), "sensu-agent-test")

	c := &Config{
		AgentID: GetDefaultAgentID(),
		API: &APIConfig{
			Host: DefaultAPIHost,
			Port: DefaultAPIPort,
		},
		BackendURLs:       []string{},
		CacheDir:          cacheDir,
		KeepaliveInterval: DefaultKeepaliveInterval,
		KeepaliveTimeout:  types.DefaultKeepaliveTimeout,
		Namespace:         DefaultNamespace,
		Password:          DefaultPassword,
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
