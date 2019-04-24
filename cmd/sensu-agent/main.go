//+build !windows

package main

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

	if err := command.Execute(); err != nil {
		logger.WithError(err).Fatal("error executing sensu-agent")
	}
}
