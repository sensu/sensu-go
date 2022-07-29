package apikey

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// RevokeCommand adds a command that deletes apikeys.
func RevokeCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "revoke [NAME]",
		Short:        "revoke api-key given name",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			name := args[0]
			if skipConfirm, _ := cmd.Flags().GetBool("skip-confirm"); !skipConfirm {
				opts := []survey.AskOpt{
					survey.WithStdio(cli.InFile, cli.OutFile, cli.ErrFile),
				}
				if confirmed := helpers.ConfirmDeleteResourceWithOpts(name, "apikey", opts...); !confirmed {
					fmt.Fprintln(cli.OutFile, "Canceled")
					return nil
				}
			}

			apikey := &corev2.APIKey{
				ObjectMeta: corev2.ObjectMeta{
					Name: name,
				},
			}
			err := cli.Client.Delete(apikey.URIPath())
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Deleted")
			return err
		},
	}

	cmd.Flags().Bool("skip-confirm", false, "skip interactive confirmation prompt")

	cmd.SetIn(cli.InFile)
	cmd.SetOutput(cli.OutFile)
	cmd.SetErr(cli.ErrFile)

	return cmd
}
