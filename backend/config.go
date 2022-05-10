package backend

import (
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/licensing"
	"golang.org/x/time/rate"
)

const (
	// DefaultEtcdName is the default etcd member node name (single-node cluster only)
	DefaultEtcdName = "default"

	// DefaultEtcdClientURL is the default URL to listen for Etcd clients
	DefaultEtcdClientURL = "http://127.0.0.1:2379"

	// DefaultEtcdPeerURL is the default URL to listen for Etcd peers (single-node cluster only)
	DefaultEtcdPeerURL = "http://127.0.0.1:2380"

	// FlagEventdWorkers defines the number of workers for eventd
	FlagEventdWorkers = "eventd-workers"
	// FlagEventdBufferSize defines the buffer size for eventd
	FlagEventdBufferSize = "eventd-buffer-size"
	// FlagKeepalivedWorkers defines the number of workers for keepalived
	FlagKeepalivedWorkers = "keepalived-workers"
	// FlagKeepalivedBufferSize defines buffer size for keepalived
	FlagKeepalivedBufferSize = "keepalived-buffer-size"
	// FlagPipelinedWorkers defines the number of workers for pipelined
	FlagPipelinedWorkers = "pipelined-workers"
	// FlagPipelinedBufferSize defines the buffer size for pipelined
	FlagPipelinedBufferSize = "pipelined-buffer-size"

	// FlagAgentWriteTimeout specifies the time in seconds to wait before
	// giving up on a write to an agent and disposing of the connection.
	FlagAgentWriteTimeout = "agent-write-timeout"

	// FlagJWTPrivateKeyFile defines the path to the private key file for JWT
	// signatures
	FlagJWTPrivateKeyFile = "jwt-private-key-file"
	// FlagJWTPublicKeyFile defines the path to the public key file for JWT
	// signatures validation
	FlagJWTPublicKeyFile = "jwt-public-key-file"
)

type StoreConfig struct {
	// ConfigurationStore specifies the selected configuration store to use (either "postgres", "etcd", or "dev")
	ConfigurationStore string

	// StateStore specifies the selected state store to use (either "postgres" or "dev")
	StateStore string

	// PostgresConfigurationStore contains postgres configuration store details. It's only valid to set this when
	// ConfigurationStore is set to be "postgres".
	PostgresConfigurationStore PostgresConfig

	// EtcdConfigurationStore contains etcd configuration store details. It's only valid to set this when
	// ConfigurationStore is set to be "etcd".
	EtcdConfigurationStore EtcdConfig

	// PostgresStateStore contains postgres state store details. It's only valid to set this when
	// StateStore is set to be "postgres".
	PostgresStateStore PostgresConfig
}

type EtcdConfig struct {
	ClientTLSInfo     etcd.TLSInfo
	URLs              []string
	Username          string
	Password          string
	LogLevel          string
	UseEmbeddedClient bool
}

type PostgresConfig struct {
	DSN string
}

// Config specifies a Backend configuration.
type Config struct {
	// Backend Configuration
	StateDir string
	CacheDir string

	// Agentd Configuration
	AgentHost         string
	AgentPort         int
	AgentTLSOptions   *corev2.TLSOptions
	AgentWriteTimeout int

	// Apid Configuration
	APIListenAddress string
	APIRequestLimit  int64
	APIURL           string
	APIWriteTimeout  time.Duration

	// AssetsRateLimit is the maximum number of assets per second that will be fetched.
	AssetsRateLimit rate.Limit

	// AssetsBurstLimit is the maximum amount of burst allowed in a rate interval.
	AssetsBurstLimit int

	// Dashboardd Configuration
	DashboardHost         string
	DashboardPort         int
	DashboardTLSCertFile  string
	DashboardTLSKeyFile   string
	DashboardWriteTimeout time.Duration

	// Pipelined Configuration
	DeregistrationHandler string

	// Labels are key-value pairs that users can provide to backend entities
	Labels map[string]string

	// Annotations are key-value pairs that users can provide to backend entities
	Annotations map[string]string

	// DevMode starts up a single-node embedded etcd server when enabled.
	DevMode bool

	TLS *corev2.TLSOptions

	LogLevel string

	LicenseGetter licensing.Getter

	DisablePlatformMetrics         bool
	PlatformMetricsLoggingInterval time.Duration
	PlatformMetricsLogFile         string

	EventLogBufferSize       int
	EventLogBufferWait       time.Duration
	EventLogFile             string
	EventLogParallelEncoders bool

	Store StoreConfig
}
