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
		&corev2.APIKey{},
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

	// synonyms provides user-friendly resource synonyms like checks, entities
	synonyms = map[string]corev2.Resource{}
)

func init() {
	for _, resource := range All {
		synonyms[resource.RBACName()] = resource
	}
}

type lifter interface {
	Lift() types.Resource
}

var resourceRE = regexp.MustCompile(`(\w+\/v\d+\.)?(\w+)`)

// ResolveResource resolves a named resource to an empty concrete type.
// The value is boxed within a types.Resource interface value.
func ResolveResource(resource string) (types.Resource, error) {
	if resource, ok := synonyms[resource]; ok {
		return resource, nil
	}
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

Dump resources to stdout or a file. Example:
$ sensuctl dump checks

The tool also supports naming types by their fully-qualified names:
$ sensuctl dump core/v2.CheckConfig,core/v2.Entity

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
	format := cli.Config.Format()
	if format != config.FormatWrappedJSON && format != config.FormatYAML {
		format = config.FormatYAML
	}
	_ = cmd.Flags().StringP("format", "", format, fmt.Sprintf(`format of data returned ("%s"|"%s")`, config.FormatWrappedJSON, config.FormatYAML))
	_ = cmd.Flags().StringP("file", "f", "", "file to dump resources to")
	_ = cmd.Flags().BoolP("types", "t", false, "list supported resource types")

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
		resource, err := ResolveResource(t)
		if err != nil {
			return nil, fmt.Errorf("invalid resource type: %s", t)
		}
		if lifter, ok := resource.(lifter); ok {
			resource = lifter.Lift()
		}
		actions = append(actions, resource)
	}
	return actions, nil
}

func printAllTypes(cli *cli.SensuCli, cmd *cobra.Command) error {
	var typeNames []string
	for _, resource := range All {
		wrapped := types.WrapResource(resource)
		typeNames = append(typeNames, fmt.Sprintf("%s.%s", wrapped.APIVersion, wrapped.Type))
	}
	switch getFormat(cli, cmd) {
	case config.FormatJSON, config.FormatWrappedJSON:
		return helpers.PrintJSON(typeNames, cmd.OutOrStdout())
	case config.FormatYAML:
		return helpers.PrintYAML(typeNames, cmd.OutOrStdout())
	default:
		for _, name := range typeNames {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), name); err != nil {
				return err
			}
		}
		return nil
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

func execute(cli *cli.SensuCli) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		printTypes, err := cmd.Flags().GetBool("types")
		if err != nil {
			return err
		}
		if printTypes {
			if len(args) > 0 {
				return errors.New("--types is mutually exclusive with positional args")
			}
			return printAllTypes(cli, cmd)
		}

		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("invalid argument(s) received")
		}
		format := getFormat(cli, cmd)
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

		for i, req := range requests {
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
				}, nil)
			if err != nil {
				// We want to ignore non-nil errors that are a result of
				// resources not existing, or features being licensed.
				err, ok := err.(client.APIError)
				if !ok {
					return fmt.Errorf("API error: %s", err)
				}
				switch actions.ErrCode(err.Code) {
				case actions.PaymentRequired, actions.NotFound, actions.PermissionDenied:
					continue
				}
				return fmt.Errorf("API error: %s", err)
			}

			val = reflect.Indirect(val)
			if val.Len() == 0 {
				continue
			}

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
				if i > 0 {
					_, _ = fmt.Fprintln(w, "---")
				}
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
