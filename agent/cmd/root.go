package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "sensu-agent",
		Short: "sensu agent",
	}
)

// Execute ...
func Execute() error {
	return rootCmd.Execute()
}
