package vet

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"reflect"

	"cuelang.org/go/cue/cuecontext"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/flags"
	shelpers "github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/resource"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/compat"
	"github.com/spf13/cobra"
)

var ChunkSize = 100
var description = `sensuctl vet

Vet resources Example:
$ sensuctl vet checks --spec "ttl: > 0"

The tool also supports naming types by their fully-qualified names:
$ sensuctl vet core/v2.CheckConfig --spec "ttl: > 0"

The tool uses cue (cuelang.org) to constrain types
`

func Command(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vet [RESOURCE TYPE] [-spec SPEC]",
		Short: "Vet resources",
		Long:  description,
		RunE:  execute(cli),
	}

	_ = cmd.Flags().StringP("spec", "s", "{}", "spec for resources to follow")
	shelpers.AddAllNamespace(cmd.Flags())
	return cmd
}

func execute(cli *cli.SensuCli) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		resourceSpec, err := cmd.Flags().GetString("spec")
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("invalid argument(s) received")
		}
		// parse the comma separated resource types and match against the defined actions
		requests, err := resource.GetResourceRequests(args[0], resource.All)
		if err != nil {
			return err
		}

		req := requests[0]
		ok, err := cmd.Flags().GetBool(flags.AllNamespaces)
		if err != nil {
			return err
		}
		if ok {
			req.SetNamespace(corev2.NamespaceTypeAll)
		} else {
			req.SetNamespace(cli.Config.Namespace())
		}

		var val reflect.Value
		if proxy, ok := req.(*corev3.V2ResourceProxy); ok {
			val = reflect.New(reflect.SliceOf(reflect.TypeOf(proxy.Resource)))
		} else {
			val = reflect.New(reflect.SliceOf(reflect.TypeOf(req)))
		}
		err = cli.Client.List(
			fmt.Sprintf("%s?types=%s", req.URIPath(), url.QueryEscape(types.WrapResource(req).Type)),
			val.Interface(), &client.ListOptions{
				ChunkSize: ChunkSize,
			}, nil)
		if err != nil {
			// We want to ignore non-nil errors that are a result of
			// resources not existing, or features being licensed.
			return fmt.Errorf("API error: %s", err)
		}
		val = reflect.Indirect(val)
		if val.Len() == 0 {
			return nil
		}

		resources := make([]corev2.Resource, val.Len())
		for i := range resources {
			resources[i] = compat.V2Resource(val.Index(i).Interface())
		}

		buf := bytes.Buffer{}
		shelpers.PrintJSON(resources, &buf)

		ctx := cuecontext.New()
		constraint := ctx.CompileString(fmt.Sprintf("[...#resource]\n#resource: {...} & %s", resourceSpec))
		if cerr := constraint.Err(); cerr != nil {
			return cerr
		}
		cv := ctx.CompileBytes(buf.Bytes())
		return constraint.Subsume(cv)

	}
}
