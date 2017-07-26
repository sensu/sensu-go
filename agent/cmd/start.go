package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/agent"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	logger *logrus.Entry
)

const (
	flagBackendURL            = "backend-url"
	flagAgentID               = "id"
	flagOrganization          = "organization"
	flagUser                  = "user"
	flagSubscriptions         = "subscriptions"
	flagDeregister            = "deregister"
	flagDeregistrationHandler = "deregistration-handler"
	flagCacheDir              = "cache-dir"
	flagKeepaliveTimeout      = "keepalive-timeout"
	flagKeepaliveInterval     = "keepalive-interval"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logger = logrus.WithFields(logrus.Fields{
		"component": "cmd",
	})

	rootCmd.AddCommand(newStartCommand())

	viper.SetEnvPrefix("sensu")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

func newStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "start the sensu agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := agent.NewConfig()
			cfg.BackendURLs = viper.GetStringSlice(flagBackendURL)
			cfg.Deregister = viper.GetBool(flagDeregister)
			cfg.DeregistrationHandler = viper.GetString(flagDeregistrationHandler)
			cfg.CacheDir = viper.GetString(flagCacheDir)
			cfg.Organization = viper.GetString(flagOrganization)
			cfg.User = viper.GetString(flagUser)

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

	var defaultCacheDir string

	switch runtime.GOOS {
	case "windows":
		programDataDir := os.Getenv("PROGRAMDATA")
		defaultCacheDir = filepath.Join(programDataDir, "sensu", "cache")
	default:
		defaultCacheDir = "/var/cache/sensu"
	}

	cmd.Flags().String(flagOrganization, "default", "agent organization")
	viper.BindPFlag(flagOrganization, cmd.Flags().Lookup(flagOrganization))

	cmd.Flags().String(flagUser, "agent", "agent user")
	viper.BindPFlag(flagUser, cmd.Flags().Lookup(flagUser))

	cmd.Flags().String(flagCacheDir, defaultCacheDir, "path to store cached data")
	viper.BindPFlag(flagCacheDir, cmd.Flags().Lookup(flagCacheDir))

	cmd.Flags().Bool(flagDeregister, false, "ephemeral agent")
	viper.BindPFlag(flagDeregister, cmd.Flags().Lookup(flagDeregister))

	cmd.Flags().String(flagDeregistrationHandler, "", "deregistration handler that should process the entity deregistration event.")
	viper.BindPFlag(flagDeregistrationHandler, cmd.Flags().Lookup(flagDeregistrationHandler))

	cmd.Flags().StringSlice(flagBackendURL, []string{"ws://localhost:8081"}, "ws/wss URL of Sensu backend server (to specify multiple backends use this flag multiple times)")
	viper.BindPFlag(flagBackendURL, cmd.Flags().Lookup(flagBackendURL))

	cmd.Flags().String(flagAgentID, "", "agent ID (defaults to hostname)")
	viper.BindPFlag(flagAgentID, cmd.Flags().Lookup(flagAgentID))

	cmd.Flags().String(flagSubscriptions, "", "comma-delimited list of agent subscriptions")
	viper.BindPFlag(flagSubscriptions, cmd.Flags().Lookup(flagSubscriptions))

	cmd.Flags().Uint(flagKeepaliveTimeout, 120, "number of seconds until agent is considered dead by backend")
	viper.BindPFlag(flagKeepaliveTimeout, cmd.Flags().Lookup(flagKeepaliveTimeout))

	cmd.Flags().Int(flagKeepaliveInterval, 20, "number of seconds to send between keepalive events")
	viper.BindPFlag(flagKeepaliveInterval, cmd.Flags().Lookup(flagKeepaliveInterval))

	return cmd
}
