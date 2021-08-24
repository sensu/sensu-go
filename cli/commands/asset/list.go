package asset

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"

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
				namespace = corev2.NamespaceTypeAll
			}

			opts, err := helpers.ListOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			// Fetch assets from API
			var header http.Header
			results := []corev2.Asset{}
			err = cli.Client.List(client.AssetsPath(namespace), &results, &opts, &header)
			if err != nil {
				return err
			}

			// Determine the user's preferred format
			var format string
			if format = helpers.GetChangedStringValueViper("format", cmd.Flags()); format == "" {
				format = cli.Config.Format()
			}

			// Print the results based on the user preferences
			resources := []corev2.Resource{}
			resultsWithBuilds := []interface{}{}
			for i := range results {
				// Break the builds into multiple assets if we use the tabular format
				if len(results[i].Builds) > 0 && format == config.FormatTabular {
					for _, build := range results[i].Builds {
						asset := corev2.Asset{
							ObjectMeta: results[i].ObjectMeta,
							URL:        build.URL,
							Sha512:     build.Sha512,
							Filters:    build.Filters,
							Headers:    build.Headers,
						}
						resultsWithBuilds = append(resultsWithBuilds, asset)
					}
				} else {
					resultsWithBuilds = append(resultsWithBuilds, results[i])
				}
				resources = append(resources, &results[i])
			}

			return helpers.PrintList(cmd, cli.Config.Format(), printToTable, resources, resultsWithBuilds, header)
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
				asset, ok := data.(corev2.Asset)
				if !ok {
					return cli.TypeError
				}
				return asset.Name
			},
		},
		{
			Title: "URL",
			CellTransformer: func(data interface{}) string {
				asset, ok := data.(corev2.Asset)
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
				asset, ok := data.(corev2.Asset)
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
