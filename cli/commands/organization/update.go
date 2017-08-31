package organization

import (
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// UpdateCommand allows the user to update organization
func UpdateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "update NAME",
		Short:        "update organization description",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Print ot usage if we do not receive one argument
			if len(args) != 1 {
				cmd.Help()
				return nil
			}

			// Fetch organizations from API
			orgName := args[0]
			org, err := cli.Client.FetchOrganization(orgName)
			if err != nil {
				return err
			}

			opts := newOrgOpts()
			opts.withOrg(org)

			opts.administerQuestionnaire(true)

			opts.Copy(org)

			if err := org.Validate(); err != nil {
				return err
			}

			if err := cli.Client.CreateOrganization(org); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	return cmd
}
