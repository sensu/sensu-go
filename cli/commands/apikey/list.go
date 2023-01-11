package apikey

import (
	"errors"
	"io"
	"net/http"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/commands/timeutil"
	"github.com/sensu/sensu-go/cli/elements/table"

	"github.com/spf13/cobra"
)

// ListCommand adds a command that displays all apikeys.
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list api-keys",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			opts, err := helpers.ListOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			var header http.Header
			apikey := &corev2.APIKey{}
			results := []corev2.APIKey{}
			err = cli.Client.List(apikey.URIPath(), &results, &opts, &header)
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
				apikey, ok := data.(corev2.APIKey)
				if !ok {
					return cli.TypeError
				}
				return apikey.Name
			},
		},
		{
			Title: "Username",
			CellTransformer: func(data interface{}) string {
				apikey, ok := data.(corev2.APIKey)
				if !ok {
					return cli.TypeError
				}
				return apikey.Username
			},
		},
		{
			Title: "Created At",
			CellTransformer: func(data interface{}) string {
				apikey, ok := data.(corev2.APIKey)
				if !ok {
					return cli.TypeError
				}
				return timeutil.HumanTimestamp(apikey.CreatedAt)
			},
		},
	})

	table.Render(writer, results)
}
