package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/agent"
	"github.com/sensu/sensu-go/types/dynamic"
	"github.com/sensu/sensu-go/util/path"
	"github.com/sensu/sensu-go/util/url"
	"github.com/sensu/sensu-go/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	logger *logrus.Entry
)

const (
	// DefaultBackendPort specifies the default port to use when a port is not
	// specified in backend urls
	DefaultBackendPort = "8081"

	flagAgentID               = "id"
	flagAPIHost               = "api-host"
	flagAPIPort               = "api-port"
	flagBackendURL            = "backend-url"
	flagCacheDir              = "cache-dir"
	flagConfigFile            = "config-file"
	flagDeregister            = "deregister"
	flagDeregistrationHandler = "deregistration-handler"
	flagEnvironment           = "environment"
	flagExtendedAttributes    = "custom-attributes"
	flagKeepaliveInterval     = "keepalive-interval"
	flagKeepaliveTimeout      = "keepalive-timeout"
	flagOrganization          = "organization"
	flagPassword              = "password"
	flagRedact                = "redact"
	flagSocketHost            = "socket-host"
	flagSocketPort            = "socket-port"
	flagSubscriptions         = "subscriptions"
	flagUser                  = "user"
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

func splitAndTrim(s string) []string {
	r := strings.Split(s, ",")
	for i := range r {
		r[i] = strings.TrimSpace(r[i])
	}
	return r
}

func newStartCommand() *cobra.Command {
	var setupErr error

	cmd := &cobra.Command{
		Use:   "start",
		Short: "start the sensu agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			if setupErr != nil {
				return setupErr
			}

			cfg := agent.NewConfig()
			cfg.API.Host = viper.GetString(flagAPIHost)
			cfg.API.Port = viper.GetInt(flagAPIPort)
			cfg.CacheDir = viper.GetString(flagCacheDir)
			cfg.Deregister = viper.GetBool(flagDeregister)
			cfg.DeregistrationHandler = viper.GetString(flagDeregistrationHandler)
			cfg.Environment = viper.GetString(flagEnvironment)
			cfg.ExtendedAttributes = []byte(viper.GetString(flagExtendedAttributes))
			cfg.KeepaliveInterval = viper.GetInt(flagKeepaliveInterval)
			cfg.KeepaliveTimeout = uint32(viper.GetInt(flagKeepaliveTimeout))
			cfg.Organization = viper.GetString(flagOrganization)
			cfg.Password = viper.GetString(flagPassword)
			cfg.Socket.Host = viper.GetString(flagSocketHost)
			cfg.Socket.Port = viper.GetInt(flagSocketPort)
			cfg.User = viper.GetString(flagUser)

			agentID := viper.GetString(flagAgentID)
			if agentID != "" {
				cfg.AgentID = agentID
			}

			for _, backendURL := range viper.GetStringSlice(flagBackendURL) {
				newURL, err := url.AppendPortIfMissing(backendURL, DefaultBackendPort)
				if err != nil {
					return err
				}
				cfg.BackendURLs = append(cfg.BackendURLs, newURL)
			}

			// Get a single or a list of redact fields
			redact := viper.GetString(flagRedact)
			if redact != "" {
				cfg.Redact = splitAndTrim(redact)
			} else {
				cfg.Redact = viper.GetStringSlice(flagRedact)
			}

			// Get a single or a list of subscriptions
			subscriptions := viper.GetString(flagSubscriptions)
			if subscriptions != "" {
				cfg.Subscriptions = splitAndTrim(subscriptions)
			} else {
				cfg.Subscriptions = viper.GetStringSlice(flagSubscriptions)
			}

			sensuAgent := agent.NewAgent(cfg)
			if err := sensuAgent.Run(); err != nil {
				return err
			}

			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				sig := <-sigs
				logger.Info("signal received: ", sig)
				sensuAgent.Stop()
			}()

			wg.Wait()
			return nil
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
	viper.SetDefault(flagAgentID, agent.GetDefaultAgentID())
	viper.SetDefault(flagAPIHost, agent.DefaultAPIHost)
	viper.SetDefault(flagAPIPort, agent.DefaultAPIPort)
	viper.SetDefault(flagBackendURL, []string{agent.DefaultBackendURL})
	viper.SetDefault(flagCacheDir, path.SystemCacheDir("sensu-agent"))
	viper.SetDefault(flagDeregister, false)
	viper.SetDefault(flagDeregistrationHandler, "")
	viper.SetDefault(flagEnvironment, agent.DefaultEnvironment)
	viper.SetDefault(flagKeepaliveInterval, agent.DefaultKeepaliveInterval)
	viper.SetDefault(flagKeepaliveTimeout, agent.DefaultKeepaliveTimeout)
	viper.SetDefault(flagOrganization, agent.DefaultOrganization)
	viper.SetDefault(flagPassword, agent.DefaultPassword)
	viper.SetDefault(flagRedact, dynamic.DefaultRedactFields)
	viper.SetDefault(flagSocketHost, agent.DefaultSocketHost)
	viper.SetDefault(flagSocketPort, agent.DefaultSocketPort)
	viper.SetDefault(flagSubscriptions, []string{})
	viper.SetDefault(flagUser, agent.DefaultUser)

	// Merge in config flag set so that it appears in command usage
	cmd.Flags().AddFlagSet(configFlagSet)

	// Flags
	// Load the configuration file but only error out if flagConfigFile is used
	cmd.Flags().Bool(flagDeregister, viper.GetBool(flagDeregister), "ephemeral agent")
	cmd.Flags().Int(flagAPIPort, viper.GetInt(flagAPIPort), "port the Sensu client HTTP API listens on")
	cmd.Flags().Int(flagKeepaliveInterval, viper.GetInt(flagKeepaliveInterval), "number of seconds to send between keepalive events")
	cmd.Flags().Int(flagSocketPort, viper.GetInt(flagSocketPort), "port the Sensu client socket listens on")
	cmd.Flags().String(flagAgentID, viper.GetString(flagAgentID), "agent ID (defaults to hostname)")
	cmd.Flags().String(flagAPIHost, viper.GetString(flagAPIHost), "address to bind the Sensu client HTTP API to")
	cmd.Flags().String(flagCacheDir, viper.GetString(flagCacheDir), "path to store cached data")
	cmd.Flags().String(flagDeregistrationHandler, viper.GetString(flagDeregistrationHandler), "deregistration handler that should process the entity deregistration event.")
	cmd.Flags().String(flagEnvironment, viper.GetString(flagEnvironment), "agent environment")
	cmd.Flags().String(flagExtendedAttributes, viper.GetString(flagExtendedAttributes), "custom attributes to include in the agent entity")
	cmd.Flags().String(flagOrganization, viper.GetString(flagOrganization), "agent organization")
	cmd.Flags().String(flagPassword, viper.GetString(flagPassword), "agent password")
	cmd.Flags().String(flagRedact, viper.GetString(flagRedact), "comma-delimited customized list of fields to redact")
	cmd.Flags().String(flagSocketHost, viper.GetString(flagSocketHost), "address to bind the Sensu client socket to")
	cmd.Flags().String(flagSubscriptions, viper.GetString(flagSubscriptions), "comma-delimited list of agent subscriptions")
	cmd.Flags().String(flagUser, viper.GetString(flagUser), "agent user")
	cmd.Flags().StringSlice(flagBackendURL, viper.GetStringSlice(flagBackendURL), "ws/wss URL of Sensu backend server (to specify multiple backends use this flag multiple times)")
	cmd.Flags().Uint32(flagKeepaliveTimeout, uint32(viper.GetInt(flagKeepaliveTimeout)), "number of seconds until agent is considered dead by backend")
	if err := viper.ReadInConfig(); err != nil && configFile != "" {
		setupErr = err
	}

	return cmd
}
