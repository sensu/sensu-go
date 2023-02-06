package cmd

import (
	"context"
	"errors"
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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sensu/sensu-go/backend/apid/middlewares"
	"github.com/sensu/sensu-go/backend/store/postgres"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend"
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
	flagCertFile              = "cert-file"
	flagKeyFile               = "key-file"
	flagTrustedCAFile         = "trusted-ca-file"
	flagInsecureSkipTLSVerify = "insecure-skip-tls-verify"
	flagDebug                 = "debug"
	flagLogLevel              = "log-level"
	flagLabels                = "labels"
	flagAnnotations           = "annotations"
	flagName                  = "name"

	// Postgres store
	flagPGDSN    = "pg-dsn"     // postgresql connection string
	flagPGMaxTPS = "pg-max-tps" // postgresql maximum transactions per second cap

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

Postgresql Store Flags:
{{ $pgcfgflags := categoryFlags "store" .LocalFlags }}{{ $pgcfgflags.FlagUsages | trimTrailingWhitespaces }}

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
type InitializeFunc func(context.Context, postgres.DBI, *backend.Config) (*backend.Backend, error)

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
				Name:                  viper.GetString(flagName),

				Labels:                         viper.GetStringMapString(flagLabels),
				Annotations:                    viper.GetStringMapString(flagAnnotations),
				DisablePlatformMetrics:         viper.GetBool(flagDisablePlatformMetrics),
				PlatformMetricsLoggingInterval: viper.GetDuration(flagPlatformMetricsLoggingInterval),
				PlatformMetricsLogFile:         viper.GetString(flagPlatformMetricsLogFile),
				EventLogBufferSize:             viper.GetInt(flagEventLogBufferSize),
				EventLogBufferWait:             viper.GetDuration(flagEventLogBufferWait),
				EventLogFile:                   viper.GetString(flagEventLogFile),
				EventLogParallelEncoders:       viper.GetBool(flagEventLogParallelEncoders),

				Store: backend.StoreConfig{
					PostgresStore: postgres.Config{
						DSN:    viper.GetString(flagPGDSN),
						MaxTPS: viper.GetInt(flagPGMaxTPS),
					},
				},
			}

			if cfg.CacheDir == "" {
				return errors.New("cache dir not set")
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

			var pgDB *pgxpool.Pool

			ctx, cancel := context.WithCancel(context.Background())

			pgDB, err = newPostgresPool(ctx, cfg.Store.PostgresStore.DSN)
			if err != nil {
				return err
			}
			defer pgDB.Close()

			sensuBackend, err := initialize(ctx, pgDB, cfg)
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
			return sensuBackend.Run(ctx)
		},
	}

	setupErr = handleConfig(cmd, os.Args[1:], true)

	return cmd
}

func newPostgresPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pgxConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// Create the event store, which runs on top of postgres
	db, err := postgres.Open(ctx, pgxConfig, true)
	if err != nil {
		return nil, err
	}
	return db, nil
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
		viper.SetDefault(flagPGMaxTPS, 1000)

		backendName, err := os.Hostname()
		if err != nil {
			// According to `man gethostname`, this should never happen, unless
			// there is a bug in Go's use of gethostname
			panic(err)
		}
		viper.SetDefault(flagName, backendName)
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

	flagSet.String(flagPGDSN, viper.GetString(flagPGDSN), "postgresql store DSN")
	_ = flagSet.SetAnnotation(flagPGDSN, "categories", []string{"store"})

	flagSet.Int(flagPGMaxTPS, viper.GetInt(flagPGMaxTPS), "postgresql max transactions per second")
	_ = flagSet.SetAnnotation(flagPGMaxTPS, "categories", []string{"store"})

	if server {
		// Main Flags
		flagSet.String(flagName, viper.GetString(flagName), "backend name")
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

		_ = flagSet.String(flagEventLogFile, "", "path to the event log file")
		_ = flagSet.Bool(flagEventLogParallelEncoders, false, "use parallel JSON encoding for the event log")

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
