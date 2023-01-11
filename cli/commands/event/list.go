package event

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/globals"
	"github.com/sensu/sensu-go/cli/elements/table"

	"github.com/spf13/cobra"
)

// ListCommand defines new list events command
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list events",
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

			// Fetch events from API
			var header http.Header
			results := []corev2.Event{}
			err = cli.Client.List(client.EventsPath(namespace), &results, &opts, &header)
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
			Title:       "Entity",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				event, ok := data.(corev2.Event)
				if !ok {
					return cli.TypeError
				}
				return event.Entity.Name
			},
		},
		{
			Title: "Check",
			CellTransformer: func(data interface{}) string {
				event, ok := data.(corev2.Event)
				if !ok {
					return cli.TypeError
				}
				return event.Check.Name
			},
		},
		{
			Title: "Output",
			CellTransformer: func(data interface{}) string {
				event, ok := data.(corev2.Event)
				if !ok {
					return cli.TypeError
				}
				return event.Check.Output
			},
		},
		{
			Title: "Status",
			CellTransformer: func(data interface{}) string {
				event, ok := data.(corev2.Event)
				if !ok {
					return cli.TypeError
				}
				return strconv.Itoa(int(event.Check.Status))
			},
		},
		{
			Title: "Silenced",
			CellTransformer: func(data interface{}) string {
				event, ok := data.(corev2.Event)
				if !ok {
					return cli.TypeError
				}
				return globals.BooleanStyleP(event.Check.IsSilenced)
			},
		},
		{
			Title: "Timestamp",
			CellTransformer: func(data interface{}) string {
				event, ok := data.(corev2.Event)
				if !ok {
					return cli.TypeError
				}
				time := time.Unix(event.Timestamp, 0)
				return time.String()
			},
		},
		{
			Title: "UUID",
			CellTransformer: func(data interface{}) string {
				event, ok := data.(corev2.Event)
				if !ok {
					return cli.TypeError
				}
				if id := event.GetUUID(); id == uuid.Nil {
					return ""
				} else {
					return id.String()
				}
			},
		},
	})

	table.Render(writer, results)
}
