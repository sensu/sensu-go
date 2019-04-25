package asset

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"

	"github.com/spf13/cobra"
)

// ListCommand defines new command responsible for listing assets
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list assets",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}
			namespace := cli.Config.Namespace()
			if ok, _ := cmd.Flags().GetBool(flags.AllNamespaces); ok {
				namespace = types.NamespaceTypeAll
			}

			fieldSelector, err := cmd.Flags().GetString(flags.FieldSelector)
			if err != nil {
				return err
			}

			labelSelector, err := cmd.Flags().GetString(flags.LabelSelector)
			if err != nil {
				return err
			}

			opts := client.ListOptions{}
			opts.FieldSelector = fieldSelector
			opts.LabelSelector = labelSelector

			// Fetch assets from API
			results, err := cli.Client.ListAssets(namespace, opts)
			if err != nil {
				return err
			}

			// Print the results based on the user preferences
			resources := []types.Resource{}
			for i := range results {
				resources = append(resources, &results[i])
			}
			return helpers.Print(cmd, cli.Config.Format(), printToTable, resources, results)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	helpers.AddAllNamespace(cmd.Flags())
	helpers.AddFieldSelectorFlag(cmd.Flags())
	helpers.AddLabelSelectorFlag(cmd.Flags())

	return cmd
}

func printToTable(results interface{}, writer io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				asset, ok := data.(types.Asset)
				if !ok {
					return cli.TypeError
				}
				return asset.Name
			},
		},
		{
			Title: "URL",
			CellTransformer: func(data interface{}) string {
				asset, ok := data.(types.Asset)
				if !ok {
					return cli.TypeError
				}
				u, err := url.Parse(asset.URL)
				if err != nil {
					return ""
				}

				_, file := path.Split(u.EscapedPath())
				return fmt.Sprintf(
					"//%s/.../%s",
					u.Hostname(),
					file,
				)
			},
		},
		{
			Title: "Hash",
			CellTransformer: func(data interface{}) string {
				asset, ok := data.(types.Asset)
				if !ok {
					return cli.TypeError
				}
				if len(asset.Sha512) >= 128 {
					return string(asset.Sha512[0:7])
				}
				return "invalid"
			},
		},
	})

	table.Render(writer, results)
}
