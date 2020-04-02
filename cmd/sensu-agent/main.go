//+build !windows

package main

import (
	"github.com/sensu/sensu-go/agent"
	"github.com/sensu/sensu-go/agent/cmd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var logger = logrus.WithFields(logrus.Fields{
	"component": "agent",
})

func main() {
	rootCmd := &cobra.Command{
		Use:   "sensu-agent",
		Short: "sensu agent",
	}

	rootCmd.AddCommand(cmd.VersionCommand())
	rootCmd.AddCommand(cmd.StartCommand(agent.Initialize))

	if err := rootCmd.Execute(); err != nil {
		logger.WithError(err).Fatal("error executing sensu-agent")
	}
}
