package asset

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/spf13/cobra"
)

// InfoCommand defines new asset info command
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:		"info [NAME]",
		Short:		"show detailed information on given asset",
		SilenceUsage:	true,
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
			flag := helpers.GetChangedStringValueViper("format", cmd.Flags())
			format := cli.Config.Format()
			return helpers.PrintFormatted(flag, format, r, cmd.OutOrStdout(), printToList)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printToList(v interface{}, writer io.Writer) error {
	r, ok := v.(*v2.Asset)
	if !ok {
		return fmt.Errorf("%t is not an Asset", v)
	}

	var cfg *list.Config

	if len(r.Builds) > 0 {
		buildRows := []*list.Row{}
		for i, build := range r.Builds {
			rows := []*list.Row{
				{
					Label:	"Build",
					Value:	strconv.Itoa(i),
				},
				{
					Label:	"URL",
					Value:	build.URL,
				},
				{
					Label:	"SHA-512 Checksum",
					Value:	build.Sha512,
				},
				{
					Label:	"Filters",
					Value:	strings.Join(build.Filters, ", "),
				},
			}
			buildRows = append(buildRows, rows...)
		}
		cfg = &list.Config{
			Title:	r.Name,
			Rows: []*list.Row{
				{
					Label:	"Name",
					Value:	r.Name,
				},
				{
					Label:	"Namespace",
					Value:	r.Namespace,
				},
				{
					Label:	"Builds",
					Value:	"",
				},
			},
		}
		cfg.Rows = append(cfg.Rows, buildRows...)
	} else {
		cfg = &list.Config{
			Title:	r.Name,
			Rows: []*list.Row{
				{
					Label:	"Name",
					Value:	r.Name,
				},
				{
					Label:	"Namespace",
					Value:	r.Namespace,
				},
				{
					Label:	"URL",
					Value:	r.URL,
				},
				{
					Label:	"SHA-512 Checksum",
					Value:	r.Sha512,
				},
				{
					Label:	"Filters",
					Value:	strings.Join(r.Filters, ", "),
				},
			},
		}
	}

	return list.Print(writer, cfg)
}
