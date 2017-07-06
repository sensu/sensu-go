package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"fmt"

	"github.com/sensu/sensu-go/backend"
	"github.com/spf13/cobra"
)

var (
	agentHost             string
	agentPort             int
	apiHost               string
	apiPort               int
	dashboardDir          string
	dashboardHost         string
	dashboardPort         int
	deregistrationHandler string
	stateDir              string

	etcdListenClientURL         string
	etcdListenPeerURL           string
	etcdInitialCluster          string
	etcdInitialClusterState     string
	etcdName                    string
	etcdInitialClusterToken     string
	etcdInitialAdvertisePeerURL string

	certFile       string
	keyFile        string
	clientCertAuth string
)

func init() {
	rootCmd.AddCommand(newStartCommand())
}

func newStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "start the sensu backend",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &backend.Config{
				AgentHost:             agentHost,
				AgentPort:             agentPort,
				APIHost:               apiHost,
				APIPort:               apiPort,
				DashboardDir:          dashboardDir,
				DashboardHost:         dashboardHost,
				DashboardPort:         dashboardPort,
				DeregistrationHandler: deregistrationHandler,
				StateDir:              stateDir,
			}

			// TODO(grep): make configuration of etcd saner.
			if etcdListenClientURL != "" {
				cfg.EtcdListenClientURL = etcdListenClientURL
			}

			if etcdListenPeerURL != "" {
				cfg.EtcdListenPeerURL = etcdListenPeerURL
			}

			if etcdInitialCluster != "" {
				cfg.EtcdInitialCluster = etcdInitialCluster
			}

			if etcdInitialClusterState != "" {
				cfg.EtcdInitialClusterState = etcdInitialClusterState
			}

			if etcdInitialAdvertisePeerURL != "" {
				cfg.EtcdInitialAdvertisePeerURL = etcdInitialAdvertisePeerURL
			}

			if etcdInitialClusterToken != "" {
				cfg.EtcdInitialClusterToken = etcdInitialClusterToken
			}

			if etcdName != "" {
				cfg.EtcdName = etcdName
			}

			if certFile != "" && keyFile != "" && clientCertAuth != "" {
				cfg.TLS = &backend.TLSConfig{certFile, keyFile, clientCertAuth}
			} else if certFile != "" || keyFile != "" || clientCertAuth != "" {
				emptyFlags := []string{}
				if certFile == "" {
					emptyFlags = append(emptyFlags, "cert-file")
				}
				if keyFile == "" {
					emptyFlags = append(emptyFlags, "key-file")
				}
				if clientCertAuth == "" {
					emptyFlags = append(emptyFlags, "client-cert-auth")
				}

				return fmt.Errorf("missing the following cert flags: %s", emptyFlags)
			}

			sensuBackend, err := backend.NewBackend(cfg)
			if err != nil {
				return err
			}

			sigs := make(chan os.Signal, 1)

			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				sig := <-sigs
				logger.Info("signal received: ", sig)
				sensuBackend.Stop()
			}()

			return sensuBackend.Run()
		},
	}

	var defaultStateDir string

	switch runtime.GOOS {
	case "windows":
		programDataDir := os.Getenv("PROGRAMDATA")
		defaultStateDir = filepath.Join(programDataDir, "sensu")
	default:
		defaultStateDir = "/var/lib/sensu"
	}

	cmd.Flags().StringVar(&agentHost, "agent-host", "0.0.0.0", "Agent listener host")
	cmd.Flags().IntVar(&agentPort, "agent-port", 8081, "Agent listener port")
	cmd.Flags().StringVar(&apiHost, "api-host", "0.0.0.0", "HTTP API listener host")
	cmd.Flags().IntVar(&apiPort, "api-port", 8080, "HTTP API port")
	cmd.Flags().StringVar(&dashboardDir, "dashboard-dir", "", "path to sensu dashboard static assets")
	cmd.Flags().StringVar(&dashboardHost, "dashboard-host", "0.0.0.0", "Dashboard listener host")
	cmd.Flags().IntVar(&dashboardPort, "dashboard-port", 3000, "Dashboard listener port")
	cmd.Flags().StringVar(&deregistrationHandler, "deregistration-handler", "", "Default deregistration handler")
	cmd.Flags().StringVarP(&stateDir, "state-dir", "d", defaultStateDir, "path to sensu state storage")

	cmd.Flags().StringVar(&etcdListenClientURL, "store-client-url", "", "store client listen URL")
	cmd.Flags().StringVar(&etcdListenPeerURL, "store-peer-url", "", "store peer URL")
	cmd.Flags().StringVar(&etcdInitialCluster, "store-initial-cluster", "", "store initial cluster")
	cmd.Flags().StringVar(&etcdInitialAdvertisePeerURL, "store-initial-advertise-peer-url", "", "store initial advertise peer URL")
	cmd.Flags().StringVar(&etcdInitialClusterState, "store-initial-cluster-state", "", "store initial cluster state")
	cmd.Flags().StringVar(&etcdInitialClusterToken, "store-initial-cluster-token", "", "store initial cluster token")
	cmd.Flags().StringVar(&etcdName, "store-node-name", "", "store cluster member node name")
	cmd.Flags().StringVar(&certFile, "cert-file", "", "tls certificate")
	cmd.Flags().StringVar(&keyFile, "key-file", "", "tls certificate")
	cmd.Flags().StringVar(&clientCertAuth, "client-cert-auth", "", "tls certificate authority")

	return cmd
}
