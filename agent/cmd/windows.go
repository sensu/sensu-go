// +build windows

package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/svc"
)

const (
	serviceName = "sensu-agent"
	serviceDesc = "monitoring agent for sensu-go (https://sensu.io)"
)

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
			return installService(serviceName, serviceDesc, "start", "-c", configFile)
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

// NewWindowsStartServiceCommand creates a cobra command that starts an
// installed sensu-agent service.
func NewWindowsStartServiceCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "start-service",
		Short:         "start the sensu-agent service",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO figure this out. Should start even exist?
			// return controlService(serviceName, svc.Start, )
			return nil
		},
	}
}

// NewWindowsStopServiceCommand creates a cobra command that stops an
// installed sensu-agent service.
func NewWindowsStopServiceCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "start-service",
		Short:         "start the sensu-agent service",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return controlService(serviceName, svc.Stop, svc.Stopped)
		},
	}
}
