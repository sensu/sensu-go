package silenced

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/globals"
	"github.com/sensu/sensu-go/cli/elements/table"

	"github.com/spf13/cobra"
)

// ListCommand lists silenceds that were queried from the server
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list silenced entries",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}
			namespace := cli.Config.Namespace()
			flg := cmd.Flags()
			if ok, err := flg.GetBool(flags.AllNamespaces); err != nil {
				return err
			} else if ok {
				namespace = corev2.NamespaceTypeAll
			}
			// Fetch silenceds from the API
			sub, err := flg.GetString("subscription")
			if err != nil {
				return err
			}

			// We do not support both subscription and all-namespaces flags together
			if sub != "" && namespace == corev2.NamespaceTypeAll {
				return fmt.Errorf("the subscription and %s flags are mutually exclusive", flags.AllNamespaces)
			}

			check, err := flg.GetString("check")
			if err != nil {
				return err

			}

			opts, err := helpers.ListOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			var header http.Header
			results, err := cli.Client.ListSilenceds(namespace, sub, check, &opts, &header)
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

	flags := cmd.Flags()
	helpers.AddFormatFlag(flags)
	helpers.AddAllNamespace(flags)
	helpers.AddFieldSelectorFlag(cmd.Flags())
	helpers.AddLabelSelectorFlag(cmd.Flags())
	helpers.AddChunkSizeFlag(cmd.Flags())

	_ = flags.StringP("subscription", "s", "", "name of the silenced subscription")
	_ = flags.StringP("check", "c", "", "name of the silenced check")

	return cmd
}

func printToTable(results interface{}, writer io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				silenced, ok := data.(corev2.Silenced)
				if !ok {
					return cli.TypeError
				}
				return silenced.Name
			},
		},
		{
			Title: "Subscription",
			CellTransformer: func(data interface{}) string {
				silenced, ok := data.(corev2.Silenced)
				if !ok {
					return cli.TypeError
				}
				return silenced.Subscription
			},
		},
		{
			Title: "Check",
			CellTransformer: func(data interface{}) string {
				silenced, ok := data.(corev2.Silenced)
				if !ok {
					return cli.TypeError
				}
				return silenced.Check
			},
		},
		{
			Title: "Begin",
			CellTransformer: func(data interface{}) string {
				silenced, ok := data.(corev2.Silenced)
				if !ok {
					return cli.TypeError
				}
				return time.Unix(silenced.Begin, 0).Format(timeFormat)
			},
		},
		{
			Title: "Expiration",
			CellTransformer: func(data interface{}) string {
				silenced, ok := data.(corev2.Silenced)
				if !ok {
					return cli.TypeError
				}
				return expireAt(silenced.ExpireAt)
			},
		},
		{
			Title: "ExpireOnResolve",
			CellTransformer: func(data interface{}) string {
				silenced, ok := data.(corev2.Silenced)
				if !ok {
					return cli.TypeError
				}
				return globals.BooleanStyleP(silenced.ExpireOnResolve)
			},
		},
		{
			Title: "Creator",
			CellTransformer: func(data interface{}) string {
				silenced, ok := data.(corev2.Silenced)
				if !ok {
					return cli.TypeError
				}
				return silenced.Creator
			},
		},
		{
			Title: "Reason",
			CellTransformer: func(data interface{}) string {
				silenced, ok := data.(corev2.Silenced)
				if !ok {
					return cli.TypeError
				}
				return silenced.Reason
			},
		},
		{
			Title:       "Namespace",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				silenced, ok := data.(corev2.Silenced)
				if !ok {
					return cli.TypeError
				}
				return silenced.Namespace
			},
		},
	})

	table.Render(writer, results)
}
