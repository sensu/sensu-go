package dump

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

var (
	// All is all the core resource types and associated sensuctl verbs (non-namespaced resources are intentionally ordered first).
	All = []types.Resource{
		&corev2.Namespace{},
		&corev2.ClusterRole{},
		&corev2.ClusterRoleBinding{},
		&corev2.User{},
		&corev2.TessenConfig{},
		&corev2.Asset{},
		&corev2.CheckConfig{},
		&corev2.Entity{},
		&corev2.Event{},
		&corev2.EventFilter{},
		&corev2.Handler{},
		&corev2.Hook{},
		&corev2.Mutator{},
		&corev2.Role{},
		&corev2.RoleBinding{},
		&corev2.Silenced{},
	}

	ChunkSize = 100
)

var resourceRE = regexp.MustCompile(`(\w+\/v\d+\.)?(\w+)`)

func resolveResource(resource string) (types.Resource, error) {
	matches := resourceRE.FindStringSubmatch(resource)
	if len(matches) != 3 {
		return nil, fmt.Errorf("bad resource qualifier: %s. hint: try something like core/v2.CheckConfig", resource)
	}
	apiVersion := strings.TrimSuffix(matches[1], ".")
	typeName := matches[2]
	if apiVersion == "" {
		apiVersion = "core/v2"
	}
	return types.ResolveType(apiVersion, typeName)
}

var description = `sensuctl dump

Dump resources to stdout or a file.

The tool accepts arguments in the form of resource types which are
comma delimited. For instance,

$ sensuctl dump core/v2.CheckConfig,core/v2.Entity

will dump all check configurations and entities.

You can also use the 'all' qualifier to dump all supported resources:

$ sensuctl dump all
`

// Command dumps generic Sensu resources to a file or STDOUT.
func Command(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "dump [RESOURCE TYPE],[RESOURCE TYPE]... [-f FILE]",
		Long: description,
		RunE: execute(cli),
	}

	helpers.AddAllNamespace(cmd.Flags())
	_ = cmd.Flags().StringP("format", "", cli.Config.Format(), fmt.Sprintf(`format of data returned ("%s"|"%s")`, config.FormatWrappedJSON, config.FormatYAML))
	_ = cmd.Flags().StringP("file", "f", "", "file to dump resources to")

	return cmd
}

func dedupTypes(arg string) []string {
	types := strings.Split(arg, ",")
	seen := make(map[string]struct{})
	result := make([]string, 0, len(types))
	for _, t := range types {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		result = append(result, t)
	}
	return result
}

func getResourceRequests(actionSpec string) ([]types.Resource, error) {
	// parse the comma separated resource types and match against the defined actions
	if actionSpec == "all" {
		return All, nil
	}
	var actions []types.Resource
	// deduplicate requested resources
	types := dedupTypes(actionSpec)

	// build resource requests for sensuctl
	for _, t := range types {
		resource, err := resolveResource(t)
		if err != nil {
			return nil, fmt.Errorf("invalid resource type: %s", t)
		}
		actions = append(actions, resource)
	}
	return actions, nil
}

func execute(cli *cli.SensuCli) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("invalid argument(s) received")
		}

		// get the configured format or the flag override
		format := cli.Config.Format()
		if flag := helpers.GetChangedStringValueFlag("format", cmd.Flags()); flag != "" {
			format = flag
		}
		switch format {
		case config.FormatYAML, config.FormatWrappedJSON:
		default:
			format = config.FormatYAML
		}

		// parse the comma separated resource types and match against the defined actions
		requests, err := getResourceRequests(args[0])
		if err != nil {
			return err
		}

		var w io.Writer = cmd.OutOrStdout()

		// if a file is requested, write data to that
		fp, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}
		if fp != "" {
			f, err := os.Create(fp)
			if err != nil {
				return err
			}
			defer f.Close()
			w = f
		}

		for _, req := range requests {
			// set the namespaces on the requests
			ok, err := cmd.Flags().GetBool(flags.AllNamespaces)
			if err != nil {
				return err
			}
			if ok {
				req.SetNamespace(corev2.NamespaceTypeAll)
			} else {
				req.SetNamespace(cli.Config.Namespace())
			}

			val := reflect.New(reflect.SliceOf(reflect.TypeOf(req)))
			err = cli.Client.List(
				req.URIPath(), val.Interface(), &client.ListOptions{
					ChunkSize: ChunkSize,
				})
			if err != nil {
				// We want to ignore non-nil errors that are a result of
				// resources not existing, or features being licensed.
				err, ok := err.(client.APIError)
				if !ok {
					return fmt.Errorf("API error: %s", err)
				}
				switch actions.ErrCode(err.Code) {
				case actions.PaymentRequired, actions.NotFound:
					continue
				}
				return fmt.Errorf("API error: %s", err)
			}

			val = reflect.Indirect(val)
			resources := make([]corev2.Resource, val.Len())
			for i := range resources {
				resources[i] = val.Index(i).Interface().(corev2.Resource)
			}

			switch format {
			case config.FormatJSON:
				err = helpers.PrintJSON(resources, w)
			case config.FormatWrappedJSON:
				err = helpers.PrintWrappedJSONList(resources, w)
			case config.FormatYAML:
				err = helpers.PrintYAML(resources, w)
			default:
				err = fmt.Errorf("invalid output format: %s", format)
			}

			if err != nil {
				return err
			}
		}

		return nil
	}
}
