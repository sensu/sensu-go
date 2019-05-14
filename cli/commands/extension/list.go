package extension

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"

	"github.com/spf13/cobra"
)

// ListCommand lists all extensions
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list extensions",
		RunE:  runList(cli.Config.Format(), cli.Client, cli.Config.Namespace(), cli.Config.Format()),
	}

	helpers.AddAllNamespace(cmd.Flags())
	helpers.AddFormatFlag(cmd.Flags())
	helpers.AddFieldSelectorFlag(cmd.Flags())
	helpers.AddLabelSelectorFlag(cmd.Flags())
	helpers.AddChunkSizeFlag(cmd.Flags())

	return cmd
}

func runList(config string, c client.APIClient, namespace, format string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			_ = cmd.Help()
			return errors.New("invalid arguments received")
		}
		if ok, _ := cmd.Flags().GetBool(flags.AllNamespaces); ok {
			namespace = types.NamespaceTypeAll
		}

		opts, err := helpers.ListOptionsFromFlags(cmd.Flags())
		if err != nil {
			return err
		}

		extensions, err := c.ListExtensions(namespace, &opts)
		if err != nil {
			return err
		}

		// Print the results based on the user preferences
		resources := []types.Resource{}
		for i := range extensions {
			resources = append(resources, &extensions[i])
		}
		return helpers.Print(cmd, config, printToTable, resources, extensions)
	}
}

func printToTable(results interface{}, writer io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				extension, ok := data.(types.Extension)
				if !ok {
					return cli.TypeError
				}
				return extension.Name
			},
		},
		{
			Title: "URL",
			CellTransformer: func(data interface{}) string {
				extension, ok := data.(types.Extension)
				if !ok {
					return cli.TypeError
				}
				u, err := url.Parse(extension.URL)
				if err != nil {
					return extension.URL
				}

				_, file := path.Split(u.EscapedPath())
				return fmt.Sprintf(
					"//%s/.../%s",
					u.Hostname(),
					file,
				)
			},
		},
	})

	table.Render(writer, results)
}
