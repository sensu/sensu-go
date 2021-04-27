package main

import (
	"github.com/sensu/sensu-go/agent/cmd"
	"github.com/spf13/cobra"
)

func addRootPlatformArguments(rootCmd *cobra.Command) {
	rootCmd.AddCommand(cmd.NewWindowsServiceCommand())
}

func addStartPlatformArguments(*cobra.Command) {}
