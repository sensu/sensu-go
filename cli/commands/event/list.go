package event

import (
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/globals"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"
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
			org := cli.Config.Organization()
			if ok, _ := cmd.Flags().GetBool(flags.AllOrgs); ok {
				org = "*"
			}

			// Fetch events from API
			results, err := cli.Client.ListEvents(org)
			if err != nil {
				return err
			}

			// Print the results based on the user preferences
			return helpers.Print(cmd, cli.Config.Format(), printToTable, results)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	helpers.AddAllOrganization(cmd.Flags())

	return cmd
}

func printToTable(results interface{}, writer io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "Entity",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				event, _ := data.(types.Event)
				return event.Entity.ID
			},
		},
		{
			Title: "Check",
			CellTransformer: func(data interface{}) string {
				event, _ := data.(types.Event)
				return event.Check.Name
			},
		},
		{
			Title: "Output",
			CellTransformer: func(data interface{}) string {
				event, _ := data.(types.Event)
				return event.Check.Output
			},
		},
		{
			Title: "Status",
			CellTransformer: func(data interface{}) string {
				event, _ := data.(types.Event)
				return strconv.Itoa(int(event.Check.Status))
			},
		},
		{
			Title: "Silenced",
			CellTransformer: func(data interface{}) string {
				event, _ := data.(types.Event)
				return globals.BooleanStyleP(len(event.Silenced) > 0)
			},
		},
		{
			Title: "Timestamp",
			CellTransformer: func(data interface{}) string {
				event, _ := data.(types.Event)
				time := time.Unix(event.Timestamp, 0)
				return time.String()
			},
		},
	})

	table.Render(writer, results)
}
