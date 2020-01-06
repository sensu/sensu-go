package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend"
	"github.com/sensu/sensu-go/backend/authentication"
	"github.com/sensu/sensu-go/backend/authentication/providers/basic"
	"github.com/sensu/sensu-go/backend/etcd"
	etcdstore "github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/sdk"
	"github.com/sensu/sensu-go/util/path"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func fallbackStringSlice(newFlag, oldFlag string) []string {
	slice := viper.GetStringSlice(newFlag)
	if len(slice) == 0 {
		slice = viper.GetStringSlice(oldFlag)
	}
	return slice
}

const (
	// Flag constants
	flagConfigFile            = "config-file"
	flagAgentHost             = "agent-host"
	flagAgentPort             = "agent-port"
	flagAPIListenAddress      = "api-listen-address"
	flagAPIURL                = "api-url"
	flagDashboardHost         = "dashboard-host"
	flagDashboardPort         = "dashboard-port"
	flagDashboardCertFile     = "dashboard-cert-file"
	flagDashboardKeyFile      = "dashboard-key-file"
	flagDeregistrationHandler = "deregistration-handler"
	flagCacheDir              = "cache-dir"
	flagStateDir              = "state-dir"
	flagCertFile              = "cert-file"
	flagKeyFile               = "key-file"
	flagTrustedCAFile         = "trusted-ca-file"
	flagInsecureSkipTLSVerify = "insecure-skip-tls-verify"
	flagDebug                 = "debug"
	flagLogLevel              = "log-level"

	// Etcd flag constants
	flagEtcdClientURLs               = "etcd-client-urls"
	flagEtcdListenClientURLs         = "etcd-listen-client-urls"
	flagEtcdPeerURLs                 = "etcd-listen-peer-urls"
	flagEtcdInitialCluster           = "etcd-initial-cluster"
	flagEtcdDiscovery                = "etcd-discovery"
	flagEtcdDiscoverySrv             = "etcd-discovery-srv"
	flagEtcdInitialAdvertisePeerURLs = "etcd-initial-advertise-peer-urls"
	flagEtcdInitialClusterState      = "etcd-initial-cluster-state"
	flagEtcdInitialClusterToken      = "etcd-initial-cluster-token"
	flagEtcdNodeName                 = "etcd-name"
	flagNoEmbedEtcd                  = "no-embed-etcd"
	flagEtcdAdvertiseClientURLs      = "etcd-advertise-client-urls"
	flagEtcdHeartbeatInterval        = "etcd-heartbeat-interval"
	flagEtcdElectionTimeout          = "etcd-election-timeout"

	// Etcd TLS flag constants
	flagEtcdCertFile           = "etcd-cert-file"
	flagEtcdKeyFile            = "etcd-key-file"
	flagEtcdClientCertAuth     = "etcd-client-cert-auth"
	flagEtcdTrustedCAFile      = "etcd-trusted-ca-file"
	flagEtcdPeerCertFile       = "etcd-peer-cert-file"
	flagEtcdPeerKeyFile        = "etcd-peer-key-file"
	flagEtcdPeerClientCertAuth = "etcd-peer-client-cert-auth"
	flagEtcdPeerTrustedCAFile  = "etcd-peer-trusted-ca-file"
	flagEtcdCipherSuites       = "etcd-cipher-suites"
	flagEtcdMaxRequestBytes    = "etcd-max-request-bytes"
	flagEtcdQuotaBackendBytes  = "etcd-quota-backend-bytes"

	// Default values

	// defaultEtcdClientURL is the default URL to listen for Etcd clients
	defaultEtcdClientURL = "http://127.0.0.1:2379"
	// defaultEtcdName is the default etcd member node name (single-node cluster
	// only)
	defaultEtcdName = "default"
	// defaultEtcdPeerURL is the default URL to listen for Etcd peers (single-node
	// cluster only)
	defaultEtcdPeerURL = "http://127.0.0.1:2380"

	// defaultEtcdAdvertiseClientURL is the default list of this member's client
	// URLs to advertise to the rest of the cluster
	defaultEtcdAdvertiseClientURL = "http://localhost:2379"
)

func SDKCommand() *cobra.Command {
	var setupErr error
	cmd := &cobra.Command{
		Use:           "sensu-sdk",
		Short:         "run the sensu developer kit",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = viper.BindPFlags(cmd.Flags())
			if setupErr != nil {
				return setupErr
			}

			cfg := &backend.Config{
				EtcdClientURLs:      fallbackStringSlice(flagEtcdClientURLs, flagEtcdAdvertiseClientURLs),
				EtcdCipherSuites:    viper.GetStringSlice(flagEtcdCipherSuites),
				EtcdMaxRequestBytes: viper.GetUint(flagEtcdMaxRequestBytes),
				NoEmbedEtcd:         true,
			}

			// Sensu APIs TLS config
			certFile := viper.GetString(flagCertFile)
			keyFile := viper.GetString(flagKeyFile)
			insecureSkipTLSVerify := viper.GetBool(flagInsecureSkipTLSVerify)
			trustedCAFile := viper.GetString(flagTrustedCAFile)

			if certFile != "" && keyFile != "" {
				cfg.TLS = &corev2.TLSOptions{
					CertFile:           certFile,
					KeyFile:            keyFile,
					TrustedCAFile:      trustedCAFile,
					InsecureSkipVerify: insecureSkipTLSVerify,
				}
			} else if certFile != "" || keyFile != "" {
				return fmt.Errorf(
					"tls configuration error, both flags --%s & --%s are required",
					flagCertFile, flagKeyFile)
			}

			// Etcd TLS config
			cfg.EtcdClientTLSInfo = etcd.TLSInfo{
				CertFile:       viper.GetString(flagEtcdCertFile),
				KeyFile:        viper.GetString(flagEtcdKeyFile),
				TrustedCAFile:  viper.GetString(flagEtcdTrustedCAFile),
				ClientCertAuth: viper.GetBool(flagEtcdClientCertAuth),
			}

			// Convert the TLS config into etcd's transport.TLSInfo
			tlsInfo := (transport.TLSInfo)(cfg.EtcdClientTLSInfo)
			tlsConfig, err := tlsInfo.ClientConfig()
			if err != nil {
				return err
			}

			clientURLs := viper.GetStringSlice(flagEtcdClientURLs)
			if len(clientURLs) == 0 {
				clientURLs = viper.GetStringSlice(flagEtcdAdvertiseClientURLs)
			}

			client, err := clientv3.New(clientv3.Config{
				Endpoints:   clientURLs,
				DialTimeout: 1 * time.Second,
				TLS:         tlsConfig,
				DialOptions: []grpc.DialOption{
					grpc.WithBlock(),
				},
			})

			if err != nil {
				return fmt.Errorf("error connecting to cluster: %s", err)
			}

			authenticator := &authentication.Authenticator{}
			basic := &basic.Provider{
				ObjectMeta: corev2.ObjectMeta{Name: basic.Type},
				Store:      etcdstore.NewStore(client, cfg.EtcdName),
			}
			authenticator.AddProvider(basic)

			apis := map[string]interface{}{
				"sensu": etcdstore.NewStore(client, cfg.EtcdName),
			}

			interpreter := sdk.NewInterpreter(authenticator, apis)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			return interpreter.Run(ctx)
		},
	}

	setupErr = handleConfig(cmd)

	return cmd
}

func handleConfig(cmd *cobra.Command) error {
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

	// Etcd defaults
	viper.SetDefault(flagEtcdAdvertiseClientURLs, defaultEtcdAdvertiseClientURL)
	viper.SetDefault(flagEtcdListenClientURLs, defaultEtcdClientURL)
	viper.SetDefault(flagEtcdPeerURLs, defaultEtcdPeerURL)
	viper.SetDefault(flagEtcdInitialCluster, "")
	viper.SetDefault(flagEtcdDiscovery, "")
	viper.SetDefault(flagEtcdDiscoverySrv, "")
	viper.SetDefault(flagEtcdInitialAdvertisePeerURLs, defaultEtcdPeerURL)
	viper.SetDefault(flagEtcdInitialClusterState, etcd.ClusterStateNew)
	viper.SetDefault(flagEtcdInitialClusterToken, "")
	viper.SetDefault(flagEtcdNodeName, defaultEtcdName)
	viper.SetDefault(flagEtcdQuotaBackendBytes, etcd.DefaultQuotaBackendBytes)
	viper.SetDefault(flagEtcdMaxRequestBytes, etcd.DefaultMaxRequestBytes)
	viper.SetDefault(flagEtcdHeartbeatInterval, etcd.DefaultTickMs)
	viper.SetDefault(flagEtcdElectionTimeout, etcd.DefaultElectionMs)

	// Merge in config flag set so that it appears in command usage
	cmd.Flags().AddFlagSet(configFlagSet)

	// Etcd client/server flags
	cmd.Flags().StringSlice(flagEtcdCipherSuites, nil, "list of ciphers to use for etcd TLS configuration")
	_ = cmd.Flags().SetAnnotation(flagEtcdCipherSuites, "categories", []string{"store"})

	// This one is really only a server flag, but because we lacked
	// --etcd-client-urls until recently, it's used as a fallback.
	cmd.Flags().StringSlice(flagEtcdAdvertiseClientURLs, viper.GetStringSlice(flagEtcdAdvertiseClientURLs), "list of this member's client URLs to advertise to clients")
	_ = cmd.Flags().SetAnnotation(flagEtcdAdvertiseClientURLs, "categories", []string{"store"})

	cmd.Flags().Uint(flagEtcdMaxRequestBytes, viper.GetUint(flagEtcdMaxRequestBytes), "maximum etcd request size in bytes (use with caution)")
	_ = cmd.Flags().SetAnnotation(flagEtcdMaxRequestBytes, "categories", []string{"store"})

	// Etcd client/server TLS flags
	cmd.Flags().String(flagEtcdCertFile, viper.GetString(flagEtcdCertFile), "path to the client server TLS cert file")
	_ = cmd.Flags().SetAnnotation(flagEtcdCertFile, "categories", []string{"store"})
	cmd.Flags().String(flagEtcdKeyFile, viper.GetString(flagEtcdKeyFile), "path to the client server TLS key file")
	_ = cmd.Flags().SetAnnotation(flagEtcdKeyFile, "categories", []string{"store"})
	cmd.Flags().Bool(flagEtcdClientCertAuth, viper.GetBool(flagEtcdClientCertAuth), "enable client cert authentication")
	_ = cmd.Flags().SetAnnotation(flagEtcdClientCertAuth, "categories", []string{"store"})
	cmd.Flags().String(flagEtcdTrustedCAFile, viper.GetString(flagEtcdTrustedCAFile), "path to the client server TLS trusted CA cert file")
	_ = cmd.Flags().SetAnnotation(flagEtcdTrustedCAFile, "categories", []string{"store"})
	cmd.Flags().String(flagEtcdClientURLs, viper.GetString(flagEtcdClientURLs), "client URLs to use when operating as an etcd client")
	_ = cmd.Flags().SetAnnotation(flagEtcdClientURLs, "categories", []string{"store"})

	// Load the configuration file but only error out if flagConfigFile is used
	if err := viper.ReadInConfig(); err != nil && configFile != "" {
		return err
	}

	viper.SetEnvPrefix("sensu_backend")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	return nil
}
