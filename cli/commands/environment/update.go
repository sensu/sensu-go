package environment

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// UpdateCommand allows the user to update environment
func UpdateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "update [NAME]",
		Short:        "update environment description",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Print out usage if we do not receive one argument
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch environment from API
			name := args[0]
			env, err := cli.Client.FetchEnvironment(name)
			if err != nil {
				return err
			}

			opts := envOpts{}
			opts.Org = cli.Config.Organization()
			opts.withEnv(env)

			if err := opts.administerQuestionnaire(true); err != nil {
				return err
			}

			opts.Copy(env)

			if err := env.Validate(); err != nil {
				return err
			}

			if err := cli.Client.UpdateEnvironment(env); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	return cmd
}
