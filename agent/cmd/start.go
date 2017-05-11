package main

import (
	"os"
	"os/signal"
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
	flagBackendURL    = "backend-url"
	flagAgentID       = "id"
	flagSubscriptions = "subscriptions"
	flagDeregister    = "deregister"
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
			cfg.BackendURL = viper.GetString(flagBackendURL)
			cfg.Deregister = viper.GetBool(flagDeregister)

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

	cmd.Flags().String(flagBackendURL, "ws://localhost:8081", "ws/wss URL of Sensu backend server(s)")
	viper.BindPFlag(flagBackendURL, cmd.Flags().Lookup(flagBackendURL))

	cmd.Flags().String(flagAgentID, "", "agent ID (defaults to hostname)")
	viper.BindPFlag(flagAgentID, cmd.Flags().Lookup(flagAgentID))

	cmd.Flags().String(flagSubscriptions, "", "comma-delimited list of agent subscriptions")
	viper.BindPFlag(flagSubscriptions, cmd.Flags().Lookup(flagSubscriptions))

	cmd.Flags().Bool(flagDeregister, false, "ephemeral agent")
	viper.BindPFlag(flagDeregister, cmd.Flags().Lookup(flagDeregister))

	return cmd
}
