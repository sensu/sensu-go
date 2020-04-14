package describetype

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/cli/resource"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

var description = `sensuctl describe-type

Describe a type by its fully qualified name:
$ sensuctl describe-type core/v2.CheckConfig

The command also supports describing types by their short names (for core/v2 types):
$ sensuctl describe-type checks,handlers

You can also use the 'all' qualifier to describe all available types:
$ sensuctl describe-type all
`

type apiResource struct {
	Name       string `json:"name"`
	ShortName  string `json:"short_name" yaml:"short_name"`
	APIVersion string `json:"api_version" yaml:"api_version"`
	Type       string `json:"type"`
	Namespaced bool   `json:"namespaced"`
}

// Command defines the describe-type command
func Command(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe-type [RESOURCE TYPE],[RESOURCE TYPE]...",
		Short: "Print details about the supported API resources types",
		Long:  description,
		RunE:  execute(cli),
	}

	format := cli.Config.Format()
	_ = cmd.Flags().StringP("format", "", format, fmt.Sprintf(`format of data returned ("%s"|"%s")`, config.FormatWrappedJSON, config.FormatYAML))

	return cmd
}

func execute(cli *cli.SensuCli) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var resources []apiResource

		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("invalid argument(s) received, expected a single one or a comma-separated list")
		}

		requested, err := resource.GetResourceRequests(args[0], resource.All)
		if err != nil {
			return err
		}

		for _, resource := range requested {
			wrapped := types.WrapResource(resource)

			// Short names are only supported for core/v2 resources
			shortName := ""
			if wrapped.APIVersion == "core/v2" {
				shortName = resource.RBACName()
			}

			r := apiResource{
				Name:       fmt.Sprintf("%s.%s", wrapped.APIVersion, wrapped.Type),
				ShortName:  shortName,
				APIVersion: wrapped.APIVersion,
				Type:       wrapped.Type,
				Namespaced: isNamespaced(resource),
			}
			resources = append(resources, r)
		}
		switch getFormat(cli, cmd) {
		case config.FormatJSON, config.FormatWrappedJSON:
			return helpers.PrintJSON(resources, cmd.OutOrStdout())
		case config.FormatYAML:
			return helpers.PrintYAML(resources, cmd.OutOrStdout())
		default:
			printToTable(resources, cmd.OutOrStdout())
			return nil
		}
	}
}

func getFormat(cli *cli.SensuCli, cmd *cobra.Command) string {
	// get the configured format or the flag override
	format := cli.Config.Format()
	if flag := helpers.GetChangedStringValueFlag("format", cmd.Flags()); flag != "" {
		format = flag
	}
	return format
}

// isNamespaced is a hack to determine whether a resource is global or
// namespaced, by relying on the SetNamespace method, which is a no-op for
// global resources, and inspecting the resulting namespace
func isNamespaced(r corev2.Resource) bool {
	r.SetNamespace("~sensu")
	return r.GetObjectMeta().Namespace == "~sensu"
}

func printToTable(results interface{}, writer io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "Fully Qualified Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				r, ok := data.(apiResource)
				if !ok {
					return cli.TypeError
				}
				return r.Name
			},
		},
		{
			Title: "Short Name",
			CellTransformer: func(data interface{}) string {
				r, ok := data.(apiResource)
				if !ok {
					return cli.TypeError
				}
				return r.ShortName
			},
		},
		{
			Title: "API Version",
			CellTransformer: func(data interface{}) string {
				r, ok := data.(apiResource)
				if !ok {
					return cli.TypeError
				}
				return r.APIVersion
			},
		},
		{
			Title: "Type",
			CellTransformer: func(data interface{}) string {
				r, ok := data.(apiResource)
				if !ok {
					return cli.TypeError
				}
				return r.Type
			},
		},
		{
			Title: "Namespaced",
			CellTransformer: func(data interface{}) string {
				r, ok := data.(apiResource)
				if !ok {
					return cli.TypeError
				}
				return strconv.FormatBool(r.Namespaced)
			},
		},
	})

	table.Render(writer, results)
}
