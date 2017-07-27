package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands"
	hooks "github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/spf13/cobra"
)

// OrganizationFlagDefault default value to use for organization
const OrganizationFlagDefault = "default"

var (
	// sensuConfigPath contains the path to sensuctl configuration files
	sensuConfigPath string

	// sensuCachePath contains the path to sensuctl cache files
	sensuCachePath string
)

func userConfigPath() string {
	switch runtime.GOOS {
	case "windows":
		appDataPath := os.Getenv("APPDATA")
		if appDataPath == "" {
			h, _ := homedir.Dir()
			appDataPath = filepath.Join(h, "AppData", "Roaming")
		}
		return filepath.Join(appDataPath, "sensu", "sensuctl")
	default:
		xdgConfigPath := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfigPath == "" {
			h, _ := homedir.Dir()
			xdgConfigPath = filepath.Join(h, ".config")
		}
		return filepath.Join(xdgConfigPath, "sensu", "sensuctl")
	}
}

func userCachePath() string {
	switch runtime.GOOS {
	case "windows":
		localAppDataPath := os.Getenv("LOCALAPP")
		if localAppDataPath == "" {
			h, _ := homedir.Dir()
			localAppDataPath = filepath.Join(h, "AppData", "Local")
		}
		return filepath.Join(localAppDataPath, "sensu", "sensuctl")
	case "darwin":
		h, _ := homedir.Dir()
		return filepath.Join(h, "Library", "Caches", "sensu", "sensuctl")
	default:
		xdgCachePath := os.Getenv("XDG_CACHE_HOME")
		if xdgCachePath == "" {
			h, _ := homedir.Dir()
			xdgCachePath = filepath.Join(h, ".cache")
		}
		return filepath.Join(xdgCachePath, "sensu", "sensuctl")
	}
}

func init() {
	sensuConfigPath = userConfigPath()
	sensuCachePath = userCachePath()
}

func main() {
	rootCmd := configureRootCmd()
	sensuCli := cli.New(rootCmd.PersistentFlags())

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		if err := hooks.ConfigurationPresent(cmd, sensuCli); err != nil {
			return err
		}

		return nil
	}

	commands.AddCommands(rootCmd, sensuCli)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func configureRootCmd() *cobra.Command {
	showVersion := false
	cmd := &cobra.Command{
		Use:          cli.SensuCmdName,
		Short:        cli.SensuCmdName + " controls Sensu instances",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			if showVersion {
				showCLIVersion()
			} else {
				cmd.Help()
			}
		},
	}

	// Templates
	cmd.SetUsageTemplate(usageTemplate)

	// Version flag
	cmd.Flags().BoolVarP(&showVersion, "version", "v", false, "print version information")

	// Global flags
	cmd.PersistentFlags().StringP("api-url", "", "", "host URL of Sensu installation")
	cmd.PersistentFlags().StringP("config-dir", "", sensuConfigPath, "path to directory containing configuration files")
	cmd.PersistentFlags().StringP("cache-dir", "", sensuCachePath, "path to directory containing cache & temporary files")
	cmd.PersistentFlags().StringP("organization", "", OrganizationFlagDefault, "organization in which we perform actions")

	return cmd
}

func showCLIVersion() {
	// TODO: 😰
	fmt.Printf("Sensu CLI version %s\n", "0.1.alpha")
}

var usageTemplate = `Usage:

{{- if not .HasSubCommands}}	{{.UseLine}}{{end}}
{{- if .HasSubCommands}}	{{ .CommandPath}} COMMAND{{end}}

{{ .Short | trim }}

{{- if .HasFlags}}

Options:
{{ wrappedFlagUsages . | trimRightSpace}}

{{- end}}

{{- if hasOperationalSubCommands . }}

Commands:

{{- range operationalSubCommands . }}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}

{{- if hasManagementSubCommands . }}

Management Commands:

{{- range managementSubCommands . }}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}

{{- if .HasSubCommands }}

Run '{{.CommandPath}} COMMAND --help' for more information on a command.
{{- end}}
`

var helpTemplate = `
{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`
