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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer cancel()
		logger.Info("signal received: ", <-sigs)
	}()

	command := cmd.NewRootCommand(ctx, os.Args)
	command.AddCommand(cmd.NewWindowsInstallCommand())
	command.AddCommand(cmd.NewWindowsUninstallCommand())
	command.AddCommand(cmd.NewWindowsStartServiceCommand())
	command.AddCommand(cmd.NewWindowsStopCommand())

	if err := command.Execute(); err != nil {
		logger.WithError(err).Fatal("error executing sensu-agent")
	}
}
