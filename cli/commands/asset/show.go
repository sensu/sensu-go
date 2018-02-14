package asset

import (
	"errors"
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
		Use:          "info [NAME]",
		Short:        "show detailed information on given asset",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch handlers from API
			assetName := args[0]
			r, err := cli.Client.FetchAsset(assetName)
			if err != nil {
				return err
			}

			// Determine the format to use to output the data
			var format string
			if format = helpers.GetChangedStringValueFlag("format", cmd.Flags()); format == "" {
				format = cli.Config.Format()
			}

			if format == "json" {
				if err := helpers.PrintJSON(r, cmd.OutOrStdout()); err != nil {
					return err
				}
			} else {
				printAssetToList(r, cmd.OutOrStdout())
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

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
