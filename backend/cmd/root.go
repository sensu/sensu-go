package cmd

import (
	"github.com/sensu/sensu-go/backend/config"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

var (
	backend Backend
	rootCmd = &cobra.Command{
		Use:   "sensu-backend",
		Short: "sensu backend",
	}
)

// Backend ...
type Backend interface {
	NewBackend(config *config.Backend) error
	Run() error
	Migration() error
	Status() types.StatusMap
	Stop()
}

// Execute ...
func Execute(b Backend) {
	backend = b
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err.Error())
	}
}
