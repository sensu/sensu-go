package rolebinding

import (
	"io"
	"net/http"
	"strconv"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"

	"github.com/spf13/cobra"
)

// ListCommand defines a command to list role bindings
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list role bindings",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace := cli.Config.Namespace()
			if ok, _ := cmd.Flags().GetBool(flags.AllNamespaces); ok {
				namespace = corev2.NamespaceTypeAll
			}

			opts, err := helpers.ListOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			// Fetch role bindings from API
			var header http.Header
			results := []corev2.RoleBinding{}
			err = cli.Client.List(client.RoleBindingsPath(namespace), &results, &opts, &header)
			if err != nil {
				return err
			}

			// Print the results based on the user preferences
			resources := []corev2.Resource{}
			for i := range results {
				resources = append(resources, &results[i])
			}
			return helpers.PrintList(cmd, cli.Config.Format(), printToTable, resources, results, header)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	helpers.AddAllNamespace(cmd.Flags())
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
				roleBinding, ok := data.(corev2.RoleBinding)
				if !ok {
					return cli.TypeError
				}
				return roleBinding.Name
			},
		},
		{
			Title: "Namespace",
			CellTransformer: func(data interface{}) string {
				roleBinding, ok := data.(corev2.RoleBinding)
				if !ok {
					return cli.TypeError
				}
				return roleBinding.Namespace
			},
		},
		{
			Title: "Cluster Role",
			CellTransformer: func(data interface{}) string {
				roleBinding, ok := data.(corev2.RoleBinding)
				if !ok {
					return cli.TypeError
				}
				if roleBinding.RoleRef.Type == "ClusterRole" {
					return roleBinding.RoleRef.Name
				}
				return ""
			},
		},
		{
			Title: "Role",
			CellTransformer: func(data interface{}) string {
				roleBinding, ok := data.(corev2.RoleBinding)
				if !ok {
					return cli.TypeError
				}
				if roleBinding.RoleRef.Type == "Role" {
					return roleBinding.RoleRef.Name
				}
				return ""
			},
		},
		{
			Title: "Subjects",
			CellTransformer: func(data interface{}) string {
				roleBinding, ok := data.(corev2.RoleBinding)
				if !ok {
					return cli.TypeError
				}
				return strconv.Itoa(len(roleBinding.Subjects))
			},
		},
	})
	table.Render(writer, results)
}
