package main

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "sensu-backend",
		Short: "sensu backend",
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err.Error())
	}
}
