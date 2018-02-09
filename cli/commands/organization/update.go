package organization

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// UpdateCommand allows the user to update organization
func UpdateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "update [NAME]",
		Short:        "update organization description",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Print out usage if we do not receive one argument
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch organizations from API
			orgName := args[0]
			org, err := cli.Client.FetchOrganization(orgName)
			if err != nil {
				return err
			}

			opts := newOrgOpts()
			opts.withOrg(org)

			if err := opts.administerQuestionnaire(true); err != nil {
				return err
			}

			opts.Copy(org)

			if err := org.Validate(); err != nil {
				return err
			}

			if err := cli.Client.UpdateOrganization(org); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	return cmd
}
