package asset

import (
	"io"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// ShowCommand defines new asset info command
func ShowCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info NAME",
		Short:        "show detailed information on given asset",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")

			if len(args) != 1 {
				cmd.Help()
				return nil
			}

			// Fetch handlers from API
			assetName := args[0]
			r, err := cli.Client.FetchAsset(assetName)
			if err != nil {
				return err
			}

			if format == "json" {
				helpers.PrintJSON(r, cmd.OutOrStdout())
			} else {
				printAssetToList(r, cmd.OutOrStdout())
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags(), cli.Config)

	return cmd
}

func printAssetToList(r *types.Asset, writer io.Writer) {
	var metadata []string
	for k, v := range r.Metadata {
		metadata = append(metadata, k+"="+v)
	}

	cfg := &list.Config{
		Title: r.Name,
		Rows: []*list.Row{
			{
				Label: "Name",
				Value: r.Name,
			},
			{
				Label: "Organization",
				Value: r.Organization,
			},
			{
				Label: "URL",
				Value: r.URL,
			},
			{
				Label: "SHA-512 Checksum",
				Value: r.Sha512,
			},
			{
				Label: "Filters",
				Value: strings.Join(r.Filters, ", "),
			},
			{
				Label: "Metadata",
				Value: strings.Join(metadata, ", "),
			},
		},
	}

	list.Print(writer, cfg)
}
