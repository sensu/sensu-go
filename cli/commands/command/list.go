package command

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/cmdmanager"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/spf13/cobra"
)

// ListCommand adds command that allows a user to list installed command
// plugins.
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "lists installed sensuctl commands",
		RunE:  listCommandExecute(cli),
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func listCommandExecute(cli *cli.SensuCli) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// If args are provided print out usage
		if len(args) > 0 {
			_ = cmd.Help()
			return errors.New("invalid argument(s) received")
		}

		manager, err := cmdmanager.NewCommandManager(cli)
		if err != nil {
			return err
		}

		commandPlugins, err := manager.FetchCommandPlugins()
		if err != nil {
			return err
		}

		var header http.Header

		// Print the results based on the user preferences
		resources := []corev3.Resource{}
		var resultsWithBuilds []interface{}
		for i := range commandPlugins {
			if len(commandPlugins[i].Asset.Builds) > 0 {
				for _, build := range commandPlugins[i].Asset.Builds {
					asset := corev2.Asset{
						ObjectMeta: commandPlugins[i].Asset.ObjectMeta,
						URL:        build.URL,
						Sha512:     build.Sha512,
						Filters:    build.Filters,
						Headers:    build.Headers,
					}
					commandPlugin := &cmdmanager.CommandPlugin{
						Alias: commandPlugins[i].Alias,
						Asset: asset,
					}
					resultsWithBuilds = append(resultsWithBuilds, commandPlugin)
				}
			} else {
				resultsWithBuilds = append(resultsWithBuilds, commandPlugins[i])
			}
			resources = append(resources, commandPlugins[i])
		}

		return helpers.PrintList(cmd, cli.Config.Format(), printToTable, resources, resultsWithBuilds, header)
	}
}

func printToTable(results interface{}, writer io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "Alias",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				commandPlugin, ok := data.(*cmdmanager.CommandPlugin)
				if !ok {
					return cli.TypeError
				}
				return commandPlugin.Alias
			},
		},
		{
			Title: "URL",
			CellTransformer: func(data interface{}) string {
				commandPlugin, ok := data.(*cmdmanager.CommandPlugin)
				if !ok {
					return cli.TypeError
				}
				u, err := url.Parse(commandPlugin.Asset.URL)
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
				commandPlugin, ok := data.(*cmdmanager.CommandPlugin)
				if !ok {
					return cli.TypeError
				}
				if len(commandPlugin.Asset.Sha512) >= 128 {
					return string(commandPlugin.Asset.Sha512[0:7])
				}
				return "invalid"
			},
		},
	})

	table.Render(writer, results)
}
