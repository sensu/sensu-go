package main

import (
	_ "net/http/pprof"

	"github.com/sensu/sensu-go/backend"
	"github.com/sensu/sensu-go/backend/cmd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var logger = logrus.WithFields(logrus.Fields{
	"component": "backend",
})

func main() {
	// Define our root command and add our commands
	rootCmd := &cobra.Command{
		Use:   "sensu-backend",
		Short: "sensu backend",
	}
	rootCmd.AddCommand(cmd.StartCommand(backend.Initialize))
	rootCmd.AddCommand(cmd.VersionCommand())

	if err := rootCmd.Execute(); err != nil {
		logger.WithError(err).Fatal("error executing sensu-backend")
	}
}
