package main

import (
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands"
	hooks "github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/spf13/cobra"
)

// OrganizationFlagDefault default value to use for organization
const OrganizationFlagDefault = "default"

var (
	// sensuPath contains the path to CLI configuration files
	sensuPath string
)

func init() {
	h, _ := homedir.Dir()
	sensuPath = filepath.Join(h, ".config", "sensu")
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
	cmd.PersistentFlags().StringP("config-dir", "d", sensuPath, "directory of configuration files to load")
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

Managment Commands:

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
