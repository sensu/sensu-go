package asset

import (
	"fmt"
	"io"
	"net/url"
	"path"

	"github.com/sensu/sensu-go/cli"
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			org := cli.Config.Organization()
			if ok, _ := cmd.Flags().GetBool("all-organizations"); ok {
				org = "*"
			}

			// Fetch assets from API
			r, err := cli.Client.ListAssets(org)
			if err != nil {
				return err
			}

			// Determine the format to use to output the data
			var format string
			if format, _ = cmd.Flags().GetString("format"); format == "" {
				format = cli.Config.Format()
			}

			if format == "json" {
				helpers.PrintJSON(r, cmd.OutOrStdout())
			} else {
				printAssetsToTable(r, cmd.OutOrStdout())
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags(), cli.Config)
	cmd.Flags().Bool("all-organizations", false, "Include records from all organizations")

	return cmd
}

func printAssetsToTable(queryResults []types.Asset, writer io.Writer) {
	rows := make([]*table.Row, len(queryResults))
	for i, result := range queryResults {
		rows[i] = &table.Row{Value: result}
	}

	table := table.New([]*table.Column{
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				asset, _ := data.(types.Asset)
				return asset.Name
			},
		},
		{
			Title: "URL",
			CellTransformer: func(data interface{}) string {
				asset, _ := data.(types.Asset)
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
				asset, _ := data.(types.Asset)

				if len(asset.Sha512) >= 7 {
					return string(asset.Sha512[0:7])
				}
				return ""
			},
		},
		{
			Title: "Metadata",
			CellTransformer: func(data interface{}) string {
				asset, _ := data.(types.Asset)
				output := ""

				for k, v := range asset.Metadata {
					output += k + ":" + v + " "
				}

				return output
			},
		},
	})

	table.Render(writer, rows)
}
