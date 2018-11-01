package asset

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

// InfoCommand defines new asset info command
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
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
			flag := helpers.GetChangedStringValueFlag("format", cmd.Flags())
			format := cli.Config.Format()
			return helpers.PrintFormatted(flag, format, r, cmd.OutOrStdout(), printToList)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printToList(v interface{}, writer io.Writer) error {
	r, ok := v.(*types.Asset)
	if !ok {
		return fmt.Errorf("%t is not an Asset", v)
	}
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
				Label: "Namespace",
				Value: r.Namespace,
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

	return list.Print(writer, cfg)
}
