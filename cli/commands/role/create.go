package role

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// CreateCommand defines new command to create roles
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create [NAME]",
		Short:        "create new roles",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			rule := types.Rule{}

			role := &types.Role{Name: args[0]}
			if err := role.Validate(); err != nil {
				return err
			}

			opts := &ruleOpts{}

			opts.Namespace = cli.Config.Namespace()

			opts.withFlags(cmd.Flags())
			opts.Role = args[0]

			opts.Copy(&rule)
			if err := rule.Validate(); err != nil {
				return err
			}
			role.Rules = append(role.Rules, rule)

			if err := cli.Client.CreateRole(role); err != nil {
				return err
			}
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Created")
			return err
		},
	}

	_ = cmd.Flags().StringP("type", "t", "",
		"type associated with the rule, "+
			"allowed values: "+strings.Join(types.AllTypes, ", "),
	)
	_ = cmd.Flags().BoolP("create", "c", false, "create permission")
	_ = cmd.Flags().BoolP("read", "r", false, "read permission")
	_ = cmd.Flags().BoolP("update", "u", false, "update permission")
	_ = cmd.Flags().BoolP("delete", "d", false, "delete permission")

	return cmd
}
