package backend

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/etcd"
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
)

// Config specifies a Backend configuration.
type Config struct {
	// Backend Configuration
	StateDir string
	CacheDir string

	// Agentd Configuration
	AgentHost       string
	AgentPort       int
	AgentTLSOptions *corev2.TLSOptions

	// Apid Configuration
	APIListenAddress string
	APIURL           string

	// Dashboardd Configuration
	DashboardHost        string
	DashboardPort        int
	DashboardTLSCertFile string
	DashboardTLSKeyFile  string

	// Pipelined Configuration
	DeregistrationHandler string

	// Etcd configuration
	EtcdAdvertiseClientURLs      []string
	EtcdInitialAdvertisePeerURLs []string
	EtcdInitialClusterToken      string
	EtcdInitialClusterState      string
	EtcdInitialCluster           string
	EtcdListenClientURLs         []string
	EtcdListenPeerURLs           []string
	EtcdName                     string
	NoEmbedEtcd                  bool

	// Etcd TLS configuration
	EtcdClientTLSInfo     etcd.TLSInfo
	EtcdPeerTLSInfo       etcd.TLSInfo
	EtcdCipherSuites      []string
	EtcdMaxRequestBytes   uint
	EtcdQuotaBackendBytes int64

	TLS *corev2.TLSOptions
}
