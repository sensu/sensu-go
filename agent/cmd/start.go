package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/agent"
	"github.com/sensu/sensu-go/util/path"
	"github.com/sensu/sensu-go/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	logger *logrus.Entry
)

const (
	flagConfigFile            = "config-file"
	flagBackendURL            = "backend-url"
	flagAgentID               = "id"
	flagEnvironment           = "environment"
	flagOrganization          = "organization"
	flagUser                  = "user"
	flagPassword              = "password"
	flagSubscriptions         = "subscriptions"
	flagDeregister            = "deregister"
	flagDeregistrationHandler = "deregistration-handler"
	flagCacheDir              = "cache-dir"
	flagKeepaliveTimeout      = "keepalive-timeout"
	flagKeepaliveInterval     = "keepalive-interval"
	flagHTTPSocketBind        = "http-socket-bind"
	flagHTTPSocketPort        = "http-socket-port"
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
		Use:   "start",
		Short: "start the sensu agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.BindPFlags(cmd.Flags())
			if setupErr != nil {
				return setupErr
			}

			cfg := agent.NewConfig()
			cfg.BackendURLs = viper.GetStringSlice(flagBackendURL)
			cfg.Deregister = viper.GetBool(flagDeregister)
			cfg.DeregistrationHandler = viper.GetString(flagDeregistrationHandler)
			cfg.CacheDir = viper.GetString(flagCacheDir)
			cfg.Environment = viper.GetString(flagEnvironment)
			cfg.Organization = viper.GetString(flagOrganization)
			cfg.User = viper.GetString(flagUser)
			cfg.Password = viper.GetString(flagPassword)
			cfg.HTTPSocket.Bind = viper.GetString(flagHTTPSocketBind)
			cfg.HTTPSocket.Port = viper.GetString(flagHTTPSocketPort)

			agentID := viper.GetString(flagAgentID)
			if agentID != "" {
				cfg.AgentID = agentID
			}

			subscriptions := viper.GetString(flagSubscriptions)
			if subscriptions != "" {
				// TODO(greg): we prooobably want someeee sort of input validation.
				cfg.Subscriptions = strings.Split(subscriptions, ",")
			}

			sensuAgent := agent.NewAgent(cfg)
			if err := sensuAgent.Run(); err != nil {
				return err
			}

			sigs := make(chan os.Signal, 1)
			done := make(chan struct{}, 1)

			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				sig := <-sigs
				logger.Info("signal received: ", sig)
				sensuAgent.Stop()
				done <- struct{}{}
			}()

			<-done
			return nil
		},
	}

	// Set up distinct flagset for handling config file
	configFlagSet := pflag.NewFlagSet("sensu", pflag.ContinueOnError)
	configFlagSet.StringP(flagConfigFile, "c", "", "path to sensu-agent config file")
	configFlagSet.SetOutput(ioutil.Discard)
	configFlagSet.Parse(os.Args[1:])

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

	// Only error out if flagConfigFile is used
	if err := viper.ReadInConfig(); err != nil && configFile != "" {
		setupErr = err
	}

	// Flag defaults
	viper.SetDefault(flagEnvironment, "default")
	viper.SetDefault(flagOrganization, "default")
	viper.SetDefault(flagUser, "agent")
	viper.SetDefault(flagPassword, "P@ssw0rd!")
	viper.SetDefault(flagCacheDir, path.SystemCacheDir("sensu-agent"))
	viper.SetDefault(flagDeregister, false)
	viper.SetDefault(flagDeregistrationHandler, "")
	viper.SetDefault(flagBackendURL, []string{"ws://127.0.0.1:8081"})
	viper.SetDefault(flagAgentID, "")
	viper.SetDefault(flagSubscriptions, "")
	viper.SetDefault(flagKeepaliveTimeout, 120)
	viper.SetDefault(flagKeepaliveInterval, 20)
	viper.SetDefault(flagHTTPSocketBind, "127.0.0.1")
	viper.SetDefault(flagHTTPSocketPort, "3031")

	// Merge in config flag set so that it appears in command usage
	cmd.Flags().AddFlagSet(configFlagSet)

	// Flags
	cmd.Flags().String(flagEnvironment, viper.GetString(flagEnvironment), "agent environment")
	cmd.Flags().String(flagOrganization, viper.GetString(flagOrganization), "agent organization")
	cmd.Flags().String(flagUser, viper.GetString(flagUser), "agent user")
	cmd.Flags().String(flagPassword, viper.GetString(flagPassword), "agent password")
	cmd.Flags().String(flagCacheDir, viper.GetString(flagCacheDir), "path to store cached data")
	cmd.Flags().Bool(flagDeregister, viper.GetBool(flagDeregister), "ephemeral agent")
	cmd.Flags().String(flagDeregistrationHandler, viper.GetString(flagDeregistrationHandler), "deregistration handler that should process the entity deregistration event.")
	cmd.Flags().StringSlice(flagBackendURL, viper.GetStringSlice(flagBackendURL), "ws/wss URL of Sensu backend server (to specify multiple backends use this flag multiple times)")
	cmd.Flags().String(flagAgentID, viper.GetString(flagAgentID), "agent ID (defaults to hostname)")
	cmd.Flags().String(flagSubscriptions, viper.GetString(flagSubscriptions), "comma-delimited list of agent subscriptions")
	cmd.Flags().Uint(flagKeepaliveTimeout, uint(viper.Get(flagKeepaliveTimeout).(int)), "number of seconds until agent is considered dead by backend")
	cmd.Flags().Int(flagKeepaliveInterval, viper.GetInt(flagKeepaliveInterval), "number of seconds to send between keepalive events")
	cmd.Flags().String(flagHTTPSocketBind, viper.GetString(flagHTTPSocketBind), "address the HTTP client socket listens on")
	cmd.Flags().String(flagHTTPSocketPort, viper.GetString(flagHTTPSocketPort), "port the HTTP client socket listens on")
	return cmd
}
