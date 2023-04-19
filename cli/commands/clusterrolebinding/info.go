package clusterrolebinding

import (
	"errors"
	"fmt"
	"io"
	"strings"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/list"
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
			flag := helpers.GetChangedStringValueViper("format", cmd.Flags())
			format := cli.Config.Format()
			return helpers.PrintFormatted(flag, format, r, cmd.OutOrStdout(), printToList)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printToList(v interface{}, writer io.Writer) error {
	clusterRoleBinding, ok := v.(*v2.ClusterRoleBinding)
	if !ok {
		return fmt.Errorf("%t is not a cluster role binding", v)
	}

	userNames := []string{}
	groupNames := []string{}

	for _, subject := range clusterRoleBinding.Subjects {
		switch subject.Type {
		case v2.GroupType:
			groupNames = append(groupNames, subject.Name)
		case v2.UserType:
			userNames = append(userNames, subject.Name)
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
				Label: "Cluster Role",
				Value: clusterRoleBinding.RoleRef.Name,
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
