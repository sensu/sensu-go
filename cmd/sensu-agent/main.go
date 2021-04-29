//+build !windows

package main

import (
	"context"

	"github.com/sensu/sensu-go/agent"
	"github.com/sensu/sensu-go/agent/cmd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var logger = logrus.WithFields(logrus.Fields{
	"component": "agent",
})

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Define our root command and add our commands
	rootCmd := &cobra.Command{
		Use:   "sensu-agent",
		Short: "sensu agent",
	}

	// Watch for the shutdown signals
	agent.GracefulShutdown(cancel)

	rootCmd.AddCommand(cmd.VersionCommand())
	startCmd, err := cmd.StartCommandWithErrorAndContext(agent.NewAgentContext, ctx)
	if err != nil {
		logger.WithError(err).Fatal("error handling agent config")
	}
	addRootPlatformArguments(rootCmd)
	addStartPlatformArguments(startCmd)
	rootCmd.AddCommand(startCmd)

	cmd.RegisterConfigAliases()

	if err := rootCmd.Execute(); err != nil {
		logger.WithError(err).Fatal("error executing sensu-agent")
	}
}
