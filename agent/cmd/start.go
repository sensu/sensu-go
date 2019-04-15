package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/sensu/sensu-go/agent"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/path"
	"github.com/sensu/sensu-go/util/url"
	"github.com/sensu/sensu-go/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

var (
	logger *logrus.Entry
)

const (
	// DefaultBackendPort specifies the default port to use when a port is not
	// specified in backend urls
	DefaultBackendPort = "8081"

	flagAgentName             = "name"
	flagAPIHost               = "api-host"
	flagAPIPort               = "api-port"
	flagBackendURL            = "backend-url"
	flagCacheDir              = "cache-dir"
	flagConfigFile            = "config-file"
	flagDeregister            = "deregister"
	flagDeregistrationHandler = "deregistration-handler"
	flagEventsRateLimit       = "events-rate-limit"
	flagEventsBurstLimit      = "events-burst-limit"
	flagKeepaliveInterval     = "keepalive-interval"
	flagKeepaliveTimeout      = "keepalive-timeout"
	flagNamespace             = "namespace"
	flagPassword              = "password"
	flagRedact                = "redact"
	flagSocketHost            = "socket-host"
	flagSocketPort            = "socket-port"
	flagStatsdDisable         = "statsd-disable"
	flagStatsdEventHandlers   = "statsd-event-handlers"
	flagStatsdFlushInterval   = "statsd-flush-interval"
	flagStatsdMetricsHost     = "statsd-metrics-host"
	flagStatsdMetricsPort     = "statsd-metrics-port"
	flagSubscriptions         = "subscriptions"
	flagUser                  = "user"
	flagDisableAPI            = "disable-api"
	flagDisableSockets        = "disable-sockets"
	flagLogLevel              = "log-level"
	flagLabels                = "labels"
	flagAnnotations           = "annotations"

	// TLS flags
	flagTrustedCAFile         = "trusted-ca-file"
	flagInsecureSkipTLSVerify = "insecure-skip-tls-verify"

	deprecatedFlagAgentID = "id"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logger = logrus.WithFields(logrus.Fields{
		"component": "cmd",
	})

	rootCmd.AddCommand(newVersionCommand())
	rootCmd.AddCommand(newStartCommand())

	viper.SetEnvPrefix("sensu")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

func newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the sensu-agent version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("sensu-agent version %s, build %s, built %s\n",
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
		Use:           "start",
		Short:         "start the sensu agent",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			if setupErr != nil {
				return setupErr
			}
			level, err := logrus.ParseLevel(viper.GetString(flagLogLevel))
			if err != nil {
				return err
			}
			logrus.SetLevel(level)

			cfg := agent.NewConfig()
			cfg.API.Host = viper.GetString(flagAPIHost)
			cfg.API.Port = viper.GetInt(flagAPIPort)
			cfg.CacheDir = viper.GetString(flagCacheDir)
			cfg.Deregister = viper.GetBool(flagDeregister)
			cfg.DeregistrationHandler = viper.GetString(flagDeregistrationHandler)
			cfg.EventsAPIRateLimit = rate.Limit(viper.GetFloat64(flagEventsRateLimit))
			cfg.EventsAPIBurstLimit = viper.GetInt(flagEventsBurstLimit)
			cfg.KeepaliveInterval = uint32(viper.GetInt(flagKeepaliveInterval))
			cfg.KeepaliveTimeout = uint32(viper.GetInt(flagKeepaliveTimeout))
			cfg.Namespace = viper.GetString(flagNamespace)
			cfg.Password = viper.GetString(flagPassword)
			cfg.Socket.Host = viper.GetString(flagSocketHost)
			cfg.Socket.Port = viper.GetInt(flagSocketPort)
			cfg.StatsdServer.Disable = viper.GetBool(flagStatsdDisable)
			cfg.StatsdServer.FlushInterval = viper.GetInt(flagStatsdFlushInterval)
			cfg.StatsdServer.Host = viper.GetString(flagStatsdMetricsHost)
			cfg.StatsdServer.Port = viper.GetInt(flagStatsdMetricsPort)
			cfg.StatsdServer.Handlers = viper.GetStringSlice(flagStatsdEventHandlers)
			cfg.Labels = viper.GetStringMapString(flagLabels)
			cfg.Annotations = viper.GetStringMapString(flagAnnotations)
			cfg.User = viper.GetString(flagUser)

			// TLS configuration
			cfg.TLS = &types.TLSOptions{}
			cfg.TLS.TrustedCAFile = viper.GetString(flagTrustedCAFile)
			cfg.TLS.InsecureSkipVerify = viper.GetBool(flagInsecureSkipTLSVerify)

			agentName := viper.GetString(flagAgentName)
			if agentName != "" {
				cfg.AgentName = agentName
			}

			for _, backendURL := range viper.GetStringSlice(flagBackendURL) {
				newURL, err := url.AppendPortIfMissing(backendURL, DefaultBackendPort)
				if err != nil {
					return err
				}
				cfg.BackendURLs = append(cfg.BackendURLs, newURL)
			}

			cfg.Redact = viper.GetStringSlice(flagRedact)
			cfg.Subscriptions = viper.GetStringSlice(flagSubscriptions)

			sensuAgent, err := agent.NewAgent(cfg)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				sig := <-sigs
				logger.Info("signal received: ", sig)
				cancel()
			}()

			if !viper.GetBool(flagDisableAPI) {
				sensuAgent.StartAPI(ctx)
			}

			if !viper.GetBool(flagDisableSockets) {
				// Agent TCP/UDP sockets are deprecated in favor of the agent rest api
				sensuAgent.StartSocketListeners(ctx)
			}

			return sensuAgent.Run(ctx)
		},
	}

	// Set up distinct flagset for handling config file
	configFlagSet := pflag.NewFlagSet("sensu", pflag.ContinueOnError)
	_ = configFlagSet.StringP(flagConfigFile, "c", "", "path to sensu-agent config file")
	configFlagSet.SetOutput(ioutil.Discard)
	_ = configFlagSet.Parse(os.Args[1:])

	// Get the given config file path
	configFile, _ := configFlagSet.GetString(flagConfigFile)
	configFilePath := configFile

	// use the default config path if flagConfigFile was not used
	if configFile == "" {
		configFilePath = filepath.Join(path.SystemConfigDir(), "agent.yml")
	}

	// Configure location of backend configuration
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configFilePath)

	// Flag defaults
	viper.SetDefault(flagAgentName, agent.GetDefaultAgentName())
	viper.SetDefault(flagAPIHost, agent.DefaultAPIHost)
	viper.SetDefault(flagAPIPort, agent.DefaultAPIPort)
	viper.SetDefault(flagBackendURL, []string{agent.DefaultBackendURL})
	viper.SetDefault(flagCacheDir, path.SystemCacheDir("sensu-agent"))
	viper.SetDefault(flagDeregister, false)
	viper.SetDefault(flagDeregistrationHandler, "")
	viper.SetDefault(flagEventsRateLimit, agent.DefaultEventsAPIRateLimit)
	viper.SetDefault(flagEventsBurstLimit, agent.DefaultEventsAPIBurstLimit)
	viper.SetDefault(flagKeepaliveInterval, agent.DefaultKeepaliveInterval)
	viper.SetDefault(flagKeepaliveTimeout, types.DefaultKeepaliveTimeout)
	viper.SetDefault(flagNamespace, agent.DefaultNamespace)
	viper.SetDefault(flagPassword, agent.DefaultPassword)
	viper.SetDefault(flagRedact, types.DefaultRedactFields)
	viper.SetDefault(flagSocketHost, agent.DefaultSocketHost)
	viper.SetDefault(flagSocketPort, agent.DefaultSocketPort)
	viper.SetDefault(flagStatsdDisable, agent.DefaultStatsdDisable)
	viper.SetDefault(flagStatsdFlushInterval, agent.DefaultStatsdFlushInterval)
	viper.SetDefault(flagStatsdMetricsHost, agent.DefaultStatsdMetricsHost)
	viper.SetDefault(flagStatsdMetricsPort, agent.DefaultStatsdMetricsPort)
	viper.SetDefault(flagStatsdEventHandlers, []string{})
	viper.SetDefault(flagSubscriptions, []string{})
	viper.SetDefault(flagUser, agent.DefaultUser)
	viper.SetDefault(flagDisableAPI, false)
	viper.SetDefault(flagDisableSockets, false)
	viper.SetDefault(flagTrustedCAFile, "")
	viper.SetDefault(flagInsecureSkipTLSVerify, false)
	viper.SetDefault(flagLogLevel, "warn")

	// Merge in config flag set so that it appears in command usage
	cmd.Flags().AddFlagSet(configFlagSet)

	// Flags
	// Load the configuration file but only error out if flagConfigFile is used
	cmd.Flags().Bool(flagDeregister, viper.GetBool(flagDeregister), "ephemeral agent")
	cmd.Flags().Int(flagAPIPort, viper.GetInt(flagAPIPort), "port the Sensu client HTTP API listens on")
	cmd.Flags().Int(flagKeepaliveInterval, viper.GetInt(flagKeepaliveInterval), "number of seconds to send between keepalive events")
	cmd.Flags().Int(flagSocketPort, viper.GetInt(flagSocketPort), "port the Sensu client socket listens on")
	cmd.Flags().String(flagAgentName, viper.GetString(flagAgentName), "agent name (defaults to hostname)")
	cmd.Flags().String(flagAPIHost, viper.GetString(flagAPIHost), "address to bind the Sensu client HTTP API to")
	cmd.Flags().String(flagCacheDir, viper.GetString(flagCacheDir), "path to store cached data")
	cmd.Flags().String(flagDeregistrationHandler, viper.GetString(flagDeregistrationHandler), "deregistration handler that should process the entity deregistration event.")
	cmd.Flags().Float64(flagEventsRateLimit, viper.GetFloat64(flagEventsRateLimit), "maximum number of events transmitted to the backend through the /events api")
	cmd.Flags().Int(flagEventsBurstLimit, viper.GetInt(flagEventsBurstLimit), "/events api burst limit")
	cmd.Flags().String(flagNamespace, viper.GetString(flagNamespace), "agent namespace")
	cmd.Flags().String(flagPassword, viper.GetString(flagPassword), "agent password")
	cmd.Flags().StringSlice(flagRedact, viper.GetStringSlice(flagRedact), "comma-delimited customized list of fields to redact")
	cmd.Flags().String(flagSocketHost, viper.GetString(flagSocketHost), "address to bind the Sensu client socket to")
	cmd.Flags().Bool(flagStatsdDisable, viper.GetBool(flagStatsdDisable), "disables the statsd listener and metrics server")
	cmd.Flags().StringSlice(flagStatsdEventHandlers, viper.GetStringSlice(flagStatsdEventHandlers), "event handlers for statsd metrics, one per flag")
	cmd.Flags().Int(flagStatsdFlushInterval, viper.GetInt(flagStatsdFlushInterval), "number of seconds between statsd flush")
	cmd.Flags().String(flagStatsdMetricsHost, viper.GetString(flagStatsdMetricsHost), "address used for the statsd metrics server")
	cmd.Flags().Int(flagStatsdMetricsPort, viper.GetInt(flagStatsdMetricsPort), "port used for the statsd metrics server")
	cmd.Flags().StringSlice(flagSubscriptions, viper.GetStringSlice(flagSubscriptions), "comma-delimited list of agent subscriptions")
	cmd.Flags().String(flagUser, viper.GetString(flagUser), "agent user")
	cmd.Flags().StringSlice(flagBackendURL, viper.GetStringSlice(flagBackendURL), "ws/wss URL of Sensu backend server (to specify multiple backends use this flag multiple times)")
	cmd.Flags().Uint32(flagKeepaliveTimeout, uint32(viper.GetInt(flagKeepaliveTimeout)), "number of seconds until agent is considered dead by backend")
	cmd.Flags().Bool(flagDisableAPI, viper.GetBool(flagDisableAPI), "disable the Agent HTTP API")
	cmd.Flags().Bool(flagDisableSockets, viper.GetBool(flagDisableSockets), "disable the Agent TCP and UDP event sockets")
	cmd.Flags().String(flagTrustedCAFile, viper.GetString(flagTrustedCAFile), "TLS CA certificate bundle in PEM format")
	cmd.Flags().Bool(flagInsecureSkipTLSVerify, viper.GetBool(flagInsecureSkipTLSVerify), "skip TLS verification (not recommended!)")
	cmd.Flags().String(flagLogLevel, viper.GetString(flagLogLevel), "logging level [panic, fatal, error, warn, info, debug]")
	cmd.Flags().StringToString(flagLabels, viper.GetStringMapString(flagLabels), "entity labels map")
	cmd.Flags().StringToString(flagAnnotations, viper.GetStringMapString(flagAnnotations), "entity annotations map")

	cmd.Flags().SetNormalizeFunc(aliasNormalizeFunc)

	if err := viper.ReadInConfig(); err != nil && configFile != "" {
		setupErr = err
	}

	deprecatedConfigAttributes()
	viper.RegisterAlias(deprecatedFlagAgentID, flagAgentName)

	return cmd
}

func aliasNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	// Wait until the command-line flags have been parsed
	if !f.Parsed() {
		return pflag.NormalizedName(name)
	}

	switch name {
	case deprecatedFlagAgentID:
		deprecatedFlagMessage(name, flagAgentName)
		name = flagAgentName
	}
	return pflag.NormalizedName(name)
}

// Look up the deprecated attributes in our config file and print a warning
// message if set
func deprecatedConfigAttributes() {
	attributes := map[string]string{
		deprecatedFlagAgentID: flagAgentName,
	}

	for old, new := range attributes {
		if viper.IsSet(old) {
			logger.Warningf(
				"config attribute %s has been deprecated, please use %s instead",
				old, new,
			)
		}
	}
}

func deprecatedFlagMessage(oldFlag, newFlag string) {
	logger.Warningf("flag --%s has been deprecated, please use --%s instead",
		oldFlag, newFlag)
}
