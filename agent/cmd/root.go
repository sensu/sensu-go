package cmd

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewRootCommand creates a new root sensu-agent command. The context provided
// is used for cancellation. The args are parsed with cobra. The first argument
// is assumed to be the binary name and is ignored.
func NewRootCommand(ctx context.Context, args []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sensu-agent",
		Short: "sensu agent",
	}

	cmd.AddCommand(newVersionCommand())
	cmd.AddCommand(newStartCommand(args, ctx))

	viper.SetEnvPrefix("sensu")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	return cmd
}
