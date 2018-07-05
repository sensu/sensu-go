package cmd

import (
	"github.com/sensu/sensu-go/backend"
	"github.com/spf13/cobra"
)

var (
	// sensuBackend *backend.Backend
	initialize initializeFunc
	rootCmd    = &cobra.Command{
		Use:   "sensu-backend",
		Short: "sensu backend",
	}
)

type initializeFunc func(*backend.Config) (*backend.Backend, error)

// Execute ...
func Execute(fn initializeFunc) error {
	initialize = fn
	return rootCmd.Execute()
}
