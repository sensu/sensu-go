package cmd

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
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

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logger := logrus.WithFields(logrus.Fields{
		"component": "cmd",
	})

	cmd.AddCommand(newVersionCommand())
	cmd.AddCommand(newStartCommand(ctx, args, logger))

	viper.SetEnvPrefix("sensu")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	return cmd
}
