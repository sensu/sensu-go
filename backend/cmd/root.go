package cmd

import (
	"github.com/sensu/sensu-go/backend"
	"github.com/spf13/cobra"
)

var (
	sensuBackend *backend.Backend
	rootCmd      = &cobra.Command{
		Use:   "sensu-backend",
		Short: "sensu backend",
	}
)

// Execute ...
func Execute(b *backend.Backend) {
	sensuBackend = b
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err.Error())
	}
}
