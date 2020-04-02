package main

// main_windows.go exists to provide a build artifact with a .exe extension,
// and to add commands to the root command that handle windows service
// management.

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sensu/sensu-go/agent/cmd"
	"github.com/sirupsen/logrus"
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
	command.AddCommand(cmd.NewWindowsServiceCommand())

	if err := rootCmd.Execute(); err != nil {
		logger.WithError(err).Fatal("error executing sensu-agent")
	}
}
