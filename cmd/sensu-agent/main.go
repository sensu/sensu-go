package main

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "sensu-agent",
		Short: "sensu agent",
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		logger.WithError(err).Fatal("error executing sensu-agent")
	}
}
