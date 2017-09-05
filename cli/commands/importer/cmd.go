package importer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/elements/globals"
	"github.com/spf13/cobra"
)

const (
	flagsLegacy  = "legacy"
	flagsForce   = "force"
	flagsVerbose = "verbose"
)

// ImportCommand adds command that allows user import resources
func ImportCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := cobra.Command{
		Use:          "import",
		Short:        "import resources from STDIN",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			stat, _ := cli.InFile.Stat()
			if stat.Mode()&os.ModeNamedPipe == 0 {
				cmd.Help() // Print out usage
				return nil
			}

			var data map[string]interface{}
			dec := json.NewDecoder(bufio.NewReader(cli.InFile))

			if err := dec.Decode(&data); err != nil {
				return err
			}

			var importer *Importer
			if legacy, _ := cmd.Flags().GetBool(flagsLegacy); legacy {
				importer = NewSensuV1SettingsImporter(
					cli.Config.Organization(),
					cli.Config.Environment(),
					cli.Client,
				)
			} else {
				fmt.Fprintln(
					cmd.OutOrStderr(),
					"Only importing of legacy settings are supported at this time.",
				)
				return nil
			}

			importer.AllowWarns, _ = cmd.Flags().GetBool(flagsForce)
			importer.Debug, _ = cmd.Flags().GetBool(flagsVerbose)
			importer.reporter.Out = cmd.OutOrStdout()

			err := importer.Run(data)
			fmt.Fprintln(cmd.OutOrStdout(), "\n==============================")

			if err != nil {
				fmt.Fprintf(
					cmd.OutOrStderr(),
					"%s %s\n",
					globals.ErrorTextStyle("ERROR"),
					err.Error(),
				)
				return nil
			}

			fmt.Fprintln(
				cmd.OutOrStdout(),
				globals.SuccessStyle("SUCCESS"),
				"all resources imported",
			)
			return nil
		},
	}

	cmd.Flags().Bool(flagsLegacy, false, "import Sensu V1 settings")
	cmd.Flags().Bool(flagsForce, false, "attempt to import resources regardless of any wanrings")
	cmd.Flags().BoolP(flagsVerbose, "v", false, "include debug messages in output")

	return &cmd
}
