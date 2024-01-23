package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/agent"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/util/path"
	"github.com/sensu/sensu-go/util/url"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

var (
	annotations               map[string]string
	labels                    map[string]string
	keepaliveCheckLabels      map[string]string
	keepaliveCheckAnnotations map[string]string
	configFileDefaultLocation = filepath.Join(path.SystemConfigDir(), "agent.yml")
)

const (
	// DefaultBackendPort specifies the default port to use when a port is not
	// specified in backend urls
	DefaultBackendPort = "8081"

	environmentPrefix = "sensu"

	flagAgentName                 = "name"
	flagAPIHost                   = "api-host"
	flagAPIPort                   = "api-port"
	flagAssetsRateLimit           = "assets-rate-limit"
	flagAssetsBurstLimit          = "assets-burst-limit"
	flagBackendURL                = "backend-url"
	flagCacheDir                  = "cache-dir"
	flagConfigFile                = "config-file"
	flagDeregister                = "deregister"
	flagDeregistrationHandler     = "deregistration-handler"
	flagDetectCloudProvider       = "detect-cloud-provider"
	flagEventsRateLimit           = "events-rate-limit"
	flagEventsBurstLimit          = "events-burst-limit"
	flagKeepaliveHandlers         = "keepalive-handlers"
	flagKeepaliveInterval         = "keepalive-interval"
	flagKeepaliveWarningTimeout   = "keepalive-warning-timeout"
	flagKeepaliveCriticalTimeout  = "keepalive-critical-timeout"
	flagKeepaliveCheckLabels      = "keepalive-check-labels"
	flagKeepaliveCheckAnnotations = "keepalive-check-annotations"
	flagKeepalivePipelines        = "keepalive-pipelines"
	flagNamespace                 = "namespace"
	flagPassword                  = "password"
	flagRedact                    = "redact"
	flagSocketHost                = "socket-host"
	flagSocketPort                = "socket-port"
	flagStatsdDisable             = "statsd-disable"
	flagStatsdEventHandlers       = "statsd-event-handlers"
	flagStatsdFlushInterval       = "statsd-flush-interval"
	flagStatsdMetricsHost         = "statsd-metrics-host"
	flagStatsdMetricsPort         = "statsd-metrics-port"
	flagSubscriptions             = "subscriptions"
	flagUser                      = "user"
	flagDisableAPI                = "disable-api"
	flagDisableAssets             = "disable-assets"
	flagDisableSockets            = "disable-sockets"
	flagLogLevel                  = "log-level"
	flagLabels                    = "labels"
	flagAnnotations               = "annotations"
	flagAllowList                 = "allow-list"
	flagBackendHandshakeTimeout   = "backend-handshake-timeout"
	flagBackendHeartbeatInterval  = "backend-heartbeat-interval"
	flagBackendHeartbeatTimeout   = "backend-heartbeat-timeout"
	flagAgentManagedEntity        = "agent-managed-entity"
	flagRetryMin                  = "retry-min"
	flagRetryMax                  = "retry-max"
	flagRetryMultiplier           = "retry-multiplier"
	flagMaxSessionLength          = "max-session-length"
	flagStripNetworks             = "strip-networks"

	// TLS flags
	flagTrustedCAFile         = "trusted-ca-file"
	flagInsecureSkipTLSVerify = "insecure-skip-tls-verify"
	flagCertFile              = "cert-file"
	flagKeyFile               = "key-file"

	// Deprecated flags
	deprecatedFlagAgentID          = "id"
	deprecatedFlagKeepaliveTimeout = "keepalive-timeout"
)

// InitializeFunc represents the signature of an initialization function, used
// to initialize the agent
type InitializeFunc func(context.Context, *agent.Config) (*agent.Agent, error)

// NewAgentConfig initializes the agent config using Viper
func NewAgentConfig(cmd *cobra.Command) (*agent.Config, error) {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return nil, err
	}
	level, err := logrus.ParseLevel(viper.GetString(flagLogLevel))
	if err != nil {
		return nil, err
	}
	logrus.SetLevel(level)

	cfg := agent.NewConfig()
	cfg.AgentManagedEntity = viper.GetBool(flagAgentManagedEntity)
	cfg.API.Host = viper.GetString(flagAPIHost)
	cfg.API.Port = viper.GetInt(flagAPIPort)
	cfg.AssetsRateLimit = rate.Limit(viper.GetFloat64(flagAssetsRateLimit))
	cfg.AssetsBurstLimit = viper.GetInt(flagAssetsBurstLimit)
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
	cfg.KeepaliveCheckLabels = viper.GetStringMapString(flagKeepaliveCheckLabels)
	cfg.KeepaliveCheckAnnotations = viper.GetStringMapString(flagKeepaliveCheckAnnotations)
	cfg.KeepalivePipelines = viper.GetStringSlice(flagKeepalivePipelines)
	cfg.Namespace = viper.GetString(flagNamespace)
	cfg.Password = viper.GetString(flagPassword)
	cfg.Socket.Host = viper.GetString(flagSocketHost)
	cfg.Socket.Port = viper.GetInt(flagSocketPort)
	cfg.StatsdServer.Disable = viper.GetBool(flagStatsdDisable)
	cfg.StatsdServer.FlushInterval = viper.GetInt(flagStatsdFlushInterval)
	cfg.StatsdServer.Host = viper.GetString(flagStatsdMetricsHost)
	cfg.StatsdServer.Port = viper.GetInt(flagStatsdMetricsPort)
	cfg.StatsdServer.Handlers = viper.GetStringSlice(flagStatsdEventHandlers)
	cfg.User = viper.GetString(flagUser)
	cfg.AllowList = viper.GetString(flagAllowList)
	cfg.BackendHandshakeTimeout = viper.GetInt(flagBackendHandshakeTimeout)
	cfg.BackendHeartbeatInterval = viper.GetInt(flagBackendHeartbeatInterval)
	cfg.BackendHeartbeatTimeout = viper.GetInt(flagBackendHeartbeatTimeout)
	cfg.RetryMin = viper.GetDuration(flagRetryMin)
	cfg.RetryMax = viper.GetDuration(flagRetryMax)
	cfg.RetryMultiplier = viper.GetFloat64(flagRetryMultiplier)
	cfg.MaxSessionLength = viper.GetDuration(flagMaxSessionLength)
	cfg.StripNetworks = viper.GetBool(flagStripNetworks)

	// Set the labels & annotations using values defined configuration files
	// and/or environment variables for now
	cfg.Labels = viper.GetStringMapString(flagLabels)
	cfg.Annotations = viper.GetStringMapString(flagAnnotations)

	// TLS configuration
	cfg.TLS = &corev2.TLSOptions{}
	cfg.TLS.TrustedCAFile = viper.GetString(flagTrustedCAFile)
	cfg.TLS.InsecureSkipVerify = viper.GetBool(flagInsecureSkipTLSVerify)
	cfg.TLS.CertFile = viper.GetString(flagCertFile)
	cfg.TLS.KeyFile = viper.GetString(flagKeyFile)

	if cfg.KeepaliveCriticalTimeout != 0 && cfg.KeepaliveCriticalTimeout < cfg.KeepaliveWarningTimeout {
		return nil, fmt.Errorf("if set, --%s must be greater than --%s",
			flagKeepaliveCriticalTimeout, flagKeepaliveWarningTimeout)
	}

	agentName := viper.GetString(flagAgentName)
	if agentName != "" {
		cfg.AgentName = agentName
	}

	for _, backendURL := range viper.GetStringSlice(flagBackendURL) {
		newURL, err := url.AppendPortIfMissing(backendURL, DefaultBackendPort)
		if err != nil {
			return nil, err
		}
		cfg.BackendURLs = append(cfg.BackendURLs, newURL)
	}

	cfg.Redact = viper.GetStringSlice(flagRedact)
	cfg.Subscriptions = viper.GetStringSlice(flagSubscriptions)

	// Workaround for https://github.com/sensu/sensu-go/issues/2357. Detect if
	// the flags for labels and annotations were changed. If so, use their
	// values since flags take precedence over config & environment
	if flag := cmd.Flags().Lookup(flagLabels); flag != nil && flag.Changed {
		cfg.Labels = labels
	}
	if flag := cmd.Flags().Lookup(flagAnnotations); flag != nil && flag.Changed {
		cfg.Annotations = annotations
	}

	cfg.DisableAPI = viper.GetBool(flagDisableAPI)
	cfg.DisableSockets = viper.GetBool(flagDisableSockets)

	// Add the ManagedByLabel label value if the agent is managed by its entity
	if viper.GetBool(flagAgentManagedEntity) {
		if len(cfg.Labels) == 0 {
			cfg.Labels = make(map[string]string)
		}
		cfg.Labels[corev2.ManagedByLabel] = "sensu-agent"
	}

	return cfg, nil
}

// NewAgentRunE intializes and executes sensu-agent, and returns any errors
// encountered
func NewAgentRunE(initialize InitializeFunc, cmd *cobra.Command) func(cmd *cobra.Command, args []string) error {
	return NewAgentRunEWithContext(initialize, cmd, context.Background())
}

// NewAgentRunEWithContext is like NewAgentRunE, but takes a context.
func NewAgentRunEWithContext(initialize InitializeFunc, cmd *cobra.Command, ctx context.Context) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		cfg, err := NewAgentConfig(cmd)
		if err != nil {
			return err
		}

		sensuAgent, err := initialize(ctx, cfg)
		if err != nil {
			return err
		}

		return sensuAgent.Run(ctx)
	}
}

// StartCommand creates a new cobra command to start sensu-agent.
func StartCommand(initialize InitializeFunc) *cobra.Command {
	cmd, err := StartCommandWithError(initialize)
	if err != nil {
		// lol
		panic(err)
	}
	return cmd
}

// StartCommandWithError is like StartCommand, but returns an error instead of
// delegating the error handling to the RunE method.
func StartCommandWithError(initialize InitializeFunc) (*cobra.Command, error) {
	return StartCommandWithErrorAndContext(initialize, context.Background())
}

// StartCommandWithErrorAndContext is like StartCommandWithError, but takes a
// context.
func StartCommandWithErrorAndContext(initialize InitializeFunc, ctx context.Context) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:           "start",
		Short:         "start the sensu agent",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.RunE = NewAgentRunEWithContext(initialize, cmd, ctx)
	return cmd, handleConfig(cmd, os.Args[1:])
}

func handleConfig(cmd *cobra.Command, arguments []string) error {
	configFlags := flagSet()
	configFlags.AddFlagSet(cmd.Flags())
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
	viper.SetDefault(flagAssetsRateLimit, asset.DefaultAssetsRateLimit)
	viper.SetDefault(flagAssetsBurstLimit, asset.DefaultAssetsBurstLimit)
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
	viper.SetDefault(flagLogLevel, "info")
	viper.SetDefault(flagBackendHandshakeTimeout, 15)
	viper.SetDefault(flagBackendHeartbeatInterval, 30)
	viper.SetDefault(flagBackendHeartbeatTimeout, 45)
	viper.SetDefault(flagRetryMin, time.Second)
	viper.SetDefault(flagRetryMax, 120*time.Second)
	viper.SetDefault(flagRetryMultiplier, 2.0)
	viper.SetDefault(flagMaxSessionLength, 0*time.Second)
	viper.SetDefault(flagStripNetworks, false)

	// Merge in flag set so that it appears in command usage
	flags := flagSet()
	cmd.Flags().AddFlagSet(flags)

	cmd.Flags().SetNormalizeFunc(aliasNormalizeFunc(logger))

	if err := viper.ReadInConfig(); err != nil && configFilePathIsDefined {
		return err
	}

	viper.SetEnvPrefix(environmentPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	deprecatedConfigAttributes(logger)

	return nil
}

func RegisterConfigAliases() {
	viper.RegisterAlias(deprecatedFlagAgentID, flagAgentName)
	viper.RegisterAlias(deprecatedFlagKeepaliveTimeout, flagKeepaliveWarningTimeout)
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

func flagSet() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("sensu-agent", pflag.ContinueOnError)

	// Config file flag
	configFileDescription := fmt.Sprintf("path to sensu-agent config file (default %q)", configFileDefaultLocation)
	_ = flagSet.StringP(flagConfigFile, "c", "", configFileDescription)

	// Common flags
	flagSet.Bool(flagDeregister, viper.GetBool(flagDeregister), "ephemeral agent")
	flagSet.Int(flagAPIPort, viper.GetInt(flagAPIPort), "port the Sensu client HTTP API listens on")
	flagSet.Int(flagSocketPort, viper.GetInt(flagSocketPort), "port the Sensu client socket listens on")
	flagSet.String(flagAgentName, viper.GetString(flagAgentName), "agent name (defaults to hostname)")
	flagSet.String(flagAPIHost, viper.GetString(flagAPIHost), "address to bind the Sensu client HTTP API to")
	flagSet.String(flagCacheDir, viper.GetString(flagCacheDir), "path to store cached data")
	flagSet.String(flagDeregistrationHandler, viper.GetString(flagDeregistrationHandler), "deregistration handler that should process the entity deregistration event")
	flagSet.Bool(flagDetectCloudProvider, viper.GetBool(flagDetectCloudProvider), "enable cloud provider detection")
	flagSet.Float64(flagAssetsRateLimit, viper.GetFloat64(flagAssetsRateLimit), "maximum number of assets fetched per second")
	flagSet.Int(flagAssetsBurstLimit, viper.GetInt(flagAssetsBurstLimit), "asset fetch burst limit")
	flagSet.Float64(flagEventsRateLimit, viper.GetFloat64(flagEventsRateLimit), "maximum number of events transmitted to the backend through the /events api")
	flagSet.Int(flagEventsBurstLimit, viper.GetInt(flagEventsBurstLimit), "/events api burst limit")
	flagSet.String(flagNamespace, viper.GetString(flagNamespace), "agent namespace")
	flagSet.String(flagPassword, viper.GetString(flagPassword), "agent password")
	flagSet.StringSlice(flagRedact, viper.GetStringSlice(flagRedact), "comma-delimited list of fields to redact, overwrites the default fields. This flag can also be invoked multiple times")
	flagSet.String(flagSocketHost, viper.GetString(flagSocketHost), "address to bind the Sensu client socket to")
	flagSet.Bool(flagStatsdDisable, viper.GetBool(flagStatsdDisable), "disables the statsd listener and metrics server")
	flagSet.StringSlice(flagStatsdEventHandlers, viper.GetStringSlice(flagStatsdEventHandlers), "comma-delimited list of event handlers for statsd metrics. This flag can also be invoked multiple times")
	flagSet.Int(flagStatsdFlushInterval, viper.GetInt(flagStatsdFlushInterval), "number of seconds between statsd flush")
	flagSet.String(flagStatsdMetricsHost, viper.GetString(flagStatsdMetricsHost), "address used for the statsd metrics server")
	flagSet.Int(flagStatsdMetricsPort, viper.GetInt(flagStatsdMetricsPort), "port used for the statsd metrics server")
	flagSet.StringSlice(flagSubscriptions, viper.GetStringSlice(flagSubscriptions), "comma-delimited list of agent subscriptions. This flag can also be invoked multiple times")
	flagSet.String(flagUser, viper.GetString(flagUser), "agent user")
	flagSet.StringSlice(flagBackendURL, viper.GetStringSlice(flagBackendURL), "comma-delimited list of ws/wss URLs of Sensu backend servers. This flag can also be invoked multiple times")
	flagSet.StringSlice(flagKeepaliveHandlers, viper.GetStringSlice(flagKeepaliveHandlers), "comma-delimited list of keepalive handlers for this entity. This flag can also be invoked multiple times")
	flagSet.Int(flagKeepaliveInterval, viper.GetInt(flagKeepaliveInterval), "number of seconds to send between keepalive events")
	flagSet.Uint32(flagKeepaliveWarningTimeout, uint32(viper.GetInt(flagKeepaliveWarningTimeout)), "number of seconds until agent is considered dead by backend to create a warning event")
	flagSet.Uint32(flagKeepaliveCriticalTimeout, uint32(viper.GetInt(flagKeepaliveCriticalTimeout)), "number of seconds until agent is considered dead by backend to create a critical event")
	flagSet.StringToStringVar(&keepaliveCheckLabels, flagKeepaliveCheckLabels, nil, "keepalive labels map to add to keepalive events")
	flagSet.StringToStringVar(&keepaliveCheckAnnotations, flagKeepaliveCheckAnnotations, nil, "keepalive annotations map to add to keepalive events")
	flagSet.StringSlice(flagKeepalivePipelines, viper.GetStringSlice(flagKeepalivePipelines), "comma-delimited list of pipeline references for keepalive event")
	flagSet.Bool(flagDisableAPI, viper.GetBool(flagDisableAPI), "disable the Agent HTTP API")
	flagSet.Bool(flagDisableAssets, viper.GetBool(flagDisableAssets), "disable check assets on this agent")
	flagSet.Bool(flagDisableSockets, viper.GetBool(flagDisableSockets), "disable the Agent TCP and UDP event sockets")
	flagSet.String(flagTrustedCAFile, viper.GetString(flagTrustedCAFile), "TLS CA certificate bundle in PEM format")
	flagSet.Bool(flagInsecureSkipTLSVerify, viper.GetBool(flagInsecureSkipTLSVerify), "skip TLS verification (not recommended!)")
	flagSet.String(flagCertFile, viper.GetString(flagCertFile), "certificate for TLS authentication")
	flagSet.String(flagKeyFile, viper.GetString(flagKeyFile), "key for TLS authentication")
	flagSet.String(flagLogLevel, viper.GetString(flagLogLevel), "logging level [panic, fatal, error, warn, info, debug]")
	flagSet.StringToStringVar(&labels, flagLabels, nil, "entity labels map")
	flagSet.StringToStringVar(&annotations, flagAnnotations, nil, "entity annotations map")
	flagSet.String(flagAllowList, viper.GetString(flagAllowList), "path to agent execution allow list configuration file")
	flagSet.Int(flagBackendHandshakeTimeout, viper.GetInt(flagBackendHandshakeTimeout), "number of seconds the agent should wait when negotiating a new WebSocket connection")
	flagSet.Int(flagBackendHeartbeatInterval, viper.GetInt(flagBackendHeartbeatInterval), "interval at which the agent should send heartbeats to the backend")
	flagSet.Int(flagBackendHeartbeatTimeout, viper.GetInt(flagBackendHeartbeatTimeout), "number of seconds the agent should wait for a response to a hearbeat")
	flagSet.Bool(flagAgentManagedEntity, viper.GetBool(flagAgentManagedEntity), "manage this entity via the agent")
	flagSet.Duration(flagRetryMin, viper.GetDuration(flagRetryMin), "minimum amount of time to wait before retrying an agent connection to the backend")
	flagSet.Duration(flagRetryMax, viper.GetDuration(flagRetryMax), "maximum amount of time to wait before retrying an agent connection to the backend")
	flagSet.Float64(flagRetryMultiplier, viper.GetFloat64(flagRetryMultiplier), "value multiplied with the current retry delay to produce a longer retry delay (bounded by --retry-max)")
	flagSet.Duration(flagMaxSessionLength, viper.GetDuration(flagMaxSessionLength), "maximum amount of time after which the agent will reconnect to one of the configured backends (no maximum by default)")
	flagSet.Bool(flagStripNetworks, viper.GetBool(flagStripNetworks), "do not include Network info in agent entity state")

	flagSet.SetOutput(ioutil.Discard)

	return flagSet
}
