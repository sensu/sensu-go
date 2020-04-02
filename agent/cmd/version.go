package cmd

import (
	"github.com/sensu/sensu-go/version"
	"github.com/spf13/cobra"
)

// VersionCommand ...
func VersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the sensu-agent version information",
		Run: func(cmd *cobra.Command, args []string) {
			version.Println("sensu-agent")
		},
	}

	return cmd
}
