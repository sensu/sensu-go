package backend

import (
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/types"
)

const (
	// DefaultEtcdName is the default etcd member node name (single-node cluster only)
	DefaultEtcdName = "default"

	// DefaultEtcdClientURL is the default URL to listen for Etcd clients
	DefaultEtcdClientURL = "http://127.0.0.1:2379"

	// DefaultEtcdPeerURL is the default URL to listen for Etcd peers (single-node cluster only)
	DefaultEtcdPeerURL = "http://127.0.0.1:2380"
)

// Config specifies a Backend configuration.
type Config struct {
	// Backend Configuration
	StateDir string
	CacheDir string

	// Agentd Configuration
	AgentHost string
	AgentPort int

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
	EtcdClientTLSInfo etcd.TLSInfo
	EtcdPeerTLSInfo   etcd.TLSInfo
	EtcdCipherSuites  []string

	TLS *types.TLSOptions
}
