package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/sensu/sensu-go/backend"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/path"
	"github.com/sensu/sensu-go/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(newVersionCommand())
	rootCmd.AddCommand(newStartCommand())
}

func newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the sensu-backend version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("sensu-backend version %s, build %s, built %s\n",
				version.WithIteration(),
				version.BuildSHA,
				version.BuildDate,
			)
		},
	}

	return cmd
}

func newStartCommand() *cobra.Command {
	var setupErr error

	cmd := &cobra.Command{
		Use:   "start",
		Short: "start the sensu backend",
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.BindPFlags(cmd.Flags())
			if setupErr != nil {
				return setupErr
			}

			cfg := &backend.Config{
				AgentHost:             viper.GetString("agent-host"),
				AgentPort:             viper.GetInt("agent-port"),
				APIHost:               viper.GetString("api-host"),
				APIPort:               viper.GetInt("api-port"),
				DashboardDir:          viper.GetString("dashboard-dir"),
				DashboardHost:         viper.GetString("dashboard-host"),
				DashboardPort:         viper.GetInt("dashboard-port"),
				DeregistrationHandler: viper.GetString("deregistration-handler"),
				StateDir:              viper.GetString("state-dir"),

				EtcdListenClientURL:         viper.GetString("store-client-url"),
				EtcdListenPeerURL:           viper.GetString("store-peer-url"),
				EtcdInitialCluster:          viper.GetString("store-initial-cluster"),
				EtcdInitialClusterState:     viper.GetString("store-initial-cluster-state"),
				EtcdInitialAdvertisePeerURL: viper.GetString("store-initial-advertise-peer-url"),
				EtcdInitialClusterToken:     viper.GetString("store-initial-cluster-token"),
				EtcdName:                    viper.GetString("store-node-name"),
			}

			certFile := viper.GetString("cert-file")
			keyFile := viper.GetString("key-file")
			trustedCAFile := viper.GetString("trusted-ca-file")
			insecureSkipTLSVerify := viper.GetBool("insecure-skip-tls-verify")

			if certFile != "" && keyFile != "" && trustedCAFile != "" {
				cfg.TLS = &types.TLSOptions{certFile, keyFile, trustedCAFile, insecureSkipTLSVerify}
			} else if certFile != "" || keyFile != "" || trustedCAFile != "" {
				emptyFlags := []string{}
				if certFile == "" {
					emptyFlags = append(emptyFlags, "cert-file")
				}
				if keyFile == "" {
					emptyFlags = append(emptyFlags, "key-file")
				}
				if trustedCAFile == "" {
					emptyFlags = append(emptyFlags, "trusted-ca-file")
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

	// Set up distinct flagset for handling config file
	configFlagSet := pflag.NewFlagSet("sensu", pflag.ContinueOnError)
	configFlagSet.StringP("config-file", "c", filepath.Join(path.SystemConfigDir(), "backend.yml"), "path to sensu-backend config file")
	configFlagSet.SetOutput(ioutil.Discard)
	configFlagSet.Parse(os.Args[1:])

	// Split given config file path / filename
	configFile, _ := configFlagSet.GetString("config-file")
	// configDir := filepath.Dir(configFile)
	// configName := filepath.Base(configFile)

	// Configure location of backend configuration
	// viper.AddConfigPath(configDir)
	// viper.SetConfigName(configName)
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configFile)
	setupErr = viper.ReadInConfig()

	// Flag defaults
	viper.SetDefault("agent-host", "[::]")
	viper.SetDefault("agent-port", 8081)
	viper.SetDefault("api-host", "[::]")
	viper.SetDefault("api-port", 8080)
	viper.SetDefault("dashboard-dir", "")
	viper.SetDefault("dashboard-host", "[::]")
	viper.SetDefault("dashboard-port", 3000)
	viper.SetDefault("deregistration-handler", "")
	viper.SetDefault("state-dir", path.SystemDataDir())
	viper.SetDefault("cert-file", "")
	viper.SetDefault("key-file", "")
	viper.SetDefault("trusted-ca-file", "")
	viper.SetDefault("insecure-skip-tls-verify", "")

	// Etcd defaults
	viper.SetDefault("store-client-url", "")
	viper.SetDefault("store-peer-url", "")
	viper.SetDefault("store-initial-cluster", "")
	viper.SetDefault("store-initial-advertise-peer-url", "")
	viper.SetDefault("store-initial-cluster-state", "")
	viper.SetDefault("store-initial-cluster-token", "")
	viper.SetDefault("store-node-name", "")

	// Merge in config flag set so that it appears in command usage
	cmd.Flags().AddFlagSet(configFlagSet)

	// Flags
	cmd.Flags().String("agent-host", viper.GetString("agent-host"), "agent listener host")
	cmd.Flags().Int("agent-port", viper.GetInt("agent-port"), "agent listener port")
	cmd.Flags().String("api-host", viper.GetString("api-host"), "http api listener host")
	cmd.Flags().Int("api-port", viper.GetInt("api-port"), "http api port")
	cmd.Flags().String("dashboard-dir", viper.GetString("dashboard-dir"), "path to sensu dashboard static assets")
	cmd.Flags().String("dashboard-host", viper.GetString("dashboard-host"), "dashboard listener host")
	cmd.Flags().Int("dashboard-port", viper.GetInt("dashboard-port"), "dashboard listener port")
	cmd.Flags().String("deregistration-handler", viper.GetString("deregistration-handler"), "default deregistration handler")
	cmd.Flags().StringP("state-dir", "d", viper.GetString("state-dir"), "path to sensu state storage")
	cmd.Flags().String("cert-file", "", "tls certificate")
	cmd.Flags().String("key-file", "", "tls certificate key")
	cmd.Flags().String("trusted-ca-file", "", "tls certificate authority")
	cmd.Flags().Bool("insecure-skip-tls-verify", false, "skip ssl verification")

	// Etcd flags
	cmd.Flags().String("store-client-url", "", "store client listen URL")
	cmd.Flags().String("store-peer-url", "", "store peer URL")
	cmd.Flags().String("store-initial-cluster", "", "store initial cluster")
	cmd.Flags().String("store-initial-advertise-peer-url", "", "store initial advertise peer URL")
	cmd.Flags().String("store-initial-cluster-state", "", "store initial cluster state")
	cmd.Flags().String("store-initial-cluster-token", "", "store initial cluster token")
	cmd.Flags().String("store-node-name", "", "store cluster member node name")

	return cmd
}
