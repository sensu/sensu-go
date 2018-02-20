package importer

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
		Use:           "import",
		Short:         "import resources from STDIN",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}
			stat, _ := cli.InFile.Stat()
			if stat.Mode()&os.ModeNamedPipe == 0 {
				_ = cmd.Help() // Print out usage
				return nil
			}

			var data map[string]interface{}
			dec := json.NewDecoder(bufio.NewReader(cli.InFile))

			if err := dec.Decode(&data); err != nil {
				printErrorMsg(
					cmd.OutOrStderr(),
					err.Error(),
				)
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
				printErrorMsg(
					cmd.OutOrStderr(),
					"Only importing of legacy settings are supported at this time.",
				)
				return errors.New("")
			}

			importer.AllowWarns, _ = cmd.Flags().GetBool(flagsForce)
			importer.Debug, _ = cmd.Flags().GetBool(flagsVerbose)
			importer.report.Out = cmd.OutOrStdout()

			err := importer.Run(data)
			fmt.Fprintln(cmd.OutOrStdout(), "\n==============================")

			if err != nil {
				printErrorMsg(cmd.OutOrStderr(), err.Error())
				return err
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

func printErrorMsg(wr io.Writer, msg string) {
	fmt.Fprintf(
		wr,
		"%s %s\n",
		globals.ErrorTextStyle("ERROR"),
		msg,
	)
}
