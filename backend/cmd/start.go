package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/sensu/sensu-go/backend"
	"github.com/spf13/cobra"
)

var (
	agentHost             string
	agentPort             int
	apiAuthentication     bool
	apiHost               string
	apiPort               int
	dashboardDir          string
	dashboardHost         string
	dashboardPort         int
	deregistrationHandler string
	stateDir              string

	etcdClientListenURL string
	etcdPeerListenURL   string
	etcdInitialCluster  string
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
				APIAuthentication:     apiAuthentication,
				APIHost:               apiHost,
				APIPort:               apiPort,
				DashboardDir:          dashboardDir,
				DashboardHost:         dashboardHost,
				DashboardPort:         dashboardPort,
				DeregistrationHandler: deregistrationHandler,
				StateDir:              stateDir,
			}

			// we have defaults for this in the backend config. this is basically
			// because we don't _actually_ want people using these flags. they're
			// mostly just for testing. can we kill these from the shipped binary?
			// - grep
			if etcdClientListenURL != "" {
				cfg.EtcdClientListenURL = etcdClientListenURL
			}

			if etcdPeerListenURL != "" {
				cfg.EtcdPeerListenURL = etcdPeerListenURL
			}

			if etcdInitialCluster != "" {
				cfg.EtcdInitialCluster = etcdInitialCluster
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
	cmd.Flags().BoolVar(&apiAuthentication, "api-authentication", false, "Enable API authentication")
	cmd.Flags().StringVar(&apiHost, "api-host", "0.0.0.0", "HTTP API listener host")
	cmd.Flags().IntVar(&apiPort, "api-port", 8080, "HTTP API port")
	cmd.Flags().StringVar(&dashboardDir, "dashboard-dir", "", "path to sensu dashboard static assets")
	cmd.Flags().StringVar(&dashboardHost, "dashboard-host", "0.0.0.0", "Dashboard listener host")
	cmd.Flags().IntVar(&dashboardPort, "dashboard-port", 3000, "Dashboard listener port")
	cmd.Flags().StringVar(&deregistrationHandler, "deregistration-handler", "", "Default deregistration handler")
	cmd.Flags().StringVarP(&stateDir, "state-dir", "d", defaultStateDir, "path to sensu state storage")

	// For now don't set defaults for these. This allows us to control defaults on NewBackend(). We may wish
	// to do something more interesting here as well--e.g. only expose these settings via some kind of compile
	// feature flag so that they're used only for testing, etc. But for now, we make these easily configurable
	// for end-to-end testing. -grep
	cmd.Flags().StringVar(&etcdClientListenURL, "store-client-url", "", "store client listen URL")
	cmd.Flags().StringVar(&etcdPeerListenURL, "store-peer-url", "", "store peer URL")
	cmd.Flags().StringVar(&etcdInitialCluster, "store-initial-cluster", "", "store initial cluster")

	return cmd
}
