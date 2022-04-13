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
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/sensu/sensu-go/backend/apid/middlewares"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/util/path"
	stringsutil "github.com/sensu/sensu-go/util/strings"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

// The DeprecateDashboardFlags is used to mark usage dashboard daemon flags
// as deprecated.
var DeprecateDashboardFlags = true

var (
	annotations               map[string]string
	labels                    map[string]string
	configFileDefaultLocation = filepath.Join(path.SystemConfigDir(), "backend.yml")
)

const (
	environmentPrefix = "sensu_backend"

	// Flag constants
	flagConfigFile            = "config-file"
	flagAgentHost             = "agent-host"
	flagAgentPort             = "agent-port"
	flagAPIListenAddress      = "api-listen-address"
	flagAPIRequestLimit       = "api-request-limit"
	flagAPIURL                = "api-url"
	flagAPIWriteTimeout       = "api-write-timeout"
	flagAssetsRateLimit       = "assets-rate-limit"
	flagAssetsBurstLimit      = "assets-burst-limit"
	flagDashboardHost         = "dashboard-host"
	flagDashboardPort         = "dashboard-port"
	flagDashboardCertFile     = "dashboard-cert-file"
	flagDashboardKeyFile      = "dashboard-key-file"
	flagDashboardWriteTimeout = "dashboard-write-timeout"
	flagDeregistrationHandler = "deregistration-handler"
	flagCacheDir              = "cache-dir"
	flagStateDir              = "state-dir"
	flagCertFile              = "cert-file"
	flagKeyFile               = "key-file"
	flagTrustedCAFile         = "trusted-ca-file"
	flagInsecureSkipTLSVerify = "insecure-skip-tls-verify"
	flagDebug                 = "debug"
	flagLogLevel              = "log-level"
	flagLabels                = "labels"
	flagAnnotations           = "annotations"

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
	flagEtcdLogLevel                 = "etcd-log-level"
	flagEtcdClientLogLevel           = "etcd-client-log-level"

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

	// Etcd Client Auth Env vars
	envEtcdClientUsername = "etcd-client-username"
	envEtcdClientPassword = "etcd-client-password"

	// Etcd unsafe constants
	flagEtcdUnsafeNoFsync = "etcd-unsafe-no-fsync"

	// Metric logging flags
	flagDisablePlatformMetrics         = "disable-platform-metrics"
	flagPlatformMetricsLoggingInterval = "platform-metrics-logging-interval"
	flagPlatformMetricsLogFile         = "platform-metrics-log-file"

	// flagEventLogBufferSize indicates the size of the events buffer
	flagEventLogBufferSize = "event-log-buffer-size"

	// flagEventLogBufferWait indicates the full buffer wait time
	flagEventLogBufferWait = "event-log-buffer-wait"

	// flagEventLogFile indicates the path to the event log file
	flagEventLogFile = "event-log-file"

	// flagEventLogParallelEncoders used to indicate parallel encoders should be used for event logging
	flagEventLogParallelEncoders = "event-log-parallel-encoders"

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

var (
	// platform metric logging defaults
	defaultDisablePlatformMetrics         = false
	defaultPlatformMetricsLoggingInterval = 60 * time.Second
	defaultPlatformMetricsLogFile         = filepath.Join(path.SystemLogDir(), "backend-stats.log")
)

// InitializeFunc represents the signature of an initialization function, used
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
				APIRequestLimit:       viper.GetInt64(flagAPIRequestLimit),
				APIURL:                viper.GetString(flagAPIURL),
				APIWriteTimeout:       viper.GetDuration(flagAPIWriteTimeout),
				AssetsRateLimit:       rate.Limit(viper.GetFloat64(flagAssetsRateLimit)),
				AssetsBurstLimit:      viper.GetInt(flagAssetsBurstLimit),
				DashboardHost:         viper.GetString(flagDashboardHost),
				DashboardPort:         viper.GetInt(flagDashboardPort),
				DashboardTLSCertFile:  viper.GetString(flagDashboardCertFile),
				DashboardTLSKeyFile:   viper.GetString(flagDashboardKeyFile),
				DashboardWriteTimeout: viper.GetDuration(flagDashboardWriteTimeout),
				DeregistrationHandler: viper.GetString(flagDeregistrationHandler),
				CacheDir:              viper.GetString(flagCacheDir),
				StateDir:              viper.GetString(flagStateDir),

				EtcdAdvertiseClientURLs:        viper.GetStringSlice(flagEtcdAdvertiseClientURLs),
				EtcdListenClientURLs:           viper.GetStringSlice(flagEtcdListenClientURLs),
				EtcdClientURLs:                 fallbackStringSlice(flagEtcdClientURLs, flagEtcdAdvertiseClientURLs),
				EtcdListenPeerURLs:             viper.GetStringSlice(flagEtcdPeerURLs),
				EtcdInitialCluster:             initialCluster,
				EtcdInitialClusterState:        viper.GetString(flagEtcdInitialClusterState),
				EtcdDiscovery:                  etcdDiscovery,
				EtcdDiscoverySrv:               SrvDiscovery,
				EtcdInitialAdvertisePeerURLs:   viper.GetStringSlice(flagEtcdInitialAdvertisePeerURLs),
				EtcdInitialClusterToken:        viper.GetString(flagEtcdInitialClusterToken),
				EtcdName:                       viper.GetString(flagEtcdNodeName),
				EtcdCipherSuites:               viper.GetStringSlice(flagEtcdCipherSuites),
				EtcdQuotaBackendBytes:          viper.GetInt64(flagEtcdQuotaBackendBytes),
				EtcdMaxRequestBytes:            viper.GetUint(flagEtcdMaxRequestBytes),
				EtcdHeartbeatInterval:          viper.GetUint(flagEtcdHeartbeatInterval),
				EtcdElectionTimeout:            viper.GetUint(flagEtcdElectionTimeout),
				EtcdLogLevel:                   viper.GetString(flagEtcdLogLevel),
				EtcdClientLogLevel:             viper.GetString(flagEtcdClientLogLevel),
				EtcdClientUsername:             viper.GetString(envEtcdClientUsername),
				EtcdClientPassword:             viper.GetString(envEtcdClientPassword),
				EtcdUnsafeNoFsync:              viper.GetBool(flagEtcdUnsafeNoFsync),
				NoEmbedEtcd:                    viper.GetBool(flagNoEmbedEtcd),
				Labels:                         viper.GetStringMapString(flagLabels),
				Annotations:                    viper.GetStringMapString(flagAnnotations),
				DisablePlatformMetrics:         viper.GetBool(flagDisablePlatformMetrics),
				PlatformMetricsLoggingInterval: viper.GetDuration(flagPlatformMetricsLoggingInterval),
				PlatformMetricsLogFile:         viper.GetString(flagPlatformMetricsLogFile),
				EventLogBufferSize:             viper.GetInt(flagEventLogBufferSize),
				EventLogBufferWait:             viper.GetDuration(flagEventLogBufferWait),
				EventLogFile:                   viper.GetString(flagEventLogFile),
				EventLogParallelEncoders:       viper.GetBool(flagEventLogParallelEncoders),
			}

			if flag := cmd.Flags().Lookup(flagLabels); flag != nil && flag.Changed {
				cfg.Labels = labels
			}
			if flag := cmd.Flags().Lookup(flagAnnotations); flag != nil && flag.Changed {
				cfg.Annotations = annotations
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

			if cf, kf := len(cfg.DashboardTLSCertFile) == 0, len(cfg.DashboardTLSKeyFile) == 0; cf != kf {
				return fmt.Errorf(
					"dashboard tls configuration error, both flags --%s and --%s are required",
					flagDashboardCertFile, flagDashboardKeyFile,
				)
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

			// Etcd log level
			if cfg.EtcdLogLevel == "" {
				switch level {
				case logrus.TraceLevel:
					cfg.EtcdLogLevel = "debug"
				case logrus.WarnLevel:
					cfg.EtcdLogLevel = "warn"
				default:
					cfg.EtcdLogLevel = level.String()
				}
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
					runtime.SetBlockProfileRate(1)
					log.Println(http.ListenAndServe("127.0.0.1:6060", nil))
				}()
			}
			return sensuBackend.RunWithInitializer(initialize)
		},
	}

	setupErr = handleConfig(cmd, os.Args[1:], true)

	return cmd
}

func handleConfig(cmd *cobra.Command, arguments []string, server bool) error {
	configFlags := flagSet(server)
	_ = configFlags.Parse(arguments)

	// Get the given config file path via flag
	configFilePath, _ := configFlags.GetString(flagConfigFile)

	// Get the environment variable value if no config file was provided via the flag
	if configFilePath == "" {
		environmentConfigFile := fmt.Sprintf("%s_%s", environmentPrefix, flagConfigFile)
		environmentConfigFile = strings.ToUpper(environmentConfigFile)
		environmentConfigFile = strings.Replace(environmentConfigFile, "-", "_", -1)
		configFilePath = os.Getenv(environmentConfigFile)
	}

	// Use the default config path as a fallback if no config file was provided
	// via the flag or the environment variable
	configFilePathIsDefined := true
	if configFilePath == "" {
		configFilePathIsDefined = false
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
		viper.SetDefault(flagAPIRequestLimit, middlewares.MaxBytesLimit)
		viper.SetDefault(flagAPIURL, "http://localhost:8080")
		viper.SetDefault(flagAPIWriteTimeout, "15s")
		viper.SetDefault(flagAssetsRateLimit, asset.DefaultAssetsRateLimit)
		viper.SetDefault(flagAssetsBurstLimit, asset.DefaultAssetsBurstLimit)
		viper.SetDefault(flagDashboardHost, "[::]")
		viper.SetDefault(flagDashboardPort, 3000)
		viper.SetDefault(flagDashboardCertFile, "")
		viper.SetDefault(flagDashboardKeyFile, "")
		viper.SetDefault(flagDashboardWriteTimeout, "15s")
		viper.SetDefault(flagDeregistrationHandler, "")
		viper.SetDefault(flagCacheDir, path.SystemCacheDir("sensu-backend"))
		viper.SetDefault(flagStateDir, path.SystemDataDir("sensu-backend"))
		viper.SetDefault(flagCertFile, "")
		viper.SetDefault(flagKeyFile, "")
		viper.SetDefault(flagTrustedCAFile, "")
		viper.SetDefault(flagInsecureSkipTLSVerify, false)
		viper.SetDefault(flagLogLevel, "warn")
		viper.SetDefault(backend.FlagEventdWorkers, 100)
		viper.SetDefault(backend.FlagEventdBufferSize, 1000)
		viper.SetDefault(backend.FlagKeepalivedWorkers, 100)
		viper.SetDefault(backend.FlagKeepalivedBufferSize, 1000)
		viper.SetDefault(backend.FlagPipelinedWorkers, 100)
		viper.SetDefault(backend.FlagPipelinedBufferSize, 1000)
		viper.SetDefault(backend.FlagAgentWriteTimeout, 15)
		viper.SetDefault(flagDisablePlatformMetrics, defaultDisablePlatformMetrics)
		viper.SetDefault(flagPlatformMetricsLoggingInterval, defaultPlatformMetricsLoggingInterval)
		viper.SetDefault(flagPlatformMetricsLogFile, defaultPlatformMetricsLogFile)
		viper.SetDefault(flagEventLogBufferWait, 10*time.Millisecond)
		viper.SetDefault(flagEventLogBufferSize, 100000)
		viper.SetDefault(flagEventLogFile, "")
		viper.SetDefault(flagEventLogParallelEncoders, false)
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
	viper.SetDefault(flagEtcdClientLogLevel, etcd.DefaultClientLogLevel)

	if server {
		viper.SetDefault(flagNoEmbedEtcd, false)
	}

	// Merge in flag set so that it appears in command usage
	flags := flagSet(server)
	cmd.Flags().AddFlagSet(flags)

	// Load the configuration file but only error out if flagConfigFile is used
	if err := viper.ReadInConfig(); err != nil && configFilePathIsDefined {
		return err
	}

	viper.SetEnvPrefix(environmentPrefix)
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

func flagSet(server bool) *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("start", pflag.ContinueOnError)

	// Config flag
	configFileDescription := fmt.Sprintf("path to sensu-backend config file (default %q)", configFileDefaultLocation)
	flagSet.StringP(flagConfigFile, "c", "", configFileDescription)

	// Etcd client/server flags
	flagSet.StringSlice(flagEtcdCipherSuites, nil, "list of ciphers to use for etcd TLS configuration")
	_ = flagSet.SetAnnotation(flagEtcdCipherSuites, "categories", []string{"store"})
	flagSet.String(flagEtcdClientLogLevel, viper.GetString(flagEtcdClientLogLevel), "etcd client logging level [panic, fatal, error, warn, info, debug]")
	_ = flagSet.SetAnnotation(flagEtcdClientLogLevel, "categories", []string{"store"})

	// This one is really only a server flag, but because we lacked
	// --etcd-client-urls until recently, it's used as a fallback.
	flagSet.StringSlice(flagEtcdAdvertiseClientURLs, viper.GetStringSlice(flagEtcdAdvertiseClientURLs), "list of this member's client URLs to advertise to clients")
	_ = flagSet.SetAnnotation(flagEtcdAdvertiseClientURLs, "categories", []string{"store"})

	flagSet.Uint(flagEtcdMaxRequestBytes, viper.GetUint(flagEtcdMaxRequestBytes), "maximum etcd request size in bytes (use with caution)")
	_ = flagSet.SetAnnotation(flagEtcdMaxRequestBytes, "categories", []string{"store"})

	// Etcd client/server TLS flags
	flagSet.String(flagEtcdCertFile, viper.GetString(flagEtcdCertFile), "path to the client server TLS cert file")
	_ = flagSet.SetAnnotation(flagEtcdCertFile, "categories", []string{"store"})
	flagSet.String(flagEtcdKeyFile, viper.GetString(flagEtcdKeyFile), "path to the client server TLS key file")
	_ = flagSet.SetAnnotation(flagEtcdKeyFile, "categories", []string{"store"})
	flagSet.Bool(flagEtcdClientCertAuth, viper.GetBool(flagEtcdClientCertAuth), "enable client cert authentication")
	_ = flagSet.SetAnnotation(flagEtcdClientCertAuth, "categories", []string{"store"})
	flagSet.String(flagEtcdTrustedCAFile, viper.GetString(flagEtcdTrustedCAFile), "path to the client server TLS trusted CA cert file")
	_ = flagSet.SetAnnotation(flagEtcdTrustedCAFile, "categories", []string{"store"})
	flagSet.String(flagEtcdClientURLs, viper.GetString(flagEtcdClientURLs), "client URLs to use when operating as an etcd client")
	_ = flagSet.SetAnnotation(flagEtcdClientURLs, "categories", []string{"store"})

	if server {
		// Main Flags
		flagSet.String(flagAgentHost, viper.GetString(flagAgentHost), "agent listener host")
		flagSet.Int(flagAgentPort, viper.GetInt(flagAgentPort), "agent listener port")
		flagSet.String(flagAPIListenAddress, viper.GetString(flagAPIListenAddress), "address to listen on for api traffic")
		flagSet.Int64(flagAPIRequestLimit, viper.GetInt64(flagAPIRequestLimit), "maximum API request body size, in bytes")
		flagSet.String(flagAPIURL, viper.GetString(flagAPIURL), "url of the api to connect to")
		flagSet.Duration(flagAPIWriteTimeout, viper.GetDuration(flagAPIWriteTimeout), "maximum duration before timing out writes of responses")
		flagSet.Float64(flagAssetsRateLimit, viper.GetFloat64(flagAssetsRateLimit), "maximum number of assets fetched per second")
		flagSet.Int(flagAssetsBurstLimit, viper.GetInt(flagAssetsBurstLimit), "asset fetch burst limit")
		flagSet.String(flagDashboardHost, viper.GetString(flagDashboardHost), "dashboard listener host")
		flagSet.Int(flagDashboardPort, viper.GetInt(flagDashboardPort), "dashboard listener port")
		flagSet.String(flagDashboardCertFile, viper.GetString(flagDashboardCertFile), "dashboard TLS certificate in PEM format")
		flagSet.String(flagDashboardKeyFile, viper.GetString(flagDashboardKeyFile), "dashboard TLS certificate key in PEM format")
		flagSet.Duration(flagDashboardWriteTimeout, viper.GetDuration(flagDashboardWriteTimeout), "maximum duration before timing out writes of responses")
		flagSet.String(flagDeregistrationHandler, viper.GetString(flagDeregistrationHandler), "default deregistration handler")
		flagSet.String(flagCacheDir, viper.GetString(flagCacheDir), "path to store cached data")
		flagSet.StringP(flagStateDir, "d", viper.GetString(flagStateDir), "path to sensu state storage")
		flagSet.String(flagCertFile, viper.GetString(flagCertFile), "TLS certificate in PEM format")
		flagSet.String(flagKeyFile, viper.GetString(flagKeyFile), "TLS certificate key in PEM format")
		flagSet.String(flagTrustedCAFile, viper.GetString(flagTrustedCAFile), "TLS CA certificate bundle in PEM format")
		flagSet.Bool(flagInsecureSkipTLSVerify, viper.GetBool(flagInsecureSkipTLSVerify), "skip TLS verification (not recommended!)")
		flagSet.Bool(flagDebug, false, "enable debugging and profiling features")
		flagSet.String(flagLogLevel, viper.GetString(flagLogLevel), "logging level [panic, fatal, error, warn, info, debug, trace]")
		flagSet.Int(backend.FlagEventdWorkers, viper.GetInt(backend.FlagEventdWorkers), "number of workers spawned for processing incoming events")
		flagSet.Int(backend.FlagEventdBufferSize, viper.GetInt(backend.FlagEventdBufferSize), "number of incoming events that can be buffered")
		flagSet.Int(backend.FlagKeepalivedWorkers, viper.GetInt(backend.FlagKeepalivedWorkers), "number of workers spawned for processing incoming keepalives")
		flagSet.Int(backend.FlagKeepalivedBufferSize, viper.GetInt(backend.FlagKeepalivedBufferSize), "number of incoming keepalives that can be buffered")
		flagSet.Int(backend.FlagPipelinedWorkers, viper.GetInt(backend.FlagPipelinedWorkers), "number of workers spawned for handling events through the event pipeline")
		flagSet.Int(backend.FlagPipelinedBufferSize, viper.GetInt(backend.FlagPipelinedBufferSize), "number of events to handle that can be buffered")
		flagSet.Int(backend.FlagAgentWriteTimeout, viper.GetInt(backend.FlagAgentWriteTimeout), "timeout in seconds for agent writes")
		flagSet.String(backend.FlagJWTPrivateKeyFile, viper.GetString(backend.FlagJWTPrivateKeyFile), "path to the PEM-encoded private key to use to sign JWTs")
		flagSet.String(backend.FlagJWTPublicKeyFile, viper.GetString(backend.FlagJWTPublicKeyFile), "path to the PEM-encoded public key to use to verify JWT signatures")
		flagSet.StringToStringVar(&labels, flagLabels, nil, "entity labels map")
		flagSet.StringToStringVar(&annotations, flagAnnotations, nil, "entity annotations map")
		flagSet.Bool(flagDisablePlatformMetrics, viper.GetBool(flagDisablePlatformMetrics), "disable platform metrics logging")
		flagSet.Duration(flagPlatformMetricsLoggingInterval, viper.GetDuration(flagPlatformMetricsLoggingInterval), "platform metrics logging interval")
		flagSet.String(flagPlatformMetricsLogFile, viper.GetString(flagPlatformMetricsLogFile), "platform metrics log file path")

		// Etcd server flags
		flagSet.StringSlice(flagEtcdPeerURLs, viper.GetStringSlice(flagEtcdPeerURLs), "list of URLs to listen on for peer traffic")
		_ = flagSet.SetAnnotation(flagEtcdPeerURLs, "categories", []string{"store"})
		flagSet.String(flagEtcdInitialCluster, viper.GetString(flagEtcdInitialCluster), "initial cluster configuration for bootstrapping")
		_ = flagSet.SetAnnotation(flagEtcdInitialCluster, "categories", []string{"store"})
		flagSet.StringSlice(flagEtcdInitialAdvertisePeerURLs, viper.GetStringSlice(flagEtcdInitialAdvertisePeerURLs), "list of this member's peer URLs to advertise to the rest of the cluster")
		_ = flagSet.SetAnnotation(flagEtcdInitialAdvertisePeerURLs, "categories", []string{"store"})
		flagSet.String(flagEtcdInitialClusterState, viper.GetString(flagEtcdInitialClusterState), "initial cluster state (\"new\" or \"existing\")")
		_ = flagSet.SetAnnotation(flagEtcdInitialClusterState, "categories", []string{"store"})
		flagSet.String(flagEtcdDiscovery, viper.GetString(flagEtcdDiscovery), "discovery URL used to bootstrap the cluster")
		_ = flagSet.SetAnnotation(flagEtcdDiscovery, "categories", []string{"store"})
		flagSet.String(flagEtcdDiscoverySrv, viper.GetString(flagEtcdDiscoverySrv), "DNS SRV record used to bootstrap the cluster")
		_ = flagSet.SetAnnotation(flagEtcdDiscoverySrv, "categories", []string{"store"})
		flagSet.String(flagEtcdInitialClusterToken, viper.GetString(flagEtcdInitialClusterToken), "initial cluster token for the etcd cluster during bootstrap")
		_ = flagSet.SetAnnotation(flagEtcdInitialClusterToken, "categories", []string{"store"})
		flagSet.StringSlice(flagEtcdListenClientURLs, viper.GetStringSlice(flagEtcdListenClientURLs), "list of etcd client URLs to listen on")
		_ = flagSet.SetAnnotation(flagEtcdListenClientURLs, "categories", []string{"store"})
		flagSet.Bool(flagNoEmbedEtcd, viper.GetBool(flagNoEmbedEtcd), "don't embed etcd, use external etcd instead")
		_ = flagSet.SetAnnotation(flagNoEmbedEtcd, "categories", []string{"store"})
		flagSet.Int64(flagEtcdQuotaBackendBytes, viper.GetInt64(flagEtcdQuotaBackendBytes), "maximum etcd database size in bytes (use with caution)")
		_ = flagSet.SetAnnotation(flagEtcdQuotaBackendBytes, "categories", []string{"store"})
		flagSet.Uint(flagEtcdHeartbeatInterval, viper.GetUint(flagEtcdHeartbeatInterval), "interval in ms with which the etcd leader will notify followers that it is still the leader")
		_ = flagSet.SetAnnotation(flagEtcdHeartbeatInterval, "categories", []string{"store"})
		flagSet.Uint(flagEtcdElectionTimeout, viper.GetUint(flagEtcdElectionTimeout), "time in ms a follower node will go without hearing a heartbeat before attempting to become leader itself")
		_ = flagSet.SetAnnotation(flagEtcdElectionTimeout, "categories", []string{"store"})
		flagSet.String(flagEtcdLogLevel, viper.GetString(flagEtcdLogLevel), "etcd server logging level [panic, fatal, error, warn, info, debug]")
		_ = flagSet.SetAnnotation(flagEtcdLogLevel, "categories", []string{"store"})

		// Etcd server TLS flags
		flagSet.String(flagEtcdPeerCertFile, viper.GetString(flagEtcdPeerCertFile), "path to the peer server TLS cert file")
		_ = flagSet.SetAnnotation(flagEtcdPeerCertFile, "categories", []string{"store"})
		flagSet.String(flagEtcdPeerKeyFile, viper.GetString(flagEtcdPeerKeyFile), "path to the peer server TLS key file")
		_ = flagSet.SetAnnotation(flagEtcdPeerKeyFile, "categories", []string{"store"})
		flagSet.Bool(flagEtcdPeerClientCertAuth, viper.GetBool(flagEtcdPeerClientCertAuth), "enable peer client cert authentication")
		_ = flagSet.SetAnnotation(flagEtcdPeerClientCertAuth, "categories", []string{"store"})
		flagSet.String(flagEtcdPeerTrustedCAFile, viper.GetString(flagEtcdPeerTrustedCAFile), "path to the peer server TLS trusted CA file")
		_ = flagSet.SetAnnotation(flagEtcdPeerTrustedCAFile, "categories", []string{"store"})
		flagSet.String(flagEtcdNodeName, viper.GetString(flagEtcdNodeName), "name for this etcd node")
		_ = flagSet.SetAnnotation(flagEtcdNodeName, "categories", []string{"store"})

		_ = flagSet.String(flagEventLogFile, "", "path to the event log file")
		_ = flagSet.Bool(flagEventLogParallelEncoders, false, "use parallel JSON encoding for the event log")

		// Etcd server unsafe flags
		_ = flagSet.Bool(flagEtcdUnsafeNoFsync, false, "disables fsync, unsafe, may cause data loss")

		// Use a default value of 100,000 messages for the buffer. A serialized event
		// takes a minimum of around 1300 bytes, so once full the buffer ring could
		// require about 130MB of memory.
		_ = flagSet.Int(flagEventLogBufferSize, 100000, "buffer size of the event logger")

		// Use a default value of 10ms for the full buffer wait time. When the buffer
		// is full, the logger will wait for the writer to consume events from the buffer.
		// This helps reduce event data loss but comes at the cost of event back-pressure
		// for the backend and its agent sessions. If the buffer fills and the wait time
		// is too low, it will dicard too many events. If the wait time is too high,
		// event back-pressure could stop the backend and its agent sessions from
		// producing and processing new events and possibly lead to a crash.
		_ = flagSet.String(flagEventLogBufferWait, "10ms", "full buffer wait time")
	}

	flagSet.SetOutput(ioutil.Discard)

	return flagSet
}
