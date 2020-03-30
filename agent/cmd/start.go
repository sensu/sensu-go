package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/sensu/sensu-go/agent"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
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
	annotations map[string]string
	labels      map[string]string
)

const (
	// DefaultBackendPort specifies the default port to use when a port is not
	// specified in backend urls
	DefaultBackendPort = "8081"

	flagAgentName                = "name"
	flagAPIHost                  = "api-host"
	flagAPIPort                  = "api-port"
	flagBackendURL               = "backend-url"
	flagCacheDir                 = "cache-dir"
	flagConfigFile               = "config-file"
	flagDeregister               = "deregister"
	flagDeregistrationHandler    = "deregistration-handler"
	flagDetectCloudProvider      = "detect-cloud-provider"
	flagEventsRateLimit          = "events-rate-limit"
	flagEventsBurstLimit         = "events-burst-limit"
	flagKeepaliveHandlers        = "keepalive-handlers"
	flagKeepaliveInterval        = "keepalive-interval"
	flagKeepaliveWarningTimeout  = "keepalive-warning-timeout"
	flagKeepaliveCriticalTimeout = "keepalive-critical-timeout"
	flagNamespace                = "namespace"
	flagPassword                 = "password"
	flagRedact                   = "redact"
	flagSocketHost               = "socket-host"
	flagSocketPort               = "socket-port"
	flagStatsdDisable            = "statsd-disable"
	flagStatsdEventHandlers      = "statsd-event-handlers"
	flagStatsdFlushInterval      = "statsd-flush-interval"
	flagStatsdMetricsHost        = "statsd-metrics-host"
	flagStatsdMetricsPort        = "statsd-metrics-port"
	flagSubscriptions            = "subscriptions"
	flagUser                     = "user"
	flagDisableAPI               = "disable-api"
	flagDisableAssets            = "disable-assets"
	flagDisableSockets           = "disable-sockets"
	flagLogLevel                 = "log-level"
	flagLabels                   = "labels"
	flagAnnotations              = "annotations"
	flagAllowList                = "allow-list"
	flagBackendHandshakeTimeout  = "backend-handshake-timeout"
	flagBackendHeartbeatInterval = "backend-heartbeat-interval"
	flagBackendHeartbeatTimeout  = "backend-heartbeat-timeout"

	// TLS flags
	flagTrustedCAFile         = "trusted-ca-file"
	flagInsecureSkipTLSVerify = "insecure-skip-tls-verify"
	flagCertFile              = "cert-file"
	flagKeyFile               = "key-file"

	deprecatedFlagAgentID          = "id"
	deprecatedFlagKeepaliveTimeout = "keepalive-timeout"
)

func newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the sensu-agent version information",
		Run: func(cmd *cobra.Command, args []string) {
			version.Println("sensu-agent")
		},
	}

	return cmd
}

func newStartCommand(ctx context.Context, args []string, logger *logrus.Entry) *cobra.Command {
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
			cfg.DetectCloudProvider = viper.GetBool(flagDetectCloudProvider)
			cfg.DisableAssets = viper.GetBool(flagDisableAssets)
			cfg.EventsAPIRateLimit = rate.Limit(viper.GetFloat64(flagEventsRateLimit))
			cfg.EventsAPIBurstLimit = viper.GetInt(flagEventsBurstLimit)
			cfg.KeepaliveHandlers = viper.GetStringSlice(flagKeepaliveHandlers)
			cfg.KeepaliveInterval = uint32(viper.GetInt(flagKeepaliveInterval))
			cfg.KeepaliveWarningTimeout = uint32(viper.GetInt(flagKeepaliveWarningTimeout))
			cfg.KeepaliveCriticalTimeout = uint32(viper.GetInt(flagKeepaliveCriticalTimeout))
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
			cfg.AllowList = viper.GetString(flagAllowList)
			cfg.BackendHandshakeTimeout = viper.GetInt(flagBackendHandshakeTimeout)
			cfg.BackendHeartbeatInterval = viper.GetInt(flagBackendHeartbeatInterval)
			cfg.BackendHeartbeatTimeout = viper.GetInt(flagBackendHeartbeatTimeout)

			// TLS configuration
			cfg.TLS = &corev2.TLSOptions{}
			cfg.TLS.TrustedCAFile = viper.GetString(flagTrustedCAFile)
			cfg.TLS.InsecureSkipVerify = viper.GetBool(flagInsecureSkipTLSVerify)
			cfg.TLS.CertFile = viper.GetString(flagCertFile)
			cfg.TLS.KeyFile = viper.GetString(flagKeyFile)

			if cfg.KeepaliveCriticalTimeout != 0 && cfg.KeepaliveCriticalTimeout < cfg.KeepaliveWarningTimeout {
				logger.Fatalf("if set, --%s must be greater than --%s",
					flagKeepaliveCriticalTimeout, flagKeepaliveWarningTimeout)
			}

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

			// Workaround for https://github.com/sensu/sensu-go/issues/2357. Detect if
			// the flags for labels and annotations were changed. If so, use their
			// values since flags take precedence over config
			if flag := cmd.Flags().Lookup(flagLabels); flag != nil && flag.Changed {
				cfg.Labels = labels
			}
			if flag := cmd.Flags().Lookup(flagAnnotations); flag != nil && flag.Changed {
				cfg.Annotations = annotations
			}

			sensuAgent, err := agent.NewAgentContext(ctx, cfg)
			if err != nil {
				return err
			}

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
	configFileDefaultLocation := filepath.Join(path.SystemConfigDir(), "agent.yml")
	configFileDefault := fmt.Sprintf("path to sensu-agent config file (default %q)", configFileDefaultLocation)
	_ = configFlagSet.StringP(flagConfigFile, "c", "", configFileDefault)
	configFlagSet.SetOutput(ioutil.Discard)
	_ = configFlagSet.Parse(args[1:])

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
	viper.SetDefault(flagDetectCloudProvider, false)
	viper.SetDefault(flagDisableAPI, false)
	viper.SetDefault(flagDisableSockets, false)
	viper.SetDefault(flagDisableAssets, false)
	viper.SetDefault(flagEventsRateLimit, agent.DefaultEventsAPIRateLimit)
	viper.SetDefault(flagEventsBurstLimit, agent.DefaultEventsAPIBurstLimit)
	viper.SetDefault(flagKeepaliveInterval, agent.DefaultKeepaliveInterval)
	viper.SetDefault(flagKeepaliveWarningTimeout, corev2.DefaultKeepaliveTimeout)
	viper.SetDefault(flagKeepaliveCriticalTimeout, 0)
	viper.SetDefault(flagNamespace, agent.DefaultNamespace)
	viper.SetDefault(flagPassword, agent.DefaultPassword)
	viper.SetDefault(flagRedact, corev2.DefaultRedactFields)
	viper.SetDefault(flagSocketHost, agent.DefaultSocketHost)
	viper.SetDefault(flagSocketPort, agent.DefaultSocketPort)
	viper.SetDefault(flagStatsdDisable, agent.DefaultStatsdDisable)
	viper.SetDefault(flagStatsdFlushInterval, agent.DefaultStatsdFlushInterval)
	viper.SetDefault(flagStatsdMetricsHost, agent.DefaultStatsdMetricsHost)
	viper.SetDefault(flagStatsdMetricsPort, agent.DefaultStatsdMetricsPort)
	viper.SetDefault(flagStatsdEventHandlers, []string{})
	viper.SetDefault(flagSubscriptions, []string{})
	viper.SetDefault(flagUser, agent.DefaultUser)
	viper.SetDefault(flagTrustedCAFile, "")
	viper.SetDefault(flagInsecureSkipTLSVerify, false)
	viper.SetDefault(flagLogLevel, "warn")
	viper.SetDefault(flagBackendHandshakeTimeout, 15)
	viper.SetDefault(flagBackendHeartbeatInterval, 30)
	viper.SetDefault(flagBackendHeartbeatTimeout, 45)

	// Merge in config flag set so that it appears in command usage
	cmd.Flags().AddFlagSet(configFlagSet)

	// Flags
	// Load the configuration file but only error out if flagConfigFile is used
	cmd.Flags().Bool(flagDeregister, viper.GetBool(flagDeregister), "ephemeral agent")
	cmd.Flags().Int(flagAPIPort, viper.GetInt(flagAPIPort), "port the Sensu client HTTP API listens on")
	cmd.Flags().Int(flagSocketPort, viper.GetInt(flagSocketPort), "port the Sensu client socket listens on")
	cmd.Flags().String(flagAgentName, viper.GetString(flagAgentName), "agent name (defaults to hostname)")
	cmd.Flags().String(flagAPIHost, viper.GetString(flagAPIHost), "address to bind the Sensu client HTTP API to")
	cmd.Flags().String(flagCacheDir, viper.GetString(flagCacheDir), "path to store cached data")
	cmd.Flags().String(flagDeregistrationHandler, viper.GetString(flagDeregistrationHandler), "deregistration handler that should process the entity deregistration event")
	cmd.Flags().Bool(flagDetectCloudProvider, viper.GetBool(flagDetectCloudProvider), "enable cloud provider detection")
	cmd.Flags().Float64(flagEventsRateLimit, viper.GetFloat64(flagEventsRateLimit), "maximum number of events transmitted to the backend through the /events api")
	cmd.Flags().Int(flagEventsBurstLimit, viper.GetInt(flagEventsBurstLimit), "/events api burst limit")
	cmd.Flags().String(flagNamespace, viper.GetString(flagNamespace), "agent namespace")
	cmd.Flags().String(flagPassword, viper.GetString(flagPassword), "agent password")
	cmd.Flags().StringSlice(flagRedact, viper.GetStringSlice(flagRedact), "comma-delimited list of fields to redact, overwrites the default fields. This flag can also be invoked multiple times")
	cmd.Flags().String(flagSocketHost, viper.GetString(flagSocketHost), "address to bind the Sensu client socket to")
	cmd.Flags().Bool(flagStatsdDisable, viper.GetBool(flagStatsdDisable), "disables the statsd listener and metrics server")
	cmd.Flags().StringSlice(flagStatsdEventHandlers, viper.GetStringSlice(flagStatsdEventHandlers), "comma-delimited list of event handlers for statsd metrics. This flag can also be invoked multiple times")
	cmd.Flags().Int(flagStatsdFlushInterval, viper.GetInt(flagStatsdFlushInterval), "number of seconds between statsd flush")
	cmd.Flags().String(flagStatsdMetricsHost, viper.GetString(flagStatsdMetricsHost), "address used for the statsd metrics server")
	cmd.Flags().Int(flagStatsdMetricsPort, viper.GetInt(flagStatsdMetricsPort), "port used for the statsd metrics server")
	cmd.Flags().StringSlice(flagSubscriptions, viper.GetStringSlice(flagSubscriptions), "comma-delimited list of agent subscriptions. This flag can also be invoked multiple times")
	cmd.Flags().String(flagUser, viper.GetString(flagUser), "agent user")
	cmd.Flags().StringSlice(flagBackendURL, viper.GetStringSlice(flagBackendURL), "comma-delimited list of ws/wss URLs of Sensu backend servers. This flag can also be invoked multiple times")
	cmd.Flags().StringSlice(flagKeepaliveHandlers, viper.GetStringSlice(flagKeepaliveHandlers), "comma-delimited list of keepalive handlers for this entity. This flag can also be invoked multiple times")
	cmd.Flags().Int(flagKeepaliveInterval, viper.GetInt(flagKeepaliveInterval), "number of seconds to send between keepalive events")
	cmd.Flags().Uint32(flagKeepaliveWarningTimeout, uint32(viper.GetInt(flagKeepaliveWarningTimeout)), "number of seconds until agent is considered dead by backend to create a warning event")
	cmd.Flags().Uint32(flagKeepaliveCriticalTimeout, uint32(viper.GetInt(flagKeepaliveCriticalTimeout)), "number of seconds until agent is considered dead by backend to create a critical event")
	cmd.Flags().Bool(flagDisableAPI, viper.GetBool(flagDisableAPI), "disable the Agent HTTP API")
	cmd.Flags().Bool(flagDisableAssets, viper.GetBool(flagDisableAssets), "disable check assets on this agent")
	cmd.Flags().Bool(flagDisableSockets, viper.GetBool(flagDisableSockets), "disable the Agent TCP and UDP event sockets")
	cmd.Flags().String(flagTrustedCAFile, viper.GetString(flagTrustedCAFile), "TLS CA certificate bundle in PEM format")
	cmd.Flags().Bool(flagInsecureSkipTLSVerify, viper.GetBool(flagInsecureSkipTLSVerify), "skip TLS verification (not recommended!)")
	cmd.Flags().String(flagCertFile, viper.GetString(flagCertFile), "certificate for TLS authentication")
	cmd.Flags().String(flagKeyFile, viper.GetString(flagKeyFile), "key for TLS authentication")
	cmd.Flags().String(flagLogLevel, viper.GetString(flagLogLevel), "logging level [panic, fatal, error, warn, info, debug]")
	cmd.Flags().StringToStringVar(&labels, flagLabels, nil, "entity labels map")
	cmd.Flags().StringToStringVar(&annotations, flagAnnotations, nil, "entity annotations map")
	cmd.Flags().String(flagAllowList, viper.GetString(flagAllowList), "path to agent execution allow list configuration file")
	cmd.Flags().Int(flagBackendHandshakeTimeout, viper.GetInt(flagBackendHandshakeTimeout), "number of seconds the agent should wait when negotiating a new WebSocket connection")
	cmd.Flags().Int(flagBackendHeartbeatInterval, viper.GetInt(flagBackendHeartbeatInterval), "interval at which the agent should send heartbeats to the backend")
	cmd.Flags().Int(flagBackendHeartbeatTimeout, viper.GetInt(flagBackendHeartbeatTimeout), "number of seconds the agent should wait for a response to a hearbeat")

	cmd.Flags().SetNormalizeFunc(aliasNormalizeFunc(logger))

	if err := viper.ReadInConfig(); err != nil && configFile != "" {
		setupErr = err
	}

	deprecatedConfigAttributes(logger)
	viper.RegisterAlias(deprecatedFlagAgentID, flagAgentName)
	viper.RegisterAlias(deprecatedFlagKeepaliveTimeout, flagKeepaliveWarningTimeout)

	return cmd
}

func aliasNormalizeFunc(logger *logrus.Entry) func(*pflag.FlagSet, string) pflag.NormalizedName {
	return func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		// Wait until the command-line flags have been parsed
		if !f.Parsed() {
			return pflag.NormalizedName(name)
		}

		switch name {
		case deprecatedFlagAgentID:
			deprecatedFlagMessage(name, flagAgentName, logger)
			name = flagAgentName
		case deprecatedFlagKeepaliveTimeout:
			deprecatedFlagMessage(name, flagKeepaliveWarningTimeout, logger)
			name = flagKeepaliveWarningTimeout
		}
		return pflag.NormalizedName(name)
	}
}

// Look up the deprecated attributes in our config file and print a warning
// message if set
func deprecatedConfigAttributes(logger *logrus.Entry) {
	attributes := map[string]string{
		deprecatedFlagAgentID:          flagAgentName,
		deprecatedFlagKeepaliveTimeout: flagKeepaliveWarningTimeout,
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

func deprecatedFlagMessage(oldFlag, newFlag string, logger *logrus.Entry) {
	logger.Warningf("flag --%s has been deprecated, please use --%s instead",
		oldFlag, newFlag)
}
