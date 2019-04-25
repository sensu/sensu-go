// +build windows

package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	serviceName        = "SensuAgent"
	serviceDisplayName = "Sensu Agent"
)

// NewWindowsServiceCommand creates a cobra command that offers subcommands
// for installing, uninstalling and running sensu-agent as a windows service.
func NewWindowsServiceCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "service",
		Short: "operate sensu-agent as a windows service",
	}

	command.AddCommand(NewWindowsInstallServiceCommand())
	command.AddCommand(NewWindowsUninstallServiceCommand())
	command.AddCommand(NewWindowsRunServiceCommand())

	return command
}

// NewWindowsInstallServiceCommand creates a cobra command that installs a
// sensu-agent service in Windows.
func NewWindowsInstallServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "install",
		Short:         "install the sensu-agent service",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := cmd.Flag(flagConfigFile).Value.String()
			p, err := filepath.Abs(configFile)
			if err != nil {
				return fmt.Errorf("error reading config file: %s", err)
			}
			fi, err := os.Stat(p)
			if err != nil {
				return fmt.Errorf("error reading config file: %s", err)
			}
			if !fi.Mode().IsRegular() {
				return errors.New("error reading config file: not a regular file")
			}
			return installService(serviceName, serviceDisplayName, "service", "run", configFile)
		},
	}
	cmd.Flags().StringP(flagConfigFile, "c", "", "path to sensu-agent config file")
	return cmd
}

// NewWindowsUninstallServiceCommand creates a cobra command that uninstalls a
// sensu-agent service in Windows.
func NewWindowsUninstallServiceCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "uninstall",
		Short:         "uninstall the sensu-agent service",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return removeService(serviceName)
		},
	}
}

func NewWindowsRunServiceCommand() *cobra.Command {
	command := &cobra.Command{
		Use:           "run",
		Short:         "run the sensu-agent service (blocking)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runService(serviceName, false)
		},
	}
	return command
}
