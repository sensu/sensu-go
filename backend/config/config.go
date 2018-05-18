package config

import "github.com/sensu/sensu-go/types"

// Config specifies a Backend configuration.
type Backend struct {
	// Backend Configuration
	StateDir string

	// Agentd Configuration
	AgentHost string
	AgentPort int

	// Apid Configuration
	APIHost string
	APIPort int

	// Dashboardd Configuration
	DashboardHost string
	DashboardPort int

	// Pipelined Configuration
	DeregistrationHandler string

	// Etcd configuration
	EtcdInitialAdvertisePeerURL string
	EtcdInitialClusterToken     string
	EtcdInitialClusterState     string
	EtcdInitialCluster          string
	EtcdListenClientURL         string
	EtcdListenPeerURL           string
	EtcdName                    string

	TLS *types.TLSOptions
}
