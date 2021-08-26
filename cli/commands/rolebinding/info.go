package rolebinding

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// InfoCommand defines new command to display detailed information about a role
// binding
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info [NAME]",
		Short:        "show detailed information about a role binding",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("a role binding name is required")
			}

			// Fetch roles from API
			r, err := cli.Client.FetchRoleBinding(args[0])
			if err != nil {
				return err
			}

			// Determine the format to use to output the data
			flag := helpers.GetChangedStringValueViper("format", cmd.Flags())
			format := cli.Config.Format()
			return helpers.PrintFormatted(flag, format, r, cmd.OutOrStdout(), printToList)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printToList(v interface{}, writer io.Writer) error {
	roleBinding, ok := v.(*types.RoleBinding)
	if !ok {
		return fmt.Errorf("%t is not a role binding", v)
	}

	userNames := []string{}
	groupNames := []string{}

	for _, subject := range roleBinding.Subjects {
		switch subject.Type {
		case types.GroupType:
			groupNames = append(groupNames, subject.Name)
		case types.UserType:
			userNames = append(userNames, subject.Name)
		}
	}

	cfg := &list.Config{
		Title: fmt.Sprintf("%s/%s", roleBinding.Namespace, roleBinding.Name),
		Rows: []*list.Row{
			{
				Label: "Name",
				Value: roleBinding.Name,
			},
			{
				Label: "Namespace",
				Value: roleBinding.Namespace,
			},
			{
				Label: roleBinding.RoleRef.Type,
				Value: roleBinding.RoleRef.Name,
			},
			{
				Label: "Subjects",
			},
		},
	}

	if len(userNames) > 0 {
		cfg.Rows = append(cfg.Rows, &list.Row{
			Label: "  Users",
			Value: strings.Join(userNames, ", "),
		})
	}

	if len(groupNames) > 0 {
		cfg.Rows = append(cfg.Rows, &list.Row{
			Label: "  Groups",
			Value: strings.Join(groupNames, ", "),
		})
	}

	return list.Print(writer, cfg)
}
