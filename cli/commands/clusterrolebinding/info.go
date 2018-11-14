package clusterrolebinding

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

// InfoCommand defines new command to display detailed information about a
// cluster role binding
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info [NAME]",
		Short:        "show detailed information about a cluster role binding",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("a cluster role binding name is required")
			}

			// Fetch roles from API
			r, err := cli.Client.FetchClusterRoleBinding(args[0])
			if err != nil {
				return err
			}

			// Determine the format to use to output the data
			flag := helpers.GetChangedStringValueFlag("format", cmd.Flags())
			format := cli.Config.Format()
			return helpers.PrintFormatted(flag, format, r, cmd.OutOrStdout(), printToList)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printToList(v interface{}, writer io.Writer) error {
	clusterRoleBinding, ok := v.(*types.ClusterRoleBinding)
	if !ok {
		return fmt.Errorf("%t is not a ClusterRoleBinding", v)
	}

	groups := []string{}
	users := []string{}
	for _, subject := range clusterRoleBinding.Subjects {
		if strings.ToLower(subject.Type) == "group" {
			groups = append(groups, subject.Name)
		} else {
			users = append(users, subject.Name)
		}
	}

	cfg := &list.Config{
		Title: clusterRoleBinding.Name,
		Rows: []*list.Row{
			{
				Label: "Name",
				Value: clusterRoleBinding.Name,
			},
			{
				Label: "ClusterRole",
				Value: clusterRoleBinding.RoleRef.Name,
			},
		},
	}

	// Add groups and users subjects if provided
	if len(groups) > 0 {
		cfg.Rows = append(cfg.Rows, &list.Row{
			Label: "Groups",
			Value: strings.Join(groups, ", "),
		})
	}
	if len(users) > 0 {
		cfg.Rows = append(cfg.Rows, &list.Row{
			Label: "Users",
			Value: strings.Join(users, ", "),
		})
	}

	return list.Print(writer, cfg)
}
