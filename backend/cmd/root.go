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
func Execute(b *backend.Backend) error {
	sensuBackend = b
	return rootCmd.Execute()
}
