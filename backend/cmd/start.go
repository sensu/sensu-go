package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/util/path"
	stringsutil "github.com/sensu/sensu-go/util/strings"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

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

	// Start command usage template
	startUsageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
	{{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
	{{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
	{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

General Flags:
{{ $flags := categoryFlags "" .LocalFlags }}{{ $flags.FlagUsages | trimTrailingWhitespaces}}

Store Flags:
{{ $storeFlags := categoryFlags "store" .LocalFlags }}{{ $storeFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
	{{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
)

// initializeFunc represents the signature of an initialization function, used
// to initialize the backend
type InitializeFunc func(context.Context, *backend.Config) (*backend.Backend, error)

func fallbackStringSlice(newFlag, oldFlag string) []string {
	slice := viper.GetStringSlice(newFlag)
	if len(slice) == 0 {
		slice = viper.GetStringSlice(oldFlag)
	}
	return slice
}

// StartCommand ...
func StartCommand(initialize InitializeFunc) *cobra.Command {
	var setupErr error

	cmd := &cobra.Command{
		Use:           "start",
		Short:         "start the sensu backend",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = viper.BindPFlags(cmd.Flags())
			if setupErr != nil {
				return setupErr
			}

			level, err := logrus.ParseLevel(viper.GetString(flagLogLevel))
			if err != nil {
				return err
			}
			logrus.SetLevel(level)

			// If no clustering options are provided, default to a static
			// cluster 'defaultEtcdName=defaultEtcdPeerURL'.
			initialCluster := viper.GetString(flagEtcdInitialCluster)
			etcdDiscovery := viper.GetString(flagEtcdDiscovery)
			SrvDiscovery := viper.GetString(flagEtcdDiscoverySrv)

			if initialCluster == "" && etcdDiscovery == "" && SrvDiscovery == "" {
				initialCluster = fmt.Sprintf("%s=%s", defaultEtcdName, defaultEtcdPeerURL)
			}

			cfg := &backend.Config{
				AgentHost:             viper.GetString(flagAgentHost),
				AgentPort:             viper.GetInt(flagAgentPort),
				AgentWriteTimeout:     viper.GetInt(backend.FlagAgentWriteTimeout),
				APIListenAddress:      viper.GetString(flagAPIListenAddress),
				APIURL:                viper.GetString(flagAPIURL),
				DashboardHost:         viper.GetString(flagDashboardHost),
				DashboardPort:         viper.GetInt(flagDashboardPort),
				DashboardTLSCertFile:  viper.GetString(flagDashboardCertFile),
				DashboardTLSKeyFile:   viper.GetString(flagDashboardKeyFile),
				DeregistrationHandler: viper.GetString(flagDeregistrationHandler),
				CacheDir:              viper.GetString(flagCacheDir),
				StateDir:              viper.GetString(flagStateDir),

				EtcdAdvertiseClientURLs:      viper.GetStringSlice(flagEtcdAdvertiseClientURLs),
				EtcdListenClientURLs:         viper.GetStringSlice(flagEtcdListenClientURLs),
				EtcdClientURLs:               fallbackStringSlice(flagEtcdClientURLs, flagEtcdAdvertiseClientURLs),
				EtcdListenPeerURLs:           viper.GetStringSlice(flagEtcdPeerURLs),
				EtcdInitialCluster:           initialCluster,
				EtcdInitialClusterState:      viper.GetString(flagEtcdInitialClusterState),
				EtcdDiscovery:                etcdDiscovery,
				EtcdDiscoverySrv:             SrvDiscovery,
				EtcdInitialAdvertisePeerURLs: viper.GetStringSlice(flagEtcdInitialAdvertisePeerURLs),
				EtcdInitialClusterToken:      viper.GetString(flagEtcdInitialClusterToken),
				EtcdName:                     viper.GetString(flagEtcdNodeName),
				EtcdCipherSuites:             viper.GetStringSlice(flagEtcdCipherSuites),
				EtcdQuotaBackendBytes:        viper.GetInt64(flagEtcdQuotaBackendBytes),
				EtcdMaxRequestBytes:          viper.GetUint(flagEtcdMaxRequestBytes),
				EtcdHeartbeatInterval:        viper.GetUint(flagEtcdHeartbeatInterval),
				EtcdElectionTimeout:          viper.GetUint(flagEtcdElectionTimeout),
				NoEmbedEtcd:                  viper.GetBool(flagNoEmbedEtcd),
			}

			// Sensu APIs TLS config
			certFile := viper.GetString(flagCertFile)
			keyFile := viper.GetString(flagKeyFile)
			insecureSkipTLSVerify := viper.GetBool(flagInsecureSkipTLSVerify)
			// TODO(ccressent gbolo): issue #2548
			// Eventually this should be changed: --insecure-skip-tls-verify --etcd-insecure-skip-tls-verify
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
			cfg.EtcdPeerTLSInfo = etcd.TLSInfo{
				CertFile:       viper.GetString(flagEtcdPeerCertFile),
				KeyFile:        viper.GetString(flagEtcdPeerKeyFile),
				TrustedCAFile:  viper.GetString(flagEtcdPeerTrustedCAFile),
				ClientCertAuth: viper.GetBool(flagEtcdPeerClientCertAuth),
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			sensuBackend, err := initialize(ctx, cfg)
			if err != nil {
				return err
			}

			sigs := make(chan os.Signal, 1)

			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				sig := <-sigs
				logger.Warn("signal received: ", sig)
				cancel()
			}()

			if viper.GetBool(flagDebug) {
				go func() {
					log.Println(http.ListenAndServe("127.0.0.1:6060", nil))
				}()
			}

			return sensuBackend.RunWithInitializer(initialize)
		},
	}

	setupErr = handleConfig(cmd, true)

	return cmd
}

func handleConfig(cmd *cobra.Command, server bool) error {
	// Set up distinct flagset for handling config file
	configFlagSet := pflag.NewFlagSet("sensu", pflag.ContinueOnError)
	configFileDefaultLocation := filepath.Join(path.SystemConfigDir(), "backend.yml")
	configFileDefault := fmt.Sprintf("path to sensu-backend config file (default %q)", configFileDefaultLocation)
	configFlagSet.StringP(flagConfigFile, "c", "", configFileDefault)
	configFlagSet.SetOutput(ioutil.Discard)
	_ = configFlagSet.Parse(os.Args[1:])

	// Get the given config file path
	configFile, _ := configFlagSet.GetString(flagConfigFile)
	configFilePath := configFile

	// use the default config path if flagConfigFile was used
	if configFile == "" {
		configFilePath = configFileDefaultLocation
	}

	// Configure location of backend configuration
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configFilePath)

	if server {
		// Flag defaults
		viper.SetDefault(flagAgentHost, "[::]")
		viper.SetDefault(flagAgentPort, 8081)
		viper.SetDefault(flagAPIListenAddress, "[::]:8080")
		viper.SetDefault(flagAPIURL, "http://localhost:8080")
		viper.SetDefault(flagDashboardHost, "[::]")
		viper.SetDefault(flagDashboardPort, 3000)
		viper.SetDefault(flagDashboardCertFile, "")
		viper.SetDefault(flagDashboardKeyFile, "")
		viper.SetDefault(flagDeregistrationHandler, "")
		viper.SetDefault(flagCacheDir, path.SystemCacheDir("sensu-backend"))
		viper.SetDefault(flagStateDir, path.SystemDataDir("sensu-backend"))
		viper.SetDefault(flagCertFile, "")
		viper.SetDefault(flagKeyFile, "")
		viper.SetDefault(flagTrustedCAFile, "")
		viper.SetDefault(flagInsecureSkipTLSVerify, false)
		viper.SetDefault(flagLogLevel, "warn")
		viper.SetDefault(backend.FlagEventdWorkers, 100)
		viper.SetDefault(backend.FlagEventdBufferSize, 100)
		viper.SetDefault(backend.FlagKeepalivedWorkers, 100)
		viper.SetDefault(backend.FlagKeepalivedBufferSize, 100)
		viper.SetDefault(backend.FlagPipelinedWorkers, 100)
		viper.SetDefault(backend.FlagPipelinedBufferSize, 100)
		viper.SetDefault(backend.FlagAgentWriteTimeout, 15)
	}

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

	if server {
		viper.SetDefault(flagNoEmbedEtcd, false)
	}

	// Merge in config flag set so that it appears in command usage
	cmd.Flags().AddFlagSet(configFlagSet)

	if server {
		// Main Flags
		cmd.Flags().String(flagAgentHost, viper.GetString(flagAgentHost), "agent listener host")
		cmd.Flags().Int(flagAgentPort, viper.GetInt(flagAgentPort), "agent listener port")
		cmd.Flags().String(flagAPIListenAddress, viper.GetString(flagAPIListenAddress), "address to listen on for api traffic")
		cmd.Flags().String(flagAPIURL, viper.GetString(flagAPIURL), "url of the api to connect to")
		cmd.Flags().String(flagDashboardHost, viper.GetString(flagDashboardHost), "dashboard listener host")
		cmd.Flags().Int(flagDashboardPort, viper.GetInt(flagDashboardPort), "dashboard listener port")
		cmd.Flags().String(flagDashboardCertFile, viper.GetString(flagDashboardCertFile), "dashboard TLS certificate in PEM format")
		cmd.Flags().String(flagDashboardKeyFile, viper.GetString(flagDashboardKeyFile), "dashboard TLS certificate key in PEM format")
		cmd.Flags().String(flagDeregistrationHandler, viper.GetString(flagDeregistrationHandler), "default deregistration handler")
		cmd.Flags().String(flagCacheDir, viper.GetString(flagCacheDir), "path to store cached data")
		cmd.Flags().StringP(flagStateDir, "d", viper.GetString(flagStateDir), "path to sensu state storage")
		cmd.Flags().String(flagCertFile, viper.GetString(flagCertFile), "TLS certificate in PEM format")
		cmd.Flags().String(flagKeyFile, viper.GetString(flagKeyFile), "TLS certificate key in PEM format")
		cmd.Flags().String(flagTrustedCAFile, viper.GetString(flagTrustedCAFile), "TLS CA certificate bundle in PEM format")
		cmd.Flags().Bool(flagInsecureSkipTLSVerify, viper.GetBool(flagInsecureSkipTLSVerify), "skip TLS verification (not recommended!)")
		cmd.Flags().Bool(flagDebug, false, "enable debugging and profiling features")
		cmd.Flags().String(flagLogLevel, viper.GetString(flagLogLevel), "logging level [panic, fatal, error, warn, info, debug]")
		cmd.Flags().Int(backend.FlagEventdWorkers, viper.GetInt(backend.FlagEventdWorkers), "number of workers spawned for processing incoming events")
		cmd.Flags().Int(backend.FlagEventdBufferSize, viper.GetInt(backend.FlagEventdBufferSize), "number of incoming events that can be buffered")
		cmd.Flags().Int(backend.FlagKeepalivedWorkers, viper.GetInt(backend.FlagKeepalivedWorkers), "number of workers spawned for processing incoming keepalives")
		cmd.Flags().Int(backend.FlagKeepalivedBufferSize, viper.GetInt(backend.FlagKeepalivedBufferSize), "number of incoming keepalives that can be buffered")
		cmd.Flags().Int(backend.FlagPipelinedWorkers, viper.GetInt(backend.FlagPipelinedWorkers), "number of workers spawned for handling events through the event pipeline")
		cmd.Flags().Int(backend.FlagPipelinedBufferSize, viper.GetInt(backend.FlagPipelinedBufferSize), "number of events to handle that can be buffered")
		cmd.Flags().Int(backend.FlagAgentWriteTimeout, viper.GetInt(backend.FlagAgentWriteTimeout), "timeout in seconds for agent writes")
		cmd.Flags().String(backend.FlagJWTPrivateKeyFile, viper.GetString(backend.FlagJWTPrivateKeyFile), "path to the PEM-encoded private key to use to sign JWTs")
		cmd.Flags().String(backend.FlagJWTPublicKeyFile, viper.GetString(backend.FlagJWTPublicKeyFile), "path to the PEM-encoded public key to use to verify JWT signatures")

		// Etcd server flags
		cmd.Flags().StringSlice(flagEtcdPeerURLs, viper.GetStringSlice(flagEtcdPeerURLs), "list of URLs to listen on for peer traffic")
		_ = cmd.Flags().SetAnnotation(flagEtcdPeerURLs, "categories", []string{"store"})
		cmd.Flags().String(flagEtcdInitialCluster, viper.GetString(flagEtcdInitialCluster), "initial cluster configuration for bootstrapping")
		_ = cmd.Flags().SetAnnotation(flagEtcdInitialCluster, "categories", []string{"store"})
		cmd.Flags().StringSlice(flagEtcdInitialAdvertisePeerURLs, viper.GetStringSlice(flagEtcdInitialAdvertisePeerURLs), "list of this member's peer URLs to advertise to the rest of the cluster")
		_ = cmd.Flags().SetAnnotation(flagEtcdInitialAdvertisePeerURLs, "categories", []string{"store"})
		cmd.Flags().String(flagEtcdInitialClusterState, viper.GetString(flagEtcdInitialClusterState), "initial cluster state (\"new\" or \"existing\")")
		_ = cmd.Flags().SetAnnotation(flagEtcdInitialClusterState, "categories", []string{"store"})
		cmd.Flags().String(flagEtcdDiscovery, viper.GetString(flagEtcdDiscovery), "discovery URL used to bootstrap the cluster")
		_ = cmd.Flags().SetAnnotation(flagEtcdDiscovery, "categories", []string{"store"})
		cmd.Flags().String(flagEtcdDiscoverySrv, viper.GetString(flagEtcdDiscoverySrv), "DNS SRV record used to bootstrap the cluster")
		_ = cmd.Flags().SetAnnotation(flagEtcdDiscoverySrv, "categories", []string{"store"})
		cmd.Flags().String(flagEtcdInitialClusterToken, viper.GetString(flagEtcdInitialClusterToken), "initial cluster token for the etcd cluster during bootstrap")
		_ = cmd.Flags().SetAnnotation(flagEtcdInitialClusterToken, "categories", []string{"store"})
		cmd.Flags().StringSlice(flagEtcdListenClientURLs, viper.GetStringSlice(flagEtcdListenClientURLs), "list of etcd client URLs to listen on")
		_ = cmd.Flags().SetAnnotation(flagEtcdListenClientURLs, "categories", []string{"store"})
		cmd.Flags().Bool(flagNoEmbedEtcd, viper.GetBool(flagNoEmbedEtcd), "don't embed etcd, use external etcd instead")
		_ = cmd.Flags().SetAnnotation(flagNoEmbedEtcd, "categories", []string{"store"})
		cmd.Flags().Int64(flagEtcdQuotaBackendBytes, viper.GetInt64(flagEtcdQuotaBackendBytes), "maximum etcd database size in bytes (use with caution)")
		_ = cmd.Flags().SetAnnotation(flagEtcdQuotaBackendBytes, "categories", []string{"store"})
		cmd.Flags().Uint(flagEtcdHeartbeatInterval, viper.GetUint(flagEtcdHeartbeatInterval), "interval in ms with which the etcd leader will notify followers that it is still the leader")
		_ = cmd.Flags().SetAnnotation(flagEtcdHeartbeatInterval, "categories", []string{"store"})
		cmd.Flags().Uint(flagEtcdElectionTimeout, viper.GetUint(flagEtcdElectionTimeout), "time in ms a follower node will go without hearing a heartbeat before attempting to become leader itself")
		_ = cmd.Flags().SetAnnotation(flagEtcdElectionTimeout, "categories", []string{"store"})

		// Etcd server TLS flags
		cmd.Flags().String(flagEtcdPeerCertFile, viper.GetString(flagEtcdPeerCertFile), "path to the peer server TLS cert file")
		_ = cmd.Flags().SetAnnotation(flagEtcdPeerCertFile, "categories", []string{"store"})
		cmd.Flags().String(flagEtcdPeerKeyFile, viper.GetString(flagEtcdPeerKeyFile), "path to the peer server TLS key file")
		_ = cmd.Flags().SetAnnotation(flagEtcdPeerKeyFile, "categories", []string{"store"})
		cmd.Flags().Bool(flagEtcdPeerClientCertAuth, viper.GetBool(flagEtcdPeerClientCertAuth), "enable peer client cert authentication")
		_ = cmd.Flags().SetAnnotation(flagEtcdPeerClientCertAuth, "categories", []string{"store"})
		cmd.Flags().String(flagEtcdPeerTrustedCAFile, viper.GetString(flagEtcdPeerTrustedCAFile), "path to the peer server TLS trusted CA file")
		_ = cmd.Flags().SetAnnotation(flagEtcdPeerTrustedCAFile, "categories", []string{"store"})
		cmd.Flags().String(flagEtcdNodeName, viper.GetString(flagEtcdNodeName), "name for this etcd node")
		_ = cmd.Flags().SetAnnotation(flagEtcdNodeName, "categories", []string{"store"})
	}

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

	// Use our custom template for the start command
	cobra.AddTemplateFunc("categoryFlags", categoryFlags)
	cmd.SetUsageTemplate(startUsageTemplate)

	return nil
}

func categoryFlags(category string, flags *pflag.FlagSet) *pflag.FlagSet {
	flagSet := pflag.NewFlagSet(category, pflag.ContinueOnError)

	flags.VisitAll(func(flag *pflag.Flag) {
		if categories, ok := flag.Annotations["categories"]; ok {
			if stringsutil.InArray(category, categories) {
				flagSet.AddFlag(flag)
			}
		} else if category == "" {
			// If no category was specified, return all flags without a category
			flagSet.AddFlag(flag)
		}
	})

	return flagSet
}
