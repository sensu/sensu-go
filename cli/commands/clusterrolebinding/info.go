package clusterrolebinding

import (
	"errors"
	"fmt"
	"io"
	"strconv"

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
	clusterRoleBinding, ok := v.(*types.RoleBinding)
	if !ok {
		return fmt.Errorf("%t is not a ClusterRoleBinding", v)
	}

	cfg := &list.Config{
		Title: clusterRoleBinding.Name,
		Rows: []*list.Row{
			{
				Label: "Name",
				Value: clusterRoleBinding.Name,
			},
			{
				Label: "Role",
				Value: fmt.Sprintf("%s (%s)", clusterRoleBinding.RoleRef.Name, clusterRoleBinding.RoleRef.Type),
			},
			// TODO (Simon) Create a row for each subject once the API is working
			{
				Label: "Subjects",
				Value: strconv.Itoa(len(clusterRoleBinding.Subjects)),
			},
		},
	}

	return list.Print(writer, cfg)
}
