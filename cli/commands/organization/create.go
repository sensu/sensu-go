package organization

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// CreateCommand adds command that allows users to create new organizations
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create [NAME]",
		Short:        "create new organization",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)
			opts := newOrgOpts()

			if len(args) > 0 {
				opts.Name = args[0]
			}

			if isInteractive {
				if err := opts.administerQuestionnaire(false); err != nil {
					return err
				}
			} else {
				opts.withFlags(cmd.Flags())
			}

			org := types.Organization{}
			opts.Copy(&org)

			if err := org.Validate(); err != nil {
				if !isInteractive {
					cmd.SilenceUsage = false
				}
				return err
			}

			if err := cli.Client.CreateOrganization(&org); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Created")
			return nil
		},
	}

	cmd.Flags().StringP("description", "", "", "Description of organization")

	helpers.AddInteractiveFlag(cmd.Flags())
	return cmd
}
