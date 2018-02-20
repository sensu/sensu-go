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

const (
	// Flag constants
	flagConfigFile            = "config-file"
	flagAgentHost             = "agent-host"
	flagAgentPort             = "agent-port"
	flagAPIHost               = "api-host"
	flagAPIPort               = "api-port"
	flagDashboardDir          = "dashboard-dir"
	flagDashboardHost         = "dashboard-host"
	flagDashboardPort         = "dashboard-port"
	flagDeregistrationHandler = "deregistration-handler"
	flagStateDir              = "state-dir"
	flagCertFile              = "cert-file"
	flagKeyFile               = "key-file"
	flagTrustedCAFile         = "trusted-ca-file"
	flagInsecureSkipTLSVerify = "insecure-skip-tls-verify"

	// Etcd flag constants
	flagStoreClientURL               = "listen-client-urls"
	flagStorePeerURL                 = "listen-peer-urls"
	flagStoreInitialCluster          = "initial-cluster"
	flagStoreInitialAdvertisePeerURL = "initial-advertise-peer-urls"
	flagStoreInitialClusterState     = "initial-cluster-state"
	flagStoreInitialClusterToken     = "initial-cluster-token"
	flagStoreNodeName                = "name"
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
				version.Semver(),
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
			_ = viper.BindPFlags(cmd.Flags())
			if setupErr != nil {
				return setupErr
			}

			cfg := &backend.Config{
				AgentHost:             viper.GetString(flagAgentHost),
				AgentPort:             viper.GetInt(flagAgentPort),
				APIHost:               viper.GetString(flagAPIHost),
				APIPort:               viper.GetInt(flagAPIPort),
				DashboardDir:          viper.GetString(flagDashboardDir),
				DashboardHost:         viper.GetString(flagDashboardHost),
				DashboardPort:         viper.GetInt(flagDashboardPort),
				DeregistrationHandler: viper.GetString(flagDeregistrationHandler),
				StateDir:              viper.GetString(flagStateDir),

				EtcdListenClientURL:         viper.GetString(flagStoreClientURL),
				EtcdListenPeerURL:           viper.GetString(flagStorePeerURL),
				EtcdInitialCluster:          viper.GetString(flagStoreInitialCluster),
				EtcdInitialClusterState:     viper.GetString(flagStoreInitialClusterState),
				EtcdInitialAdvertisePeerURL: viper.GetString(flagStoreInitialAdvertisePeerURL),
				EtcdInitialClusterToken:     viper.GetString(flagStoreInitialClusterToken),
				EtcdName:                    viper.GetString(flagStoreNodeName),
			}

			certFile := viper.GetString(flagCertFile)
			keyFile := viper.GetString(flagKeyFile)
			trustedCAFile := viper.GetString(flagTrustedCAFile)
			insecureSkipTLSVerify := viper.GetBool(flagInsecureSkipTLSVerify)

			if certFile != "" && keyFile != "" && trustedCAFile != "" {
				cfg.TLS = &types.TLSOptions{
					CertFile:           certFile,
					KeyFile:            keyFile,
					TrustedCAFile:      trustedCAFile,
					InsecureSkipVerify: insecureSkipTLSVerify,
				}
			} else if certFile != "" || keyFile != "" || trustedCAFile != "" {
				emptyFlags := []string{}
				if certFile == "" {
					emptyFlags = append(emptyFlags, flagCertFile)
				}
				if keyFile == "" {
					emptyFlags = append(emptyFlags, flagKeyFile)
				}
				if trustedCAFile == "" {
					emptyFlags = append(emptyFlags, flagTrustedCAFile)
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

			if len(args) == 1 && args[0] == "migration" {
				return sensuBackend.Migration()
			}

			return sensuBackend.Run()
		},
	}

	// Set up distinct flagset for handling config file
	configFlagSet := pflag.NewFlagSet("sensu", pflag.ContinueOnError)
	configFlagSet.StringP(flagConfigFile, "c", "", "path to sensu-backend config file")
	configFlagSet.SetOutput(ioutil.Discard)
	_ = configFlagSet.Parse(os.Args[1:])

	// Get the given config file path
	configFile, _ := configFlagSet.GetString(flagConfigFile)
	configFilePath := configFile

	// use the default config path if flagConfigFile was used
	if configFile == "" {
		configFilePath = filepath.Join(path.SystemConfigDir(), "backend.yml")
	}

	// Configure location of backend configuration
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configFilePath)

	// Flag defaults
	viper.SetDefault(flagAgentHost, "[::]")
	viper.SetDefault(flagAgentPort, 8081)
	viper.SetDefault(flagAPIHost, "[::]")
	viper.SetDefault(flagAPIPort, 8080)
	viper.SetDefault(flagDashboardDir, "")
	viper.SetDefault(flagDashboardHost, "[::]")
	viper.SetDefault(flagDashboardPort, 3000)
	viper.SetDefault(flagDeregistrationHandler, "")
	viper.SetDefault(flagStateDir, path.SystemDataDir())
	viper.SetDefault(flagCertFile, "")
	viper.SetDefault(flagKeyFile, "")
	viper.SetDefault(flagTrustedCAFile, "")
	viper.SetDefault(flagInsecureSkipTLSVerify, false)

	// Etcd defaults
	viper.SetDefault(flagStoreClientURL, "")
	viper.SetDefault(flagStorePeerURL, "")
	viper.SetDefault(flagStoreInitialCluster, "")
	viper.SetDefault(flagStoreInitialAdvertisePeerURL, "")
	viper.SetDefault(flagStoreInitialClusterState, "")
	viper.SetDefault(flagStoreInitialClusterToken, "")
	viper.SetDefault(flagStoreNodeName, "")

	// Merge in config flag set so that it appears in command usage
	cmd.Flags().AddFlagSet(configFlagSet)

	// Flags
	cmd.Flags().String(flagAgentHost, viper.GetString(flagAgentHost), "agent listener host")
	cmd.Flags().Int(flagAgentPort, viper.GetInt(flagAgentPort), "agent listener port")
	cmd.Flags().String(flagAPIHost, viper.GetString(flagAPIHost), "http api listener host")
	cmd.Flags().Int(flagAPIPort, viper.GetInt(flagAPIPort), "http api port")
	cmd.Flags().String(flagDashboardDir, viper.GetString(flagDashboardDir), "path to sensu dashboard static assets")
	cmd.Flags().String(flagDashboardHost, viper.GetString(flagDashboardHost), "dashboard listener host")
	cmd.Flags().Int(flagDashboardPort, viper.GetInt(flagDashboardPort), "dashboard listener port")
	cmd.Flags().String(flagDeregistrationHandler, viper.GetString(flagDeregistrationHandler), "default deregistration handler")
	cmd.Flags().StringP(flagStateDir, "d", viper.GetString(flagStateDir), "path to sensu state storage")
	cmd.Flags().String(flagCertFile, viper.GetString(flagCertFile), "tls certificate")
	cmd.Flags().String(flagKeyFile, viper.GetString(flagKeyFile), "tls certificate key")
	cmd.Flags().String(flagTrustedCAFile, viper.GetString(flagTrustedCAFile), "tls certificate authority")
	cmd.Flags().Bool(flagInsecureSkipTLSVerify, viper.GetBool(flagInsecureSkipTLSVerify), "skip ssl verification")

	// Etcd flags
	cmd.Flags().String(flagStoreClientURL, viper.GetString(flagStoreClientURL), "store listen client URL")
	cmd.Flags().String(flagStorePeerURL, viper.GetString(flagStorePeerURL), "store listen peer URL")
	cmd.Flags().String(flagStoreInitialCluster, viper.GetString(flagStoreInitialCluster), "store initial cluster")
	cmd.Flags().String(flagStoreInitialAdvertisePeerURL, viper.GetString(flagStoreInitialAdvertisePeerURL), "store initial advertise peer URL")
	cmd.Flags().String(flagStoreInitialClusterState, viper.GetString(flagStoreInitialClusterState), "store initial cluster state")
	cmd.Flags().String(flagStoreInitialClusterToken, viper.GetString(flagStoreInitialClusterToken), "store initial cluster token")
	cmd.Flags().String(flagStoreNodeName, viper.GetString(flagStoreNodeName), "store cluster member node name")

	// Load the configuration file but only error out if flagConfigFile is used
	if err := viper.ReadInConfig(); err != nil && configFile != "" {
		setupErr = err
	}

	return cmd
}
