package clusterrolebinding

import (
	"io"
	"net/http"
	"strconv"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"

	"github.com/spf13/cobra"
)

// ListCommand defines a command to list cluster role bindings
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list cluster role bindings",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := helpers.ListOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			// Fetch role bindings from API
			var header http.Header
			results := []corev2.ClusterRoleBinding{}
			err = cli.Client.List(client.ClusterRoleBindingsPath(), &results, &opts, &header)
			if err != nil {
				return err
			}

			// Print the results based on the user preferences
			resources := []corev3.Resource{}
			for i := range results {
				resources = append(resources, &results[i])
			}
			return helpers.PrintList(cmd, cli.Config.Format(), printToTable, resources, results, header)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	helpers.AddFieldSelectorFlag(cmd.Flags())
	helpers.AddLabelSelectorFlag(cmd.Flags())
	helpers.AddChunkSizeFlag(cmd.Flags())

	return cmd
}
func printToTable(results interface{}, writer io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				clusterRoleBinding, ok := data.(corev2.ClusterRoleBinding)
				if !ok {
					return cli.TypeError
				}
				return clusterRoleBinding.Name
			},
		},
		{
			Title: "Cluster Role",
			CellTransformer: func(data interface{}) string {
				clusterRoleBinding, ok := data.(corev2.ClusterRoleBinding)
				if !ok {
					return cli.TypeError
				}
				return clusterRoleBinding.RoleRef.Name
			},
		},
		{
			Title: "Subjects",
			CellTransformer: func(data interface{}) string {
				clusterRoleBinding, ok := data.(corev2.ClusterRoleBinding)
				if !ok {
					return cli.TypeError
				}
				return strconv.Itoa(len(clusterRoleBinding.Subjects))
			},
		},
	})
	table.Render(writer, results)
}
