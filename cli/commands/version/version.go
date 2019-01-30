package version

import (
	"fmt"

	"github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/sensu/sensu-go/version"
	"github.com/spf13/cobra"
)

// Command defines the version command
func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show the sensuctl version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("sensuctl version %s, build %s, built %s\n",
				version.Semver(),
				version.BuildSHA,
				version.BuildDate,
			)
		},
		Annotations: map[string]string{
			hooks.ConfigurationRequirement: hooks.ConfigurationNotRequired,
		},
	}
}
